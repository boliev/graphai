package bot

import "fmt"

type Bot struct {
}

func New() *Bot {
	return &Bot{}
}

func (b *Bot) Start() error {
	fmt.Println("Bot is running")

	return nil
}
