package main_test

import (
	"aletheiaware.com/authgo"
	"aletheiaware.com/authgo/authtest"
	"aletheiaware.com/authgo/authtest/handler"
	"aletheiaware.com/conveyearthgo/database"
	"embed"
	"github.com/golang-migrate/migrate/v4"
	"github.com/stretchr/testify/assert"
	"io/fs"
	"path"
	"testing"
	"time"
)

//go:embed assets
var embeddedFS embed.FS

func NewSqlAuthenticator(t *testing.T) authgo.Authenticator {
	/*
		CREATE DATABASE testdb;
		CREATE USER 'tester'@'localhost' IDENTIFIED BY 'tester123';
		GRANT ALL PRIVILEGES ON testdb.* TO 'tester'@'localhost';
		FLUSH PRIVILEGES;

		SHOW GRANTS FOR 'tester'@'localhost';

		DROP USER 'tester'@'localhost';

		migrate -database 'mysql://tester:tester123@tcp(localhost:3306)/testdb' -path cmd/server/assets/database/migrations/ up
	*/

	// Create database
	db := NewSqlDatabase(t)

	// Create Email Verifier
	ev := authtest.NewEmailVerifier()

	// Create an Authenticator
	return authgo.NewAuthenticator(db, ev)
}

func NewSqlDatabase(t *testing.T) *database.Sql {
	// Create SQL Database
	db, err := database.NewSql("testdb", "tester", "tester123", "", "", false)
	assert.Nil(t, err)

	migrationFS, err := fs.Sub(embeddedFS, path.Join("assets", "database", "migrations"))
	assert.Nil(t, err)

	m, err := db.Migrator(migrationFS)
	assert.Nil(t, err)

	// Migrate Down to wipe database
	if err := m.Down(); err != nil {
		assert.Equal(t, migrate.ErrNoChange, err)
	}

	// Migrate Up
	assert.Nil(t, m.Up())
	return db
}

func TestSqlAccount(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping SQL database test in short mode.")
	}
	handler.Account(t, NewSqlAuthenticator)
}

func TestSqlAccountPassword(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping SQL database test in short mode.")
	}
	handler.AccountPassword(t, NewSqlAuthenticator)
}

func TestSqlAccountRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping SQL database test in short mode.")
	}
	handler.AccountRecovery(t, NewSqlAuthenticator)
}

func TestSqlAccountRecoveryVerification(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping SQL database test in short mode.")
	}
	handler.AccountRecoveryVerification(t, NewSqlAuthenticator)
}

func TestSqlSignIn(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping SQL database test in short mode.")
	}
	handler.SignIn(t, NewSqlAuthenticator)
}

func TestSqlSignOut(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping SQL database test in short mode.")
	}
	handler.SignOut(t, NewSqlAuthenticator)
}

func TestSqlSignUp(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping SQL database test in short mode.")
	}
	handler.SignUp(t, NewSqlAuthenticator)
}

func TestSqlSignUpVerification(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping SQL database test in short mode.")
	}
	handler.SignUpVerification(t, NewSqlAuthenticator)
}

func TestSqlSignUpSignOutSignInAccount(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping SQL database test in short mode.")
	}
	handler.SignUpSignOutSignInAccount(t, NewSqlAuthenticator)
}

func TestSqlAccountPasswordSignOutSignInAccount(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping SQL database test in short mode.")
	}
	handler.SignUpSignOutSignInAccount(t, NewSqlAuthenticator)
}

func TestSqlAccountRecoveryAccountPasswordAccount(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping SQL database test in short mode.")
	}
	handler.SignUpSignOutSignInAccount(t, NewSqlAuthenticator)
}

func assertBalance(t *testing.T, db *database.Sql, user, balance int64) {
	t.Helper()
	charges, err := db.LookupCharges(user)
	assert.Nil(t, err)
	yields, err := db.LookupYields(user)
	assert.Nil(t, err)
	purchases, err := db.LookupPurchases(user)
	assert.Nil(t, err)
	assert.Equal(t, balance, purchases+yields-charges)
}

func TestSql_LookupAccountBalance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping SQL database test in short mode.")
	}

	db := NewSqlDatabase(t)

	created := time.Now()

	// Add User
	hash, err := authgo.GeneratePasswordHash([]byte(authtest.TEST_PASSWORD))
	assert.Nil(t, err)
	user, err := db.CreateUser(authtest.TEST_EMAIL, authtest.TEST_USERNAME, hash, created)
	assert.Nil(t, err)

	assertBalance(t, db, user, 0)

	// Add Purchase
	purchase, err := db.CreatePurchase(user, "sessionID", "customerID", "paymentIntentID", "currency", 100, 2000, created)
	assert.Nil(t, err)
	assert.NotEqual(t, 0, purchase)

	assertBalance(t, db, user, 2000)

	// Add Conversation, Message, and Charge
	conversation, err := db.CreateConversation(user, "topic", created)
	assert.Nil(t, err)
	assert.NotEqual(t, 0, conversation)
	message, err := db.CreateMessage(user, conversation, 0, created)
	assert.Nil(t, err)
	assert.NotEqual(t, 0, message)
	charge, err := db.CreateCharge(user, conversation, message, 500, created)
	assert.Nil(t, err)
	assert.NotEqual(t, 0, charge)

	assertBalance(t, db, user, 1500)

	// Create Reply, Charge and Yield
	reply, err := db.CreateMessage(user, conversation, message, created)
	assert.Nil(t, err)
	assert.NotEqual(t, 0, reply)
	charge2, err := db.CreateCharge(user, conversation, reply, 1000, created)
	assert.Nil(t, err)
	assert.NotEqual(t, 0, charge2)
	yield, err := db.CreateYield(user, conversation, reply, message, 500, created)
	assert.Nil(t, err)
	assert.NotEqual(t, 0, yield)

	assertBalance(t, db, user, 1000)
}
