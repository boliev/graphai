package bot

import (
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Messages struct {
	Caption string
	ChatID  int64
	ReplyId int
	Paths   []string
	Dt      time.Time
}

type Collector struct {
	data      map[string]*Messages
	bot       *tgbotapi.BotAPI
	sender    *Sender
	commander *Commander
}

func NewProcessor(data map[string]*Messages, bot *tgbotapi.BotAPI, sender *Sender, commander *Commander) *Collector {
	return &Collector{
		data:      data,
		bot:       bot,
		sender:    sender,
		commander: commander,
	}
}

func (c *Collector) Run() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := c.bot.GetUpdatesChan(u)

	for update := range updates {
		m := update.Message
		if m != nil {
			if m.IsCommand() {
				err := c.commander.ExecuteCommand(m)
				if err != nil {
					log.Println(err)
				}
				continue
			}

			if len(m.Photo) > 0 {
				mgID := m.MediaGroupID
				if _, ok := c.data[mgID]; !ok {
					c.data[m.MediaGroupID] = &Messages{
						Caption: "",
						ChatID:  m.Chat.ID,
						ReplyId: m.MessageID,
						Paths:   []string{},
						Dt:      time.Now(),
					}
				}

				// обычно последний элемент — самое большое фото
				photo := m.Photo[len(m.Photo)-1]
				p := tgbotapi.NewPhoto(m.Chat.ID, tgbotapi.FileID(photo.FileID))

				file, err := c.bot.GetFile(tgbotapi.FileConfig{FileID: photo.FileID})
				if err != nil {
					return err
				}
				url := file.Link(c.bot.Token)
				c.data[mgID].Paths = append(c.data[mgID].Paths, url)

				if m.Caption != "" {
					c.data[mgID].Caption = m.Caption
				}
				p.Caption = m.Caption // подпись к фото (если есть)
				p.ReplyToMessageID = m.MessageID
				continue
			} else {
				if err := c.sender.HelpPleaseSendPhoto(m.Chat.ID); err != nil {
					log.Println("HelpPleaseSendPhoto error:", err)
				}
				continue
			}
		}

		if update.CallbackQuery != nil {
			cb := update.CallbackQuery
			_, err := c.bot.Request(tgbotapi.NewCallback(cb.ID, ""))
			if err != nil {
				log.Println(err)
			}

			err = c.commander.ExecuteCallback(cb)
			if err != nil {
				log.Println(err)
			}

			continue
		}
	}
	return nil
}
