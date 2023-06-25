package chat

import "github.com/sashabaranov/go-openai"

type RoleType int32

const (
	RoleTypeUser   RoleType = 0
	RoleTypeBot    RoleType = 1
	RoleTypeSystem RoleType = 2
)

type ChatMessage struct {
	Content  string
	Author   string
	RoleType RoleType
}

func (c *ChatMessage) ParseRoleType() string {
	switch c.RoleType {
	case RoleTypeBot:
		return openai.ChatMessageRoleAssistant
	case RoleTypeSystem:
		return openai.ChatMessageRoleSystem
	case RoleTypeUser:
		return openai.ChatMessageRoleUser
	default:
		return openai.ChatMessageRoleUser
	}
}
