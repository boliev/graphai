package domain

const MessageButtonTypeData = "data"
const MessageButtonTypeUrl = "url"

type Message struct {
	ChatId  int64
	Text    string
	Buttons []MessageButton
}

type MessageButton struct {
	Type string
	Text string
	Data string
}

func (m *Message) AddButtons(buttons ...MessageButton) {
	m.Buttons = append(m.Buttons, buttons...)
}

func NewMessageTypeButtonData(text string, data string) MessageButton {
	return MessageButton{
		Type: MessageButtonTypeData,
		Text: text,
		Data: data,
	}
}

func NewMessageTypeButtonUrl(text string, data string) MessageButton {
	return MessageButton{
		Type: MessageButtonTypeUrl,
		Text: text,
		Data: data,
	}
}

type ImageMessage struct {
	ChatId           int64
	ReplyToMessageId int
	Ext              string
	Image            []byte
}
