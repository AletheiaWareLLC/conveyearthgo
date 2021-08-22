package conveyearthgo

import (
	"aletheiaware.com/authgo"
	"errors"
	"log"
	"time"
)

var ErrInsufficientBalance = errors.New("Insufficient Balance")

type AccountDatabase interface {
	SelectUser(string) (int64, string, []byte, time.Time, error)
	SelectCharges(int64) (int64, error)
	SelectYields(int64) (int64, error)
	SelectPurchases(int64) (int64, error)
	CreatePurchase(int64, string, string, string, string, int64, int64, time.Time) (int64, error)
}

func AccountBalance(db AccountDatabase, user int64) (int64, error) {
	charges, err := db.SelectCharges(user)
	if err != nil {
		return 0, err
	}
	yields, err := db.SelectYields(user)
	if err != nil {
		return 0, err
	}
	purchases, err := db.SelectPurchases(user)
	if err != nil {
		return 0, err
	}
	return purchases + yields - charges, nil
}

type AccountManager interface {
	Account(string) (*authgo.Account, error)
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

func (m *accountManager) Account(username string) (*authgo.Account, error) {
	id, email, _, created, err := m.database.SelectUser(username)
	if err != nil {
		return nil, err
	}
	return &authgo.Account{
		ID:       id,
		Username: username,
		Email:    email,
		Created:  created,
	}, nil
}

func (m *accountManager) AccountBalance(user int64) (int64, error) {
	return AccountBalance(m.database, user)
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
