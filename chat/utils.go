package chat

import "errors"

func GetResponseFromGpt(content string) (string, error) {
	if !Acquire() {
		return "", errors.New("system rate limiter")
	}

	g := &Gpt{Token: token}
	msgList := []*ChatMessage{
		&ChatMessage{
			Content:  content,
			RoleType: RoleTypeUser,
		},
	}

	return GetResponse(g, msgList)
}

func GetResponse(comp Completion, msgList []*ChatMessage) (string, error) {
	return comp.ChatCompletion(msgList)
}
