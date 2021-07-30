package conveyearthgo

import (
	"time"
)

type Message struct {
	ID           int64
	User         string
	Conversation int64
	Parent       int64
	Created      time.Time
	Cost         int64
	Yield        int64
}
