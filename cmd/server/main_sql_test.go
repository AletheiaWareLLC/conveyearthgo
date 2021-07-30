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

	// Create Email Verifier
	ev := authtest.NewEmailVerifier()

	// Create an Authenticator
	return authgo.NewAuthenticator(db, ev)
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
