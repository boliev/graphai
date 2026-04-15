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

const GEMINI_RESENDTRIES = 3

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
	lp.MessageNew(func(_ context.Context, obj events.MessageNewObject) {
		ctx := context.Background()
		msg := obj.Message
		p.logger.Info("got new message")
		user, err := p.userService.Upsert(ctx, p.userFromMessage(msg))
		if err != nil {
			p.logger.Warn("userService.Upsert() failed", "error", err)
			return
		}
		p.logger.Info("loaded user", "user", user.ID)

		if msg.Payload != "" {
			err := p.command(msg)
			if err != nil {
				p.logger.Warn("command failed", "error", err)
			}
			return
		}

		if !user.HasBalance() {
			p.sender.NotEnoughMoney(user.PeerID)
			p.logger.Info("not enough balance", "user", user.ID, "peerID", msg.PeerID)
			return
		}
		p.logger.Info("balance check done", "user", user.ID, "peerID", msg.PeerID)

		fullMsg, err := p.sender.loadFullMessage(msg)
		if err != nil {
			p.logger.Warn("loadFullMessage failed", "error", err, "peerID", msg.PeerID)
			fullMsg = msg // fallback на то, что пришло в событии
		}
		p.logger.Info("full message loaded", "user", user.ID, "peerID", msg.PeerID)

		photoURLs := p.sender.extractPhotoURLs(fullMsg.Attachments)
		if len(photoURLs) == 0 {
			err = p.sender.sendKB(msg.PeerID)
			if err != nil {
				p.logger.Error("sender.sendKB failed", "error", err)
			}
			return
		}
		p.logger.Info("user photos extracted", "user", user.ID, "photosCount", len(photoURLs), "peerID", msg.PeerID)

		resp, err := p.sendWithRetries(ctx, msg.Text, photoURLs)
		if err != nil {
			p.logger.Error("cannot send prompt after %d tries", "tries", GEMINI_RESENDTRIES, "error", err, "prompt", msg.Text, "user", user.ID, "userVkID", user.UserVKID, "peerID", msg.PeerID)
		}
		p.logger.Info("Got response from Gemini", "user", user.ID, "peerID", msg.PeerID)

		resultPhoto, err := p.sender.uploadMessagesPhoto(int64(msg.PeerID), resp.Photo)
		if err != nil {
			p.logger.Error("upload messages photo failed", "error", err, "user", user.ID, "userVkID", user.UserVKID, "peerID", msg.PeerID)
			return
		}
		p.logger.Info("uploaded result photo to VK", "user", user.ID, "peerID", msg.PeerID)

		err = p.sender.send(msg.PeerID, msg.ID, resultPhoto[0].OwnerID, resultPhoto[0].ID)
		if err != nil {
			p.logger.Error("failed to send vk messages", "error", err, "user", user.ID, "userVkID", user.UserVKID)
			return
		}
		p.logger.Info("Sent response to the user", "user", user.ID, "peerID", msg.PeerID)

		if user.HasBalance() {
			err = p.userService.ReduceCredits(ctx, user)
			if err != nil {
				p.logger.Error("failed to ReduceCredits() failed", "error", err, "user", user.ID, "userVkID", user.UserVKID)
				return
			}
			p.logger.Info("reduced user balance", "user", user.ID, "peerID", msg.PeerID)

			pr := &prompt.Prompt{
				UserID: user.ID,
				Prompt: msg.Text,
			}
			err := p.txService.Create(ctx, pr)
			if err != nil {
				p.logger.Error("failed to save prompt", "error", err, "user", user.ID, "userVkID", user.UserVKID, "prompt", msg.Text)
				return
			}
			p.logger.Info("save prompt", "user", user.ID, "peerID", msg.PeerID)
		}
	})

	p.logger.Info("VK long poll started")
	if err := lp.Run(); err != nil {
		p.logger.Error("VK long poll stopped", "error", err)
		return err
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
		return fmt.Errorf("invalid payload: %v, raw=%s", err, msg.Payload)
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
	for i := 1; i <= GEMINI_RESENDTRIES; i++ {
		resp, err := p.aiClient.Send(ctx, prompt, photoURLs)
		if err == nil {
			return resp, nil
		}
		p.logger.Warn("gemini aiClient.Send() failed", "error", err, "try", i, "prompt", prompt)
	}
	return nil, err
}
