package conveyearthgo

import (
	"aletheiaware.com/authgo"
	"time"
)

type Message struct {
	ID             int64
	Author         *authgo.Account
	ConversationID int64
	ParentID       int64
	Created        time.Time
	Cost           int64
	Yield          int64
}
