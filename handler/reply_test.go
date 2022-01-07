package handler_test

import (
	"aletheiaware.com/authgo"
	"aletheiaware.com/authgo/authtest"
	"aletheiaware.com/conveyearthgo"
	"aletheiaware.com/conveyearthgo/conveytest"
	"aletheiaware.com/conveyearthgo/database"
	"aletheiaware.com/conveyearthgo/filesystem"
	"aletheiaware.com/conveyearthgo/handler"
	"fmt"
	"github.com/stretchr/testify/assert"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestReply(t *testing.T) {
	dir, err := os.MkdirTemp("", "test")
	assert.Nil(t, err)
	fs := filesystem.NewOnDisk(dir)
	defer os.RemoveAll(dir)
	tmpl, err := template.New("reply.go.html").Parse(`{{.Error}}{{with .Account}}{{.Username}}{{end}}`)
	assert.Nil(t, err)
	t.Run("Returns 200 When Signed In And Conversation And Message Exist", func(t *testing.T) {
		db := database.NewInMemory()
		ev := authtest.NewEmailVerifier()
		auth := authgo.NewAuthenticator(db, ev)
		acc := authtest.NewTestAccount(t, auth)
		token, _ := authtest.SignIn(t, auth)
		am := conveyearthgo.NewAccountManager(db)
		cm := conveyearthgo.NewContentManager(db, fs)
		nm := conveyearthgo.NewNotificationManager(db, conveytest.NewNotificationSender())
		c, m, _ := conveytest.NewConversation(t, cm, acc)
		mux := http.NewServeMux()
		handler.AttachReplyHandler(mux, auth, am, cm, nm, tmpl)
		request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/reply?conversation=%d&message=%d", c.ID, m.ID), nil)
		request.AddCookie(auth.NewSignInSessionCookie(token))
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusOK, result.StatusCode)
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Equal(t, authtest.TEST_USERNAME, string(body))
	})
	t.Run("Returns 404 When Conversation And Message Do Not Exist", func(t *testing.T) {
		db := database.NewInMemory()
		ev := authtest.NewEmailVerifier()
		auth := authgo.NewAuthenticator(db, ev)
		authtest.NewTestAccount(t, auth)
		token, _ := authtest.SignIn(t, auth)
		am := conveyearthgo.NewAccountManager(db)
		cm := conveyearthgo.NewContentManager(db, fs)
		nm := conveyearthgo.NewNotificationManager(db, conveytest.NewNotificationSender())
		mux := http.NewServeMux()
		handler.AttachReplyHandler(mux, auth, am, cm, nm, tmpl)
		request := httptest.NewRequest(http.MethodGet, "/reply?conversation=10&message=10", nil)
		request.AddCookie(auth.NewSignInSessionCookie(token))
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusNotFound, result.StatusCode)
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusText(http.StatusNotFound)+"\n", string(body))
	})
	t.Run("Redirects When Not Signed In", func(t *testing.T) {
		db := database.NewInMemory()
		ev := authtest.NewEmailVerifier()
		auth := authgo.NewAuthenticator(db, ev)
		am := conveyearthgo.NewAccountManager(db)
		cm := conveyearthgo.NewContentManager(db, fs)
		nm := conveyearthgo.NewNotificationManager(db, conveytest.NewNotificationSender())
		mux := http.NewServeMux()
		handler.AttachReplyHandler(mux, auth, am, cm, nm, tmpl)
		request := httptest.NewRequest(http.MethodGet, "/reply?conversation=10&message=10", nil)
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusFound, result.StatusCode)
		u, err := result.Location()
		assert.Nil(t, err)
		assert.Equal(t, "/sign-in?next=%2Freply%3Fconversation%3D10%26message%3D10", u.String())
	})
	// TODO Reply Success
	// TODO Reply Conversation Doesn't Exist
	// TODO Reply Message Doesn't Exist
	// TODO Reply Content Too Short
	// TODO Reply Insufficient Balance
	// TODO
}
