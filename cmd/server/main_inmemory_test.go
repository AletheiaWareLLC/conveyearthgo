package main_test

import (
	"aletheiaware.com/authgo"
	"aletheiaware.com/authgo/authtest"
	"aletheiaware.com/authgo/authtest/handler"
	"aletheiaware.com/conveyearthgo/database"
	"testing"
)

func NewInMemoryAuthenticator(t *testing.T) authgo.Authenticator {
	// Create In-Memory Database
	db := database.NewInMemory()

	// Create Email Verifier
	ev := authtest.NewEmailVerifier()

	// Create an Authenticator
	return authgo.NewAuthenticator(db, ev)
}

func TestInMemoryAccount(t *testing.T) {
	handler.Account(t, NewInMemoryAuthenticator)
}

func TestInMemoryAccountPassword(t *testing.T) {
	handler.AccountPassword(t, NewInMemoryAuthenticator)
}

func TestInMemoryAccountRecovery(t *testing.T) {
	handler.AccountRecovery(t, NewInMemoryAuthenticator)
}

func TestInMemoryAccountRecoveryVerification(t *testing.T) {
	handler.AccountRecoveryVerification(t, NewInMemoryAuthenticator)
}

func TestInMemorySignIn(t *testing.T) {
	handler.SignIn(t, NewInMemoryAuthenticator)
}

func TestInMemorySignOut(t *testing.T) {
	handler.SignOut(t, NewInMemoryAuthenticator)
}

func TestInMemorySignUp(t *testing.T) {
	handler.SignUp(t, NewInMemoryAuthenticator)
}

func TestInMemorySignUpVerification(t *testing.T) {
	handler.SignUpVerification(t, NewInMemoryAuthenticator)
}

func TestInMemorySignUpSignOutSignInAccount(t *testing.T) {
	handler.SignUpSignOutSignInAccount(t, NewInMemoryAuthenticator)
}

func TestInMemoryAccountPasswordSignOutSignInAccount(t *testing.T) {
	handler.SignUpSignOutSignInAccount(t, NewInMemoryAuthenticator)
}

func TestInMemoryAccountRecoveryAccountPasswordAccount(t *testing.T) {
	handler.SignUpSignOutSignInAccount(t, NewInMemoryAuthenticator)
}
