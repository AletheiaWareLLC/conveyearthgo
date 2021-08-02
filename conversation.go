package conveyearthgo

import (
	"time"
)

type Conversation struct {
	ID      int64
	User    string
	Topic   string
	Cost    int64
	Yield   int64
	Created time.Time
}
