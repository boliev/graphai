package config

import (
	"os"
	"strconv"
)

type Cfg struct {
	BotToken      string
	VKGroupToken  string
	GeminiToken   string
	PGConnect     string
	CommunityLink string
	PaymentAppID  int64
	VKApiPort     string
	VkSecureKey   string
}

func New() (*Cfg, error) {
	paymentAppId, err := strconv.ParseInt(os.Getenv("PAYMENT_APP_ID"), 10, 64)
	if err != nil {
		return nil, err
	}

	return &Cfg{
		BotToken:      os.Getenv("BOT_TOKEN"),
		VKGroupToken:  os.Getenv("VK_GROUP_TOKEN"),
		GeminiToken:   os.Getenv("GEMINI_TOKEN"),
		PGConnect:     os.Getenv("PG_CONNECT"),
		CommunityLink: os.Getenv("COMMUNITY_LINK"),
		PaymentAppID:  paymentAppId,
		VKApiPort:     os.Getenv("VK_API_PORT"),
		VkSecureKey:   os.Getenv("VK_SECURE_KEY"),
	}, nil
}
