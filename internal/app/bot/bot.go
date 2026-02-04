package bot

import (
	"fmt"
	"log"
	"sync"

	"github.com/boliev/graphai/internal/domain/tgbot"
	"github.com/boliev/graphai/internal/pkg/config"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
}

func New() *Bot {
	return &Bot{}
}

func (b *Bot) Start() error {
	fmt.Println("Bot is running")

	cfg := config.New()
	data := make(map[string]*tgbot.Messages)

	bot, err := b.createBot(cfg)
	if err != nil {
		panic(err)
	}
	collector := tgbot.NewCollector(data, bot)
	processor := tgbot.NewProcessor(data, bot)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err := collector.Run()
		if err != nil {
			wg.Done()
			log.Fatal(err)
		}
	}()

	wg.Add(1)
	go func() {
		err := processor.Run()
		if err != nil {
			wg.Done()
			log.Fatal(err)
		}
	}()
	wg.Wait()

	//MG := make(map[string]mediaGroup, 0)
	//

	return nil
}

func (b *Bot) createBot(cfg *config.Cfg) (*tgbotapi.BotAPI, error) {

	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, err
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	bot.Debug = true

	return bot, nil
}
