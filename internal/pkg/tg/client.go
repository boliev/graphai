package tg

import (
	"fmt"

	"github.com/boliev/graphai/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Client struct {
	bot *tgbotapi.BotAPI
}

func NewClient(bot *tgbotapi.BotAPI) *Client {
	return &Client{
		bot: bot,
	}
}

func (c *Client) Send(msg domain.Message) error {
	message := tgbotapi.NewMessage(msg.ChatId, msg.Text)

	buttons := make([]tgbotapi.InlineKeyboardButton, 0, len(msg.Buttons))
	for _, button := range msg.Buttons {
		if button.Type == domain.MessageButtonTypeData {
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(button.Text, button.Data))...)
		}
		if button.Type == domain.MessageButtonTypeUrl {
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonURL(button.Text, button.Data))...)
		}
	}
	message.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttons)

	_, err := c.bot.Send(message)
	return err
}

func (c *Client) SendImage(msg domain.ImageMessage) error {
	m := tgbotapi.NewPhoto(msg.ChatId, tgbotapi.FileBytes{
		Name:  fmt.Sprintf("response.%s", msg.Ext),
		Bytes: msg.Image,
	})

	m.ReplyToMessageID = msg.ReplyToMessageId
	if _, err := c.bot.Send(m); err != nil {
		return err
	}

	return nil
}
