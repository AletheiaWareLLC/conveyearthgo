package conveyearthgo_test

import (
	"aletheiaware.com/authgo/authtest"
	authdb "aletheiaware.com/authgo/database"
	"aletheiaware.com/conveyearthgo"
	"aletheiaware.com/conveyearthgo/database"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAccount(t *testing.T) {
	now := time.Now()
	db := database.NewInMemory()
	id, err := db.CreateUser(authtest.TEST_EMAIL, authtest.TEST_USERNAME, []byte(authtest.TEST_PASSWORD), now)
	assert.NoError(t, err)
	am := conveyearthgo.NewAccountManager(db)
	a, err := am.Account(authtest.TEST_USERNAME)
	assert.NoError(t, err)
	assert.Equal(t, id, a.ID)
	assert.Equal(t, authtest.TEST_EMAIL, a.Email)
	assert.Equal(t, authtest.TEST_USERNAME, a.Username)
	assert.Equal(t, now, a.Created)
}

func TestAccountBalance(t *testing.T) {
	u1 := authtest.TEST_USER_ID
	u2 := authtest.TEST_USER_ID + 1
	for name, tt := range map[string]struct {
		charge1, charge2,
		yield1,
		purchase1, purchase2,
		award1, award2,
		gift1, gift2,
		amount1, amount2 int64
	}{
		"Empty": {},
		"Charge": {
			charge1: 34,
			charge2: 43,
			amount1: -34,
			amount2: -43,
		},
		"Yield": {
			yield1:  34,
			amount1: 34,
		},
		"Purchase": {
			purchase1: 34,
			purchase2: 43,
			amount1:   34,
			amount2:   43,
		},
		"Award": {
			award1:  34,
			award2:  43,
			amount1: 34,
			amount2: 43,
		},
		"Gift1": {
			gift1:   34,
			amount1: 34,
			amount2: -34,
		},
		"Gift2": {
			gift2:   43,
			amount1: -43,
			amount2: 43,
		},
		"Charge_Yield": {
			charge1: 34,
			yield1:  56,
			amount1: 22,
		},
		"Charge_Purchase": {
			charge1:   34,
			purchase1: 56,
			amount1:   22,
		},
		"Charge_Award": {
			charge1: 34,
			award1:  56,
			amount1: 22,
		},
		"Charge_Gift1": {
			charge1: 34,
			gift1:   56,
			amount1: 22,
			amount2: -56,
		},
		"Charge_Gift2": {
			charge1: 34,
			gift2:   56,
			amount1: -90,
			amount2: 56,
		},
		"Yield_Purchase": {
			yield1:    34,
			purchase1: 56,
			amount1:   90,
		},
		"Yield_Award": {
			yield1:  34,
			award1:  56,
			amount1: 90,
		},
		"Yield_Gift1": {
			yield1:  34,
			gift1:   56,
			amount1: 90,
			amount2: -56,
		},
		"Yield_Gift2": {
			yield1:  34,
			gift2:   56,
			amount1: -22,
			amount2: 56,
		},
		"Purchase_Award": {
			purchase1: 34,
			award1:    56,
			amount1:   90,
		},
		"Purchase_Gift1": {
			purchase1: 34,
			gift1:     56,
			amount1:   90,
			amount2:   -56,
		},
		"Purchase_Gift2": {
			purchase1: 34,
			gift2:     56,
			amount1:   -22,
			amount2:   56,
		},
		"Award_Gift1": {
			award1:  34,
			gift1:   56,
			amount1: 90,
			amount2: -56,
		},
		"Award_Gift2": {
			award1:  34,
			gift2:   56,
			amount1: -22,
			amount2: 56,
		},
		"Gift1_Gift2": {
			gift1:   34,
			gift2:   56,
			amount1: -22,
			amount2: 22,
		},
		"Charge_Yield_Purchase_Award_Gift1_Gift2": {
			charge1:   34,
			charge2:   43,
			yield1:    56,
			purchase1: 78,
			purchase2: 87,
			award1:    90,
			award2:    9,
			gift1:     123,
			gift2:     456,
			amount1:   -143,
			amount2:   386,
		},
	} {
		t.Run(name, func(t *testing.T) {
			now := time.Now()
			db := database.NewInMemory()

			// User 1 starts a conversation
			c1, err := db.CreateConversation(u1, "", now)
			assert.NoError(t, err)
			m1, err := db.CreateMessage(u1, c1, 0, now)
			assert.NoError(t, err)
			_, err = db.CreateCharge(u1, c1, m1, tt.charge1, now)
			assert.NoError(t, err)

			// User 2 gifts to User 1's message
			_, err = db.CreateGift(u2, c1, m1, tt.gift1, now)
			assert.NoError(t, err)

			// User 2 replies to User 1
			m2, err := db.CreateMessage(u2, c1, m1, now)
			assert.NoError(t, err)
			_, err = db.CreateCharge(u2, c1, m2, tt.charge2, now)
			assert.NoError(t, err)
			_, err = db.CreateYield(u2, c1, m2, m1, tt.yield1, now)
			assert.NoError(t, err)

			// User 1 gifts to User 2's reply
			_, err = db.CreateGift(u1, c1, m2, tt.gift2, now)
			assert.NoError(t, err)

			// User 1 purchases coins
			_, err = db.CreatePurchase(u1, "", "", "", "", 0, tt.purchase1, now)
			assert.NoError(t, err)

			// User 2 purchases coins
			_, err = db.CreatePurchase(u2, "", "", "", "", 0, tt.purchase2, now)
			assert.NoError(t, err)

			// User 1 receives an award
			id := authdb.NextId()
			db.AwardId[id] = true
			db.AwardUser[id] = u1
			db.AwardAmount[id] = tt.award1
			db.AwardCreated[id] = now

			// User 2 receives an award
			id = authdb.NextId()
			db.AwardId[id] = true
			db.AwardUser[id] = u2
			db.AwardAmount[id] = tt.award2
			db.AwardCreated[id] = now

			am := conveyearthgo.NewAccountManager(db)

			// Check User 1 balance
			a1, err := am.AccountBalance(u1)
			assert.NoError(t, err)
			assert.Equal(t, tt.amount1, a1)

			// Check User 2 balance
			a2, err := am.AccountBalance(u2)
			assert.NoError(t, err)
			assert.Equal(t, tt.amount2, a2)
		})
	}
}

func TestNewPurchase(t *testing.T) {
	sessionID := "sessionID"
	customerID := "customerID"
	paymentIntentID := "paymentIntentID"
	currency := "USD"
	amount := int64(100)
	size := int64(1000)
	db := database.NewInMemory()
	am := conveyearthgo.NewAccountManager(db)
	err := am.NewPurchase(authtest.TEST_USER_ID, sessionID, customerID, paymentIntentID, currency, amount, size)
	assert.NoError(t, err)

	// Pick first purchase id
	var id int64
	for i, ok := range db.PurchaseId {
		if ok {
			id = i
			break
		}
	}
	assert.Equal(t, authtest.TEST_USER_ID, db.PurchaseUser[id])
	assert.Equal(t, sessionID, db.PurchaseStripeSession[id])
	assert.Equal(t, customerID, db.PurchaseStripeCustomer[id])
	assert.Equal(t, paymentIntentID, db.PurchaseStripePaymentIntent[id])
	assert.Equal(t, currency, db.PurchaseStripeCurrency[id])
	assert.Equal(t, amount, db.PurchaseStripeAmount[id])
	assert.Equal(t, size, db.PurchaseBundleSize[id])
}
