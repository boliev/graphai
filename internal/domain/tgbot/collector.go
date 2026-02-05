package tgbot

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
	data map[string]*Messages
	bot  *tgbotapi.BotAPI
}

func NewCollector(data map[string]*Messages, bot *tgbotapi.BotAPI) *Collector {
	return &Collector{
		data: data,
		bot:  bot,
	}
}

func (c *Collector) Run() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := c.bot.GetUpdatesChan(u)

	for update := range updates {
		m := update.Message
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

			//if _, err := bot.Send(p); err != nil {
			//	log.Println("send photo:", err)
			//}
			continue
		} else {
			msg := tgbotapi.NewMessage(m.Chat.ID, "Пришлите пожалуйста фото")
			msg.ReplyToMessageID = m.MessageID
			if _, err := c.bot.Send(msg); err != nil {
				log.Println("send text:", err)
			}
			continue
		}
	}
	return nil
}
