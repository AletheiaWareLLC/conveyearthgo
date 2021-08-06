package conveyearthgo

import (
	"errors"
	"log"
	"time"
)

var ErrInsufficientBalance = errors.New("Insufficient Balance")

type AccountDatabase interface {
	LookupAccountBalance(int64) (int64, error)
	CreatePurchase(int64, string, string, string, string, int64, int64, time.Time) (int64, error)
}

type AccountManager interface {
	AccountBalance(int64) (int64, error)
	NewPurchase(int64, string, string, string, string, int64, int64) error
}

func NewAccountManager(db AccountDatabase) AccountManager {
	return &accountManager{
		database: db,
	}
}

type accountManager struct {
	database AccountDatabase
}

func (m *accountManager) AccountBalance(user int64) (int64, error) {
	balance, err := m.database.LookupAccountBalance(user)
	if err != nil {
		return 0, err
	}
	return balance, nil
}

func (m *accountManager) NewPurchase(user int64, sessionID, customerID, paymentIntentID, currency string, amount, size int64) error {
	created := time.Now()
	purchase, err := m.database.CreatePurchase(user, sessionID, customerID, paymentIntentID, currency, amount, size, created)
	if err != nil {
		return err
	}
	log.Println("Created Purchase", purchase)
	return nil
}
