package conveyearthgo

import (
	"aletheiaware.com/authgo"
	"time"
)

type Database interface {
	authgo.Database

	LookupAccountBalance(int64) (int64, error)

	CreateConversation(int64, string, time.Time) (int64, error)
	SelectConversation(int64) (int64, string, string, time.Time, error)
	LookupBestConversations(func(int64, int64, string, string, time.Time, int64, int64) error, time.Time, int64) error
	LookupRecentConversations(func(int64, int64, string, string, time.Time, int64, int64) error, int64) error

	CreateMessage(int64, int64, int64, time.Time) (int64, error)
	SelectMessage(int64) (int64, string, int64, int64, time.Time, int64, int64, error)
	LookupMessage(int64, int64) (int64, int64, time.Time, int64, int64, error)
	LookupMessages(int64, func(int64, int64, string, int64, time.Time, int64, int64) error) error
	LookupMessageParent(int64) (int64, error)

	CreateFile(int64, string, string, time.Time) (int64, error)
	SelectFile(int64) (int64, string, string, time.Time, error)
	LookupFiles(int64, func(int64, string, string, time.Time) error) error

	CreateCharge(int64, int64, int64, int64, time.Time) (int64, error)
	CreateYield(int64, int64, int64, int64, int64, time.Time) (int64, error)
	CreatePurchase(int64, string, string, string, string, int64, int64, time.Time) (int64, error)
}
