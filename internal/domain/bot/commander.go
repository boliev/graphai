package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Commander struct {
	sender *Sender
}

func NewCommander(sender *Sender) *Commander {
	return &Commander{
		sender: sender,
	}
}

func (c *Commander) ExecuteCommand(msg *tgbotapi.Message) error {
	if msg.Command() == "start" {
		err := c.sender.Greet(msg.Chat.ID)
		if err != nil {
			return err
		}
	}

	if msg.Command() == "make_image" {
		err := c.sender.HelpMakeImage(msg.Chat.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Commander) ExecuteCallback(cq *tgbotapi.CallbackQuery) error {
	if cq.Data == "make_image" {
		err := c.sender.HelpMakeImage(cq.Message.Chat.ID)
		if err != nil {
			return err
		}
	}

	return nil
}
