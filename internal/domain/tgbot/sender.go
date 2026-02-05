package tgbot

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Sender struct {
}

func (s *Sender) Send(ctx context.Context, msg *tgbotapi.Message, photo []byte) error {

	return nil
}
