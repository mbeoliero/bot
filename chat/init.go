package chat

import "github.com/mbeoliero/bot/conf"

type Completion interface {
	ChatCompletion(msgList []*ChatMessage) (string, error)
}

var token string

func Init() {
	InitLimiter()

	token = conf.Get().Gpt.ApiKey
}
