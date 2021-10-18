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
	SelectChargesForUser(int64) (int64, error)
	SelectYieldsForUser(int64) (int64, error)
	SelectPurchasesForUser(int64) (int64, error)
	SelectAwardsForUser(int64) (int64, error)
	SelectGiftsForUser(int64) (int64, error)
	SelectGiftsFromUser(int64) (int64, error)
	CreatePurchase(int64, string, string, string, string, int64, int64, time.Time) (int64, error)
}

func AccountBalance(db AccountDatabase, user int64) (int64, error) {
	charges, err := db.SelectChargesForUser(user)
	if err != nil {
		return 0, err
	}
	yields, err := db.SelectYieldsForUser(user)
	if err != nil {
		return 0, err
	}
	purchases, err := db.SelectPurchasesForUser(user)
	if err != nil {
		return 0, err
	}
	awards, err := db.SelectAwardsForUser(user)
	if err != nil {
		return 0, err
	}
	received, err := db.SelectGiftsForUser(user)
	if err != nil {
		return 0, err
	}
	given, err := db.SelectGiftsFromUser(user)
	if err != nil {
		return 0, err
	}
	return received + awards + purchases + yields - charges - given, nil
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
