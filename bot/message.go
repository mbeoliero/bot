package bot

import (
	"net/url"
	"os"
	"strings"

	"github.com/Mrs4s/MiraiGo/message"
)

const (
	MentionAll    = "all"
	MentionSingle = "single"
)

type Element interface {
	ToElem() message.IMessageElement
}

type TextMessage struct {
	Text string
}

func NewTextMessage(text string) *TextMessage {
	return &TextMessage{Text: text}
}

func (m *TextMessage) ToElem() message.IMessageElement {
	return message.NewText(m.Text)
}

type MentionMessage struct {
	MentionType string
	Target      int64
	Name        string
}

func NewMentionMessage(id int64, name string, mentionType string) *MentionMessage {
	return &MentionMessage{
		MentionType: mentionType,
		Target:      id,
		Name:        name,
	}
}

func (m *MentionMessage) ToElem() message.IMessageElement {
	switch m.MentionType {
	case MentionAll:
		return message.AtAll()
	case MentionSingle:
		if len(m.Name) > 0 {
			m.Name = "@" + m.Name
		}

		return message.NewAt(m.Target, m.Name)
	default:
		return nil
	}
}

type JsonMessage struct {
	Data string
}

func NewJsonMessage(data string) *JsonMessage {
	return &JsonMessage{Data: data}
}

func (m *JsonMessage) ToElem() message.IMessageElement {
	return message.NewRichJson(m.Data)
}

type ImageMessage struct {
	Filename string
}

func (m *ImageMessage) ToElem() message.IMessageElement {
	file, err := url.Parse(m.Filename)
	if err != nil {
		return nil
	}

	info, err := os.Stat(file.Path)
	if err != nil {
		return nil
	}

	if info.Size() == 0 || info.Size() >= 1024*1024*30 { // 30MB
		return nil
	}

	return &LocalImage{File: file.Path, URL: m.Filename}
}

type ReplyMessage struct {
	Mid    int32
	Time   int32
	Sender int64
	Text   string
}

func NewReplyMessage(mid, time int32, sender int64, text string) *ReplyMessage {
	return &ReplyMessage{
		Mid:    mid,
		Time:   time,
		Sender: sender,
		Text:   text,
	}
}

func (m *ReplyMessage) ToElem() message.IMessageElement {
	return &message.ReplyElement{
		ReplySeq: m.Mid,
		Sender:   m.Sender,
		//GroupID:  0,
		Time:     m.Time,
		Elements: []message.IMessageElement{message.NewText(m.Text)},
	}
}

func GetMessageContent(elemList []message.IMessageElement) (content string, target int64) {
	res := make([]string, 0)
	for _, elem := range elemList {
		switch o := elem.(type) {
		case *message.TextElement:
			res = append(res, o.Content)
		case *message.AtElement:
			if o.Target > 0 {
				target = o.Target
			}
		case *message.ReplyElement:
			newElem := o.Elements[0]
			if v, ok := newElem.(*message.AtElement); ok {
				target = v.Target
			}
		}
	}

	content = strings.Join(res, "\n")
	return
}
