package main_test

import (
	"aletheiaware.com/authgo"
	"aletheiaware.com/authgo/authtest"
	"aletheiaware.com/conveyearthgo"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type DB interface {
	authgo.Database
	conveyearthgo.AccountDatabase
	conveyearthgo.ContentDatabase
}

func assertBalance(t *testing.T, db DB, user, balance int64) {
	t.Helper()
	b, err := conveyearthgo.AccountBalance(db, user)
	assert.Nil(t, err)
	assert.Equal(t, balance, b)
}

func AccountBalance(t *testing.T, db DB) {
	t.Helper()
	created := time.Now()

	// Add 2 Users
	hash, err := authgo.GeneratePasswordHash([]byte(authtest.TEST_PASSWORD))
	assert.Nil(t, err)
	user1, err := db.CreateUser("1"+authtest.TEST_EMAIL, authtest.TEST_USERNAME+"1", hash, created)
	assert.Nil(t, err)
	user2, err := db.CreateUser("2"+authtest.TEST_EMAIL, authtest.TEST_USERNAME+"2", hash, created)
	assert.Nil(t, err)

	assertBalance(t, db, user1, 0)
	assertBalance(t, db, user2, 0)

	// Add Purchases
	purchase1, err := db.CreatePurchase(user1, "sessionID1", "customerID1", "paymentIntentID1", "currency", 100, 2000, created)
	assert.Nil(t, err)
	assert.NotEqual(t, 0, purchase1)
	purchase2, err := db.CreatePurchase(user2, "sessionID2", "customerID2", "paymentIntentID2", "currency", 100, 2000, created)
	assert.Nil(t, err)
	assert.NotEqual(t, 0, purchase2)

	assertBalance(t, db, user1, 2000)
	assertBalance(t, db, user2, 2000)

	// Add Conversation, Message, and Charge
	conversation, err := db.CreateConversation(user1, "topic", created)
	assert.Nil(t, err)
	assert.NotEqual(t, 0, conversation)
	message, err := db.CreateMessage(user1, conversation, 0, created)
	assert.Nil(t, err)
	assert.NotEqual(t, 0, message)
	charge1, err := db.CreateCharge(user1, conversation, message, 500, created)
	assert.Nil(t, err)
	assert.NotEqual(t, 0, charge1)

	assertBalance(t, db, user1, 1500)

	// Create Reply, Charge and Yield
	reply, err := db.CreateMessage(user2, conversation, message, created)
	assert.Nil(t, err)
	assert.NotEqual(t, 0, reply)
	charge2, err := db.CreateCharge(user2, conversation, reply, 1000, created)
	assert.Nil(t, err)
	assert.NotEqual(t, 0, charge2)
	yield, err := db.CreateYield(user2, conversation, reply, message, 500, created)
	assert.Nil(t, err)
	assert.NotEqual(t, 0, yield)

	assertBalance(t, db, user1, 2000)
	assertBalance(t, db, user2, 1000)

	// Create Gift
	gift, err := db.CreateGift(user2, conversation, message, 1000, created)
	assert.Nil(t, err)
	assert.NotEqual(t, 0, gift)

	assertBalance(t, db, user1, 3000)
	assertBalance(t, db, user2, 0)
}
