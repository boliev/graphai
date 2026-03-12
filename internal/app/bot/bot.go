package bot

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/boliev/graphai/internal/domain/ai"
	"github.com/boliev/graphai/internal/domain/bot"
	"github.com/boliev/graphai/internal/domain/user"
	"github.com/boliev/graphai/internal/infra/pg/repository"
	"github.com/boliev/graphai/internal/pkg/config"
	"github.com/boliev/graphai/internal/pkg/gemini"
	"github.com/boliev/graphai/internal/pkg/tg"
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
	data := make(map[string]*bot.Messages)

	tgBotApi, err := b.createTgBotApi(cfg)
	if err != nil {
		panic(err)
	}

	tgClient := tg.NewClient(tgBotApi)

	pool, err := pgxpool.New(ctx, cfg.PGConnect)
	if err != nil {
		panic(err)
	}

	userRepo := repository.NewUser(pool)
	userService := user.NewService(userRepo)

	sender := bot.NewSender(tgClient)
	commander := bot.NewCommander(sender)
	tgProcessor := bot.NewProcessor(data, tgBotApi, sender, commander, userService)

	aiClient, err := gemini.NewGemini(ctx, cfg.GeminiToken)
	if err != nil {
		panic(err)
	}
	aiProcessor := ai.NewProcessor(data, aiClient, sender)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err := tgProcessor.Run()
		if err != nil {
			wg.Done()
			log.Fatal(err)
		}
	}()

	wg.Add(1)
	go func() {
		err := aiProcessor.Run()
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
