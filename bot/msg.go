package bot

import (
	"io"

	"github.com/Mrs4s/MiraiGo/message"
)

// LocalImage local image
type LocalImage struct {
	Stream io.ReadSeeker
	File   string
	URL    string

	Flash    bool
	EffectID int32
}

// Type implements the message.IMessageElement.
func (e *LocalImage) Type() message.ElementType {
	return message.Image
}
