package vk

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/SevereCloud/vksdk/v2/events"
	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/boliev/graphai/internal/domain"
	"github.com/boliev/graphai/internal/domain/prompt"
	"github.com/boliev/graphai/internal/domain/user"
)

const geminiResentTries = 3

type ai interface {
	Send(ctx context.Context, description string, files []string) (*domain.AIResponse, error)
}

type Processor struct {
	token       string
	sender      *Sender
	aiClient    ai
	userService *user.Service
	txService   *prompt.Service
	logger      *slog.Logger
}

func NewProcessor(token string, sender *Sender, ai ai, userService *user.Service, proptsService *prompt.Service, logger *slog.Logger) *Processor {
	return &Processor{
		token:       token,
		sender:      sender,
		aiClient:    ai,
		userService: userService,
		txService:   proptsService,
		logger:      logger,
	}
}

func (p *Processor) Run() error {
	// Создаем Long Poll клиент для сообщества.
	lp, err := p.sender.getLP()
	if err != nil {
		p.logger.Warn("longpoll.NewLongPoll failed", "error", err)
	}

	// Обрабатываем новые сообщения.
	lp.MessageNew(func(ctx context.Context, obj events.MessageNewObject) {
		p.logger.Info("got new message")

		msg := obj.Message

		usr, err := p.userService.Upsert(ctx, p.userFromMessage(msg))
		if err != nil {
			p.logger.Warn("userService.Upsert() failed", "error", err)
			return
		}

		p.logger.Info("loaded user", "user", usr.ID)

		err = p.handleMessage(ctx, msg, usr)
		if err != nil {
			p.logger.Error(err.Error(), "error", err, "user", usr.ID, "peerID", usr.PeerID)
		}
	})

	p.logger.Info("VK long poll started")

	if err := lp.Run(); err != nil {
		p.logger.Error("VK long poll stopped", "error", err)
		return err
	}

	return nil
}

func (p *Processor) handleMessage(ctx context.Context, msg object.MessagesMessage, usr *user.User) error {
	if msg.Payload != "" {
		err := p.command(msg)
		if err != nil {
			return fmt.Errorf("command failed: %w", err)
		}

		return nil
	}

	if !usr.HasBalance() {
		err := p.sender.NotEnoughMoney(usr.PeerID)
		if err != nil {
			return fmt.Errorf("sender.NotEnoughMoney failed: %w", err)
		}

		p.logDebug("not enough balance", usr)

		return nil
	}

	p.logDebug("balance check done", usr)

	fullMsg, err := p.sender.loadFullMessage(ctx, msg)
	if err != nil {
		p.logger.Warn("loadFullMessage failed", "error", err, "peerID", msg.PeerID, "user", usr.ID)
		fullMsg = msg // fallback на то, что пришло в событии
	}

	p.logDebug("full message loaded", usr)

	photoURLs := p.sender.extractPhotoURLs(fullMsg.Attachments)
	if len(photoURLs) == 0 {
		err = p.sender.sendKB(msg.PeerID)
		if err != nil {
			return fmt.Errorf("sender.sendKB failed: %w", err)
		}

		return nil
	}

	p.logDebug("user photos extracted", usr)

	resp, err := p.sendWithRetries(ctx, msg.Text, photoURLs)
	if err != nil || resp == nil {
		p.logger.Error("cannot send prompt",
			"tries", geminiResentTries,
			"error", err, "prompt",
			msg.Text, "user",
			usr.ID, "userVkID",
			usr.UserVKID,
			"peerID", msg.PeerID,
		)

		return fmt.Errorf("cannot send prompt: %w", err)
	}

	p.logDebug("Got response from Gemini", usr)

	resultPhoto, err := p.sender.uploadMessagesPhoto(int64(msg.PeerID), resp.Photo)
	if err != nil {
		return fmt.Errorf("cannot upload messages photo: %w", err)
	}

	p.logDebug("uploaded result photo to VK", usr)

	err = p.sender.send(msg.PeerID, msg.ID, resultPhoto[0].OwnerID, resultPhoto[0].ID)
	if err != nil {
		return fmt.Errorf("failed to send vk messages: %w", err)
	}

	p.logDebug("Sent response to the user", usr)

	if usr.HasBalance() {
		err = p.userService.ReduceCredits(ctx, usr)
		if err != nil {
			return fmt.Errorf("failed to reduce credits: %w", err)
		}

		p.logDebug("reduced user balance", usr)

		pr := &prompt.Prompt{
			UserID: usr.ID,
			Prompt: msg.Text,
		}

		err := p.txService.Create(ctx, pr)
		if err != nil {
			return fmt.Errorf("failed to save prompt: %w", err)
		}

		p.logDebug("save prompt", usr)
	}

	return nil
}

func (p *Processor) command(msg object.MessagesMessage) error {
	type ButtonPayload struct {
		Cmd string `json:"cmd"`
	}

	var c ButtonPayload

	err := json.Unmarshal([]byte(msg.Payload), &c)
	if err != nil {
		return fmt.Errorf("invalid payload: %w, raw=%s", err, msg.Payload)
	}

	switch c.Cmd {
	case "prices":
		err = p.sender.prices(int64(msg.PeerID))
	default:
		err = p.sender.help(int64(msg.PeerID))
	}

	if err != nil {
		return err
	}

	return nil
}

func (p *Processor) userFromMessage(msg object.MessagesMessage) *user.User {
	return &user.User{
		UserVKID: int64(msg.FromID),
		PeerID:   int64(msg.PeerID),
	}
}

func (p *Processor) sendWithRetries(ctx context.Context, prompt string, photoURLs []string) (*domain.AIResponse, error) {
	var err error
	for i := 1; i <= geminiResentTries; i++ {
		resp, err := p.aiClient.Send(ctx, prompt, photoURLs)
		if err == nil {
			return resp, nil
		}

		p.logger.Warn("gemini aiClient.Send() failed", "error", err, "try", i, "prompt", prompt)
	}

	return nil, err
}

func (p *Processor) logDebug(msg string, usr *user.User) {
	p.logger.Info(msg, "user", usr.ID, "peerID", usr.PeerID)
}
