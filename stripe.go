package conveyearthgo

import (
	"aletheiaware.com/authgo"
	"fmt"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/account"
	"log"
	"strings"
	"time"
)

func FormatStripeAmount(amount float64, currency stripe.Currency) string {
	switch currency {
	case stripe.CurrencyUSD:
		// Format as decimal with dollar size
		price := fmt.Sprintf("$%.2f", amount/100.0)
		// Remove trailing zeros
		price = strings.TrimRight(price, "0")
		// Remove trailing point
		price = strings.TrimRight(price, ".")
		return price
	default:
		log.Println("Unhandled currency:", currency)
	}
	return ""
}

type StripeDatabase interface {
	CreateStripeAccount(int64, string, time.Time) (int64, error)
	SelectStripeAccount(int64) (string, time.Time, error)
}

type StripeManager interface {
	NewStripeAccount(*authgo.Account, *stripe.Account) error
	StripeAccount(*authgo.Account) (*stripe.Account, error)
}

func NewStripeManager(db StripeDatabase) StripeManager {
	return &stripeManager{
		database: db,
	}
}

type stripeManager struct {
	database StripeDatabase
}

func (m *stripeManager) NewStripeAccount(a *authgo.Account, s *stripe.Account) error {
	created := time.Now()
	sa, err := m.database.CreateStripeAccount(a.ID, s.ID, created)
	if err != nil {
		return err
	}
	log.Println("Created Stripe Account", sa)
	return nil
}

func (m *stripeManager) StripeAccount(a *authgo.Account) (*stripe.Account, error) {
	id, _, err := m.database.SelectStripeAccount(a.ID)
	if err != nil {
		return nil, err
	}
	s, err := account.GetByID(id, nil)
	if err != nil {
		return nil, err
	}
	return s, nil
}
