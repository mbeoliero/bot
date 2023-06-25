package chat

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

type Gpt struct {
	Token string
}

func NewGpt(token string) *Gpt {
	return &Gpt{
		Token: token,
	}
}

func (g *Gpt) ParseChatMessage(msgList []*ChatMessage) []openai.ChatCompletionMessage {
	result := make([]openai.ChatCompletionMessage, 0, len(msgList))
	for _, msg := range msgList {
		role := msg.ParseRoleType()

		result = append(result, openai.ChatCompletionMessage{
			Role:    role,
			Content: msg.Content,
		})
	}

	return result
}

func (g *Gpt) ChatCompletion(msgList []*ChatMessage) (string, error) {
	client := openai.NewClient(g.Token)
	req := openai.ChatCompletionRequest{
		Model:    openai.GPT3Dot5Turbo,
		Messages: g.ParseChatMessage(msgList),
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		req,
	)
	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}
