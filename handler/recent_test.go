package handler_test

import (
	"aletheiaware.com/authgo"
	"aletheiaware.com/authgo/authtest"
	"aletheiaware.com/conveyearthgo"
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
	"strings"
	"testing"
	"time"
)

func TestRecent(t *testing.T) {
	dir, err := os.MkdirTemp("", "test")
	assert.Nil(t, err)
	fs := filesystem.NewOnDisk(dir)
	defer os.RemoveAll(dir)
	tmpl, err := template.New("recent.go.html").Parse(`{{with .Account}}{{.Username}}{{end}}{{range .Conversations}}{{.Topic}}{{end}}`)
	assert.Nil(t, err)
	t.Run("Returns 200 With No Conversations", func(t *testing.T) {
		db := database.NewInMemory()
		ev := authtest.NewEmailVerifier()
		auth := authgo.NewAuthenticator(db, ev)
		authtest.NewTestAccount(t, auth)
		token, _ := authtest.SignIn(t, auth)
		cm := conveyearthgo.NewContentManager(db, fs)
		mux := http.NewServeMux()
		handler.AttachRecentHandler(mux, auth, cm, tmpl, 1, 1)
		request := httptest.NewRequest(http.MethodGet, "/recent", nil)
		request.AddCookie(auth.NewSignInSessionCookie(token))
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusOK, result.StatusCode)
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Equal(t, authtest.TEST_USERNAME, string(body))
	})
	t.Run("Returns 200 With One Conversation", func(t *testing.T) {
		db := database.NewInMemory()
		ev := authtest.NewEmailVerifier()
		auth := authgo.NewAuthenticator(db, ev)
		acc := authtest.NewTestAccount(t, auth)
		token, _ := authtest.SignIn(t, auth)
		_, err := db.CreateConversation(acc.ID, "FooBar", time.Now())
		assert.Nil(t, err)
		cm := conveyearthgo.NewContentManager(db, fs)
		mux := http.NewServeMux()
		handler.AttachRecentHandler(mux, auth, cm, tmpl, 1, 1)
		request := httptest.NewRequest(http.MethodGet, "/recent", nil)
		request.AddCookie(auth.NewSignInSessionCookie(token))
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusOK, result.StatusCode)
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Equal(t, authtest.TEST_USERNAME+"FooBar", string(body))
	})
	t.Run("Results Not Limited When Signed In", func(t *testing.T) {
		db := database.NewInMemory()
		cm := conveyearthgo.NewContentManager(db, fs)
		ev := authtest.NewEmailVerifier()
		auth := authgo.NewAuthenticator(db, ev)
		acc := authtest.NewTestAccount(t, auth)
		token, _ := authtest.SignIn(t, auth)

		maximum := 2
		limit := maximum * 3
		for i := 1; i <= limit; i++ {
			// Create Conversation
			topic := fmt.Sprintf("FooBar%d", i)
			hash, size, err := cm.AddText([]byte(fmt.Sprintf("Hello World%d!", i)))
			assert.NoError(t, err)
			mime := "text/plain"
			c, m, _, err := cm.NewConversation(acc, topic, []string{hash}, []string{mime}, []int64{size})
			assert.NoError(t, err)

			// Add a Reply
			hash, size, err = cm.AddText([]byte(strings.Repeat("Hi!", i)))
			assert.NoError(t, err)
			_, _, err = cm.NewMessage(acc, c.ID, m.ID, []string{hash}, []string{mime}, []int64{size})
			assert.NoError(t, err)
		}
		mux := http.NewServeMux()
		handler.AttachRecentHandler(mux, auth, cm, tmpl, 1, int64(maximum))
		request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/recent?limit=%d", limit), nil)
		request.AddCookie(auth.NewSignInSessionCookie(token))
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusOK, result.StatusCode)
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		// Should display 6 (limit) recent conversations
		assert.Equal(t, authtest.TEST_USERNAME+"FooBar6FooBar5FooBar4FooBar3FooBar2FooBar1", string(body))
	})
	t.Run("Results Limited When Not Signed In", func(t *testing.T) {
		db := database.NewInMemory()
		cm := conveyearthgo.NewContentManager(db, fs)
		ev := authtest.NewEmailVerifier()
		auth := authgo.NewAuthenticator(db, ev)
		acc := authtest.NewTestAccount(t, auth)

		maximum := 2
		limit := maximum * 3
		for i := 1; i <= limit; i++ {
			// Create Conversation
			topic := fmt.Sprintf("FooBar%d", i)
			hash, size, err := cm.AddText([]byte(fmt.Sprintf("Hello World%d!", i)))
			assert.NoError(t, err)
			mime := "text/plain"
			c, m, _, err := cm.NewConversation(acc, topic, []string{hash}, []string{mime}, []int64{size})
			assert.NoError(t, err)

			// Add a Reply
			hash, size, err = cm.AddText([]byte(strings.Repeat("Hi!", i)))
			assert.NoError(t, err)
			_, _, err = cm.NewMessage(acc, c.ID, m.ID, []string{hash}, []string{mime}, []int64{size})
			assert.NoError(t, err)
		}
		mux := http.NewServeMux()
		handler.AttachRecentHandler(mux, auth, cm, tmpl, 1, int64(maximum))
		request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/recent?limit=%d", limit), nil)
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusOK, result.StatusCode)
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		// Should only display 2 (maximum) recent conversations
		assert.Equal(t, "FooBar6FooBar5", string(body))
	})
}
