package vk

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/SevereCloud/vksdk/v2/events"
	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/boliev/graphai/internal/domain"
	"github.com/boliev/graphai/internal/domain/prompt"
	"github.com/boliev/graphai/internal/domain/user"
)

type ai interface {
	Send(ctx context.Context, description string, files []string) (*domain.AIResponse, error)
}

type Processor struct {
	token       string
	sender      *Sender
	aiClient    ai
	userService *user.Service
	txService   *prompt.Service
}

func NewProcessor(token string, sender *Sender, ai ai, userService *user.Service, proptsService *prompt.Service) *Processor {
	return &Processor{
		token:       token,
		sender:      sender,
		aiClient:    ai,
		userService: userService,
		txService:   proptsService,
	}
}

func (p Processor) Run() error {
	// Создаем Long Poll клиент для сообщества.
	lp, err := p.sender.getLP()
	if err != nil {
		log.Fatalf("longpoll.NewLongPoll failed: %v", err)
	}

	// Обрабатываем новые сообщения.
	lp.MessageNew(func(_ context.Context, obj events.MessageNewObject) {
		ctx := context.Background()
		msg := obj.Message
		user, err := p.userService.Upsert(ctx, p.userFromMessage(msg))
		if err != nil {
			log.Fatalf("userService.Upsert() failed: %v", err)
			return
		}

		if msg.Payload != "" {
			err := p.command(msg)
			if err != nil {
				log.Printf("command failed: %v", err)
			}
			return
		}

		if !user.HasBalance() {
			p.sender.NotEnoughMoney(user.PeerID)
			return
		}

		fullMsg, err := p.sender.loadFullMessage(msg)
		if err != nil {
			log.Printf("load full message failed: %v", err)
			fullMsg = msg // fallback на то, что пришло в событии
		}

		photoURLs := p.sender.extractPhotoURLs(fullMsg.Attachments)
		log.Printf("full message attachments=%d photos=%d", len(fullMsg.Attachments), len(photoURLs))
		if len(photoURLs) == 0 {
			err = p.sender.sendKB(msg.PeerID)
			if err != nil {
				log.Printf("sender.sendKB failed: %v", err)
			}
			return
		}

		resp, err := p.aiClient.Send(ctx, msg.Text, photoURLs)
		if err != nil {
			log.Printf("ai.Send() failed: %v", err)
			return
		}

		resultPhoto, err := p.sender.uploadMessagesPhoto(int64(msg.PeerID), resp.Photo)
		if err != nil {
			log.Printf("upload messages photo failed: %v", err)
			return
		}

		err = p.sender.send(msg.PeerID, msg.ID, resultPhoto[0].OwnerID, resultPhoto[0].ID)
		if err != nil {
			log.Printf("messages.send failed: %v", err)
			return
		}

		if user.HasBalance() {
			err = p.userService.ReduceCredits(ctx, user)
			if err != nil {
				log.Printf("userService.ReduceCredits() failed: %v", err)
				return
			}

			pr := &prompt.Prompt{
				UserID: user.ID,
				Prompt: msg.Text,
			}
			err := p.txService.Create(ctx, pr)
			if err != nil {
				log.Printf("txService.Create() failed: %v", err)
				return
			}
		}
	})

	log.Println("VK long poll started")
	if err := lp.Run(); err != nil {
		log.Fatalf("long poll stopped: %v", err)
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
