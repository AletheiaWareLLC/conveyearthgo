package conveyearthgo

import (
	"aletheiaware.com/authgo"
	"errors"
	"time"
)

var ErrSelfGiftingNotPermitted = errors.New("Self-Gifting Not Permitted")

type Gift struct {
	ID             int64
	Author         *authgo.Account
	ConversationID int64
	MessageID      int64
	Amount         int64
	Created        time.Time
}
