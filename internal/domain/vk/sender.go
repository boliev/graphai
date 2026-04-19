package vk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/longpoll-bot"
	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/boliev/graphai/internal/pkg/config"
)

const greetMessage = `Привет! Я помогу красиво обработать фото ✨
Отправь фотографию и напиши, что хочешь изменить или добавить.
Например: “сделай фото более нежным”, “замени фон”, “сделай красивую открытку” 💖`

const helpMessage = `Пришлите фото и напишите, что хотите изменить.
Например: заменить фон, сделать фото нежнее, убрать лишнее или оформить как открытку.
Чем понятнее описание, тем лучше результат ✨`

const pricesMessage = `первое фото бесплатно
 - 1 фото — 10 Голосов
 - 5 фото — 45 Голосов
 - 10 фото — 80 Голосов
 - 25 фото — 200 Голосов`

const notEnoughMoneyMessage = `Похоже, бесплатная первая обработка уже использована ✨
Сейчас на балансе недостаточно кредитов для новой обработки.
Пополните баланс, и я с радостью помогу подготовить для вас красивый результат 💖`

type Sender struct {
	vk            *api.VK
	groupID       int
	communityLink string
	paymentAppID  int64
	log           *slog.Logger
}

func NewSender(vk *api.VK, cfg *config.Cfg, log *slog.Logger) (*Sender, error) {
	groupResp, err := vk.GroupsGetByID(api.Params{})
	if err != nil {
		return nil, fmt.Errorf("groups.getById failed: %w", err)
	}

	if len(groupResp) == 0 {
		return nil, errors.New("groups.getById returned empty response")
	}

	return &Sender{
		vk:            vk,
		groupID:       groupResp[0].ID,
		communityLink: cfg.CommunityLink,
		paymentAppID:  cfg.PaymentAppID,
		log:           log,
	}, nil
}

func (s *Sender) send(peerID, messageID, ownerID, photoID int) error {
	attachment := fmt.Sprintf("photo%d_%d", ownerID, photoID)

	kb, err := s.KB(false)
	if err != nil {
		return err
	}

	_, err = s.vk.MessagesSend(api.Params{
		"peer_id":    peerID,
		"random_id":  time.Now().UnixNano(),
		"message":    "",
		"attachment": attachment,
		"reply_to":   messageID,
		"keyboard":   kb,
	})

	return err
}

func (s *Sender) sendText(peerID int64, text string) error {
	kb, err := s.KB(true)
	if err != nil {
		return err
	}

	_, err = s.vk.MessagesSend(api.Params{
		"peer_id":   peerID,
		"random_id": time.Now().UnixNano(),
		"message":   text,
		"keyboard":  kb,
	})

	return err
}

func (s *Sender) sendKB(peerID int) error {
	kb, err := s.KB(true)
	if err != nil {
		return err
	}

	_, err = s.vk.MessagesSend(api.Params{
		"peer_id":   peerID,
		"random_id": time.Now().UnixNano(),
		"message":   greetMessage,
		"keyboard":  kb,
	})

	return err
}

func (s *Sender) extractPhotoURLs(attachments []object.MessagesMessageAttachment) []string {
	var urls []string

	for _, att := range attachments {
		if att.Type != "photo" {
			continue
		}

		mx := att.Photo.MaxSize()
		if mx.URL != "" {
			urls = append(urls, mx.URL)
		}
	}

	return urls
}

func (s *Sender) loadFullMessage(ctx context.Context, msg object.MessagesMessage) (object.MessagesMessage, error) {
	// Предпочтительно дочитывать по conversation_message_id внутри конкретного peer.
	if msg.ConversationMessageID != 0 {
		resp, err := s.vk.MessagesGetByConversationMessageID(api.Params{
			"peer_id":                  msg.PeerID,
			"conversation_message_ids": msg.ConversationMessageID,
		}.WithContext(ctx))
		if err == nil && len(resp.Items) > 0 {
			return resp.Items[0], nil
		}
	}

	// Фолбэк на обычный message_id.
	if msg.ID != 0 {
		resp, err := s.vk.MessagesGetByID(api.Params{
			"message_ids": msg.ID,
		}.WithContext(ctx))
		if err == nil && len(resp.Items) > 0 {
			return resp.Items[0], nil
		}
	}

	return msg, errors.New("full message not found")
}

func (s *Sender) getGroupId() (int, error) {
	groupResp, err := s.vk.GroupsGetByID(api.Params{})
	if err != nil {
		return 0, fmt.Errorf("groups.getById failed: %w", err)
	}

	if len(groupResp) == 0 {
		return 0, errors.New("groups.getById returned empty response")
	}

	return groupResp[0].ID, nil
}

func (s *Sender) getLP() (*longpoll.LongPoll, error) {
	groupID, err := s.getGroupId()
	if err != nil {
		return nil, err
	}

	lp, err := longpoll.NewLongPoll(s.vk, groupID)
	if err != nil {
		return nil, err
	}

	return lp, nil
}

func (s *Sender) uploadMessagesPhoto(peerID int64, photo []byte) (api.PhotosSaveMessagesPhotoResponse, error) {
	resultPhoto, err := s.vk.UploadMessagesPhoto(int(peerID), bytes.NewReader(photo))
	if err != nil {
		s.log.Warn("upload vk messages photo", "error", err)
		return nil, err
	}

	if len(resultPhoto) == 0 {
		s.log.Warn("upload vk messages photo return empty response", "error", err)
		return nil, err
	}

	return resultPhoto, nil
}

func (s *Sender) help(peerID int64) error {
	return s.sendText(peerID, helpMessage)
}

func (s *Sender) prices(peerID int64) error {
	return s.sendText(peerID, pricesMessage)
}

func (s *Sender) notEnoughMoney(peerID int64) error {
	return s.sendText(peerID, notEnoughMoneyMessage)
}

func (s *Sender) KB(inline bool) (string, error) {
	type Action struct {
		Type    string `json:"type"` // text, open_link, callback ...
		Label   string `json:"label"`
		Payload string `json:"payload,omitempty"`
		Link    string `json:"link,omitempty"`
		AppID   int64  `json:"app_id,omitempty"`
	}

	type Button struct {
		Action Action `json:"action"`
		Color  string `json:"color,omitempty"`
	}

	type Keyboard struct {
		OneTime bool       `json:"one_time"`
		Inline  bool       `json:"inline"`
		Buttons [][]Button `json:"buttons"`
	}

	kb := Keyboard{
		OneTime: false, // false = клавиатура будет оставаться в диалоге
		Inline:  inline,
		Buttons: [][]Button{
			{
				{
					Action: Action{
						Type:    "text",
						Label:   "❓ Как пользоваться",
						Payload: `{"cmd":"help"}`,
					},
					Color: "secondary",
				},
				{
					Action: Action{
						Type:  "open_link",
						Label: "💖 Идеи и вдохновение",
						Link:  s.communityLink,
					},
				},
			},
			{
				{
					Action: Action{
						Type:    "text",
						Label:   "💰 Цены",
						Payload: `{"cmd":"prices"}`,
					},
					Color: "secondary",
				},
				{
					Action: Action{
						Type:  "open_app",
						Label: "💳 Оплатить",
						AppID: s.paymentAppID,
					},
				},
			},
		},
	}

	kbJSON, err := json.Marshal(kb)
	if err != nil {
		return "", err
	}

	return string(kbJSON), nil
}
