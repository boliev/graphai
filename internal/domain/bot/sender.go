package bot

import (
	"github.com/boliev/graphai/internal/domain"
)

type Client interface {
	Send(msg domain.Message) error
	SendImage(message domain.ImageMessage) error
}
type Sender struct {
	client Client
}

func NewSender(client Client) *Sender {
	return &Sender{
		client: client,
	}
}
func (s *Sender) Send(msg domain.Message) error {
	return s.client.Send(msg)
}

func (s *Sender) Greet(chatId int64) error {
	msg := domain.Message{
		ChatId: chatId,
		Text:   "Приветствую тебя бро!",
	}

	msg.AddButtons(
		domain.NewMessageTypeButtonData("🎨 Хочу красивую картинку", "make_image"),
		domain.NewMessageTypeButtonUrl("💎 Посмотреть идеи", "https://t.me/your_channel"),
	)

	return s.client.Send(msg)
}

func (s *Sender) HelpMakeImage(chatId int64) error {
	msg := domain.Message{
		ChatId: chatId,
		Text:   "Отлично, давай создадим твой новый образ ✨\n\n📸 Просто отправь мне в одном сообщении: фото и описание, что с ним нужно сделать.\n\nНе знаешь, что написать? Жми 👉 Посмотреть идеи (https://t.me/some_some)\n\nЖду фото и описание 👇",
	}

	return s.client.Send(msg)
}

func (s *Sender) SendError(chatId int64) error {
	msg := domain.Message{
		ChatId: chatId,
		Text:   "К сожалению у нас не получилось обработать запрос. Поробуйте другой промт.",
	}

	return s.client.Send(msg)
}

func (s *Sender) HelpPleaseSendPhoto(chatId int64) error {
	msg := domain.Message{
		ChatId: chatId,
		Text:   "Пришлите пожалуйста фото.",
	}

	return s.client.Send(msg)
}

func (s *Sender) SendImage(message domain.ImageMessage) error {
	return s.client.SendImage(message)
}
