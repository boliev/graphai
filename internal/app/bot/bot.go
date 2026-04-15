package bot

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/boliev/graphai/internal/domain/prompt"
	"github.com/boliev/graphai/internal/domain/user"
	"github.com/boliev/graphai/internal/domain/vk"
	"github.com/boliev/graphai/internal/infra/pg/repository"
	"github.com/boliev/graphai/internal/pkg/config"
	"github.com/boliev/graphai/internal/pkg/gemini"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Bot struct {
}

func New() *Bot {
	return &Bot{}
}

func (b *Bot) Start() error {
	ctx := context.Background()
	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})).With("project", "graphai", "service", "vkBot")

	logger.Info("Bot is running")

	aiClient, err := gemini.NewGemini(ctx, cfg.GeminiToken)
	if err != nil {
		logger.Error(err.Error())
		panic(err)
	}

	vkApi := api.NewVK(cfg.VKGroupToken)
	vkSender, err := vk.NewSender(vkApi, cfg, logger)
	if err != nil {
		logger.Error(err.Error())
		panic(err)
	}

	pool, err := pgxpool.New(ctx, cfg.PGConnect)
	if err != nil {
		logger.Error(err.Error())
		panic(err)
	}

	userRepo := repository.NewUserRepo(pool)
	userService := user.NewService(userRepo)

	txRepo := repository.NewPromptsRepo(pool)
	promptsService := prompt.NewService(txRepo)

	vkProcessor := vk.NewProcessor(cfg.VKGroupToken, vkSender, aiClient, userService, promptsService, logger)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		err := vkProcessor.Run()
		if err != nil {
			wg.Done()
			logger.Error("vkProcessor failed", "error", err.Error())
		}
	}()

	wg.Wait()

	return nil
}
