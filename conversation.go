package conveyearthgo

import (
	"time"
)

type Conversation struct {
	ID      int64
	User    string
	Topic   string
	Created time.Time
}
