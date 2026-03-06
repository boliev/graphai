package tgbot

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/boliev/graphai/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type AI interface {
	Send(ctx context.Context, description string, files []string) (*domain.AIResponse, error)
}

type Processor struct {
	data map[string]*Messages
	bot  *tgbotapi.BotAPI
	ai   AI
}

func NewProcessor(data map[string]*Messages, bot *tgbotapi.BotAPI, ai AI) *Processor {
	return &Processor{
		data: data,
		bot:  bot,
		ai:   ai,
	}
}

func (s *Processor) Run() error {
	for {
		for mediaGroupId, msg := range s.data {
			if time.Since(msg.Dt) > 2*time.Second {
				ctx := context.Background()
				resp, err := s.ai.Send(ctx, msg.Caption, msg.Paths)
				if err != nil {
					log.Printf("failed to send message: %v", err)
				}

				m := tgbotapi.NewPhoto(msg.ChatID, tgbotapi.FileBytes{
					Name:  fmt.Sprintf("response.%s", resp.Ext),
					Bytes: resp.Photo,
				})

				//m.ReplyToMessageID = msg.ReplyId
				if _, err := s.bot.Send(m); err != nil {
					log.Println("send text:", err)
				}
				delete(s.data, mediaGroupId)
				continue
			}
		}
	}
}
