package conveyearthgo

import (
	"errors"
	"log"
	"time"
)

var ErrInsufficientBalance = errors.New("Insufficient Balance")

type AccountManager interface {
	AccountBalance(int64) (int64, error)
	NewPurchase(int64, string, string, string, string, int64, int64) error
}

func NewAccountManager(db Database) AccountManager {
	return &accountManager{
		database: db,
	}
}

type accountManager struct {
	database Database
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
