package bot

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/boliev/graphai/internal/domain/user"
	"github.com/boliev/graphai/internal/domain/vk"
	"github.com/boliev/graphai/internal/infra/pg/repository"
	"github.com/boliev/graphai/internal/pkg/config"
	"github.com/boliev/graphai/internal/pkg/gemini"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Bot struct {
}

func New() *Bot {
	return &Bot{}
}

func (b *Bot) Start() error {
	fmt.Println("Bot is running")

	ctx := context.Background()
	cfg := config.New()

	aiClient, err := gemini.NewGemini(ctx, cfg.GeminiToken)
	if err != nil {
		panic(err)
	}

	vkApi := api.NewVK(cfg.VKGroupToken)
	vkSender := vk.NewSender(vkApi, cfg)

	pool, err := pgxpool.New(ctx, cfg.PGConnect)
	if err != nil {
		panic(err)
	}

	userRepo := repository.NewUserRepo(pool)
	userService := user.NewService(userRepo)

	vkProcessor := vk.NewProcessor(cfg.VKGroupToken, vkSender, aiClient, userService)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		err := vkProcessor.Run()
		if err != nil {
			wg.Done()
			log.Fatal(err)
		}
	}()

	wg.Wait()

	return nil
}

func (b *Bot) createTgBotApi(cfg *config.Cfg) (*tgbotapi.BotAPI, error) {

	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, err
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	bot.Debug = true

	return bot, nil
}
