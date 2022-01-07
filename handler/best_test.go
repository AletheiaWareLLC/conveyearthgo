package handler_test

import (
	"aletheiaware.com/authgo"
	"aletheiaware.com/authgo/authtest"
	"aletheiaware.com/conveyearthgo"
	"aletheiaware.com/conveyearthgo/database"
	"aletheiaware.com/conveyearthgo/filesystem"
	"aletheiaware.com/conveyearthgo/handler"
	"github.com/stretchr/testify/assert"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
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
		handler.AttachBestHandler(mux, auth, cm, tmpl)
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

		// Create Conversation
		topic := "FooBar"
		hash, size, err := cm.AddText([]byte("Hello World!"))
		assert.NoError(t, err)
		mime := "text/plain"
		c, m, _, err := cm.NewConversation(acc, topic, []string{hash}, []string{mime}, []int64{size})
		assert.NoError(t, err)

		// Add a Reply
		hash, size, err = cm.AddText([]byte("Hi!"))
		assert.NoError(t, err)
		_, _, err = cm.NewMessage(acc, c.ID, m.ID, []string{hash}, []string{mime}, []int64{size})
		assert.NoError(t, err)

		mux := http.NewServeMux()
		handler.AttachBestHandler(mux, auth, cm, tmpl)
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
	// TODO Many Conversations (Limit/More button)
	// TODO Best of the Day
	// TODO Best of the Week
	// TODO Best of the Month
	// TODO Best of the Year
	// TODO Best of All
}
