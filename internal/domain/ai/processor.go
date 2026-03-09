package ai

import (
	"context"
	"log"
	"time"

	"github.com/boliev/graphai/internal/domain"
	"github.com/boliev/graphai/internal/domain/bot"
)

type AI interface {
	Send(ctx context.Context, description string, files []string) (*domain.AIResponse, error)
}

type Processor struct {
	data   map[string]*bot.Messages
	ai     AI
	sender *bot.Sender
}

func NewProcessor(data map[string]*bot.Messages, ai AI, sender *bot.Sender) *Processor {
	return &Processor{
		data:   data,
		ai:     ai,
		sender: sender,
	}
}

func (s *Processor) Run() error {
	for {
		for mediaGroupId, msg := range s.data {
			if time.Since(msg.Dt) > 2*time.Second {
				ctx := context.Background()
				resp, err := s.ai.Send(ctx, msg.Caption, msg.Paths)
				if err != nil {
					log.Printf("failed to send message to AI: %v", err)
					err = s.sender.SendError(msg.ChatID)
					if err != nil {
						log.Printf("failed to send error: %v", err)
					}
					delete(s.data, mediaGroupId)
					continue
				}

				err = s.sender.SendImage(domain.ImageMessage{
					ChatId:           msg.ChatID,
					ReplyToMessageId: msg.ReplyId,
					Ext:              resp.Ext,
					Image:            resp.Photo,
				})
				if err != nil {
					log.Println("send text:", err)

				}
				delete(s.data, mediaGroupId)
				continue
			}
		}
	}
}
