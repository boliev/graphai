package config

import "os"

type Cfg struct {
	BotToken      string
	VKGroupToken  string
	GeminiToken   string
	PGConnect     string
	CommunityLink string
}

func New() *Cfg {
	return &Cfg{
		BotToken:      os.Getenv("BOT_TOKEN"),
		VKGroupToken:  os.Getenv("VK_GROUP_TOKEN"),
		GeminiToken:   os.Getenv("GEMINI_TOKEN"),
		PGConnect:     os.Getenv("PG_CONNECT"),
		CommunityLink: os.Getenv("COMMUNITY_LINK"),
	}
}
