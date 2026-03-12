package config

import "os"

type Cfg struct {
	BotToken    string
	GeminiToken string
	PGConnect   string
}

func New() *Cfg {
	return &Cfg{
		BotToken:    os.Getenv("BOT_TOKEN"),
		GeminiToken: os.Getenv("GEMINI_TOKEN"),
		PGConnect:   os.Getenv("PG_CONNECT"),
	}
}
