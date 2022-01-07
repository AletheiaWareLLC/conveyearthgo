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
	"strings"
	"testing"
)

func TestBest(t *testing.T) {
	dir, err := os.MkdirTemp("", "test")
	assert.Nil(t, err)
	fs := filesystem.NewOnDisk(dir)
	defer os.RemoveAll(dir)
	tmpl, err := template.New("best.go.html").Parse(`{{with .Account}}{{.Username}}{{end}}{{range .Conversations}}{{.Topic}}{{end}}`)
	assert.Nil(t, err)
	t.Run("Returns 200 With No Conversations", func(t *testing.T) {
		db := database.NewInMemory()
		ev := authtest.NewEmailVerifier()
		auth := authgo.NewAuthenticator(db, ev)
		authtest.NewTestAccount(t, auth)
		token, _ := authtest.SignIn(t, auth)
		cm := conveyearthgo.NewContentManager(db, fs)
		mux := http.NewServeMux()
		handler.AttachBestHandler(mux, auth, cm, tmpl, 1, 1)
		request := httptest.NewRequest(http.MethodGet, "/best", nil)
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
		cm := conveyearthgo.NewContentManager(db, fs)
		ev := authtest.NewEmailVerifier()
		auth := authgo.NewAuthenticator(db, ev)
		acc := authtest.NewTestAccount(t, auth)
		token, _ := authtest.SignIn(t, auth)
		c, m, _ := conveytest.NewConversation(t, cm, acc)
		conveytest.NewReply(t, cm, acc, c, m)
		mux := http.NewServeMux()
		handler.AttachBestHandler(mux, auth, cm, tmpl, 1, 1)
		request := httptest.NewRequest(http.MethodGet, "/best", nil)
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
			c, m, _, err := cm.NewConversation(acc, topic, []string{hash}, []string{conveyearthgo.MIME_TEXT_PLAIN}, []int64{size})
			assert.NoError(t, err)

			// Add a Reply
			hash, size, err = cm.AddText([]byte(strings.Repeat("Hi!", limit-i+1)))
			assert.NoError(t, err)
			_, _, err = cm.NewMessage(acc, c.ID, m.ID, []string{hash}, []string{conveyearthgo.MIME_TEXT_PLAIN}, []int64{size})
			assert.NoError(t, err)
		}
		mux := http.NewServeMux()
		handler.AttachBestHandler(mux, auth, cm, tmpl, 1, int64(maximum))
		request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/best?limit=%d", limit), nil)
		request.AddCookie(auth.NewSignInSessionCookie(token))
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusOK, result.StatusCode)
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		// Should display 6 (limit) best conversations
		assert.Equal(t, authtest.TEST_USERNAME+"FooBar1FooBar2FooBar3FooBar4FooBar5FooBar6", string(body))
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
			c, m, _, err := cm.NewConversation(acc, topic, []string{hash}, []string{conveyearthgo.MIME_TEXT_PLAIN}, []int64{size})
			assert.NoError(t, err)

			// Add a Reply
			hash, size, err = cm.AddText([]byte(strings.Repeat("Hi!", limit-i+1)))
			assert.NoError(t, err)
			_, _, err = cm.NewMessage(acc, c.ID, m.ID, []string{hash}, []string{conveyearthgo.MIME_TEXT_PLAIN}, []int64{size})
			assert.NoError(t, err)
		}
		mux := http.NewServeMux()
		handler.AttachBestHandler(mux, auth, cm, tmpl, 1, int64(maximum))
		request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/best?limit=%d", limit), nil)
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusOK, result.StatusCode)
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		// Should only display 2 (maximum) best conversations
		assert.Equal(t, "FooBar1FooBar2", string(body))
	})
	// TODO Best of the Day
	// TODO Best of the Week
	// TODO Best of the Month
	// TODO Best of the Year
	// TODO Best of All
}
