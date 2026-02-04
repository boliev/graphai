package config

import "os"

type Cfg struct {
	BotToken string
}

func New() *Cfg {
	return &Cfg{
		BotToken: os.Getenv("BOT_TOKEN"),
	}
}
