package conveytest

import (
	"aletheiaware.com/authgo"
	"aletheiaware.com/conveyearthgo"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	TEST_SESSION_ID        = "sessionID"
	TEST_CUSTOMER_ID       = "customerID"
	TEST_PAYMENT_INTENT_ID = "paymentIntentID"
	TEST_CURRENCY          = "USD"
	TEST_PURCHASE_AMOUNT   = 10
	TEST_PURCHASE_SIZE     = 1000
)

func NewPurchase(t *testing.T, am conveyearthgo.AccountManager, acc *authgo.Account) {
	err := am.NewPurchase(acc.ID, TEST_SESSION_ID, TEST_CUSTOMER_ID, TEST_PAYMENT_INTENT_ID, TEST_CURRENCY, TEST_PURCHASE_AMOUNT, TEST_PURCHASE_SIZE)
	assert.NoError(t, err)
}
