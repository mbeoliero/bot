package chat

import (
	"time"

	"golang.org/x/time/rate"
)

var limiter *rate.Limiter

func InitLimiter() {
	limit := rate.Every(20 * time.Second)
	limiter = rate.NewLimiter(limit, 3)
}

func Acquire() bool {
	return limiter.Allow()
}
