package conveyearthgo

import (
	"aletheiaware.com/authgo"
	"time"
)

type Conversation struct {
	ID      int64
	Author  *authgo.Account
	Topic   string
	Cost    int64
	Yield   int64
	Created time.Time
}
