package tgbot

import (
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Processor struct {
	data map[string]*Messages
	bot  *tgbotapi.BotAPI
}

func NewProcessor(data map[string]*Messages, bot *tgbotapi.BotAPI) *Processor {
	return &Processor{
		data: data,
		bot:  bot,
	}
}

func (s *Processor) Run() error {
	for {
		for mediaGroupId, msg := range s.data {
			if time.Since(msg.Dt) > 2*time.Second {
				m := tgbotapi.NewMessage(msg.ChatID, fmt.Sprintf("Вы отправили % d фото с описанием `%s`", len(msg.Files), msg.Caption))
				m.ReplyToMessageID = msg.ReplyId
				if _, err := s.bot.Send(m); err != nil {
					log.Println("send text:", err)
				}
				delete(s.data, mediaGroupId)
				continue
			}
		}
	}

	return nil
}
