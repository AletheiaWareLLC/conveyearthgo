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
	"path/filepath"
	"testing"
)

func TestIndex(t *testing.T) {
	dir, err := os.MkdirTemp("", "test")
	assert.Nil(t, err)
	fs := filesystem.NewOnDisk(dir)
	defer os.RemoveAll(dir)
	tmpl, err := template.New("index.go.html").Parse(`{{with .Account}}{{.Username}}{{end}}{{range .Best}}{{.Topic}}{{end}}{{range .Recent}}{{.Topic}}{{end}}{{range .Editions}}{{.}}{{end}}`)
	assert.Nil(t, err)
	t.Run("Returns 200 When Signed In", func(t *testing.T) {
		db := database.NewInMemory()
		ev := authtest.NewEmailVerifier()
		auth := authgo.NewAuthenticator(db, ev)
		authtest.NewTestAccount(t, auth)
		token, _ := authtest.SignIn(t, auth)
		cm := conveyearthgo.NewContentManager(db, fs)
		mux := http.NewServeMux()
		handler.AttachIndexHandler(mux, auth, cm, tmpl, dir)
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.AddCookie(auth.NewSignInSessionCookie(token))
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusOK, result.StatusCode)
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Equal(t, authtest.TEST_USERNAME, string(body))
	})
	t.Run("Returns 200 When Not Signed In", func(t *testing.T) {
		db := database.NewInMemory()
		ev := authtest.NewEmailVerifier()
		auth := authgo.NewAuthenticator(db, ev)
		cm := conveyearthgo.NewContentManager(db, fs)
		mux := http.NewServeMux()
		handler.AttachIndexHandler(mux, auth, cm, tmpl, dir)
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusOK, result.StatusCode)
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Empty(t, string(body))
	})
	t.Run("With Best And Recent", func(t *testing.T) {
		db := database.NewInMemory()
		ev := authtest.NewEmailVerifier()
		auth := authgo.NewAuthenticator(db, ev)
		acc := authtest.NewTestAccount(t, auth)
		token, _ := authtest.SignIn(t, auth)
		cm := conveyearthgo.NewContentManager(db, fs)

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
		handler.AttachIndexHandler(mux, auth, cm, tmpl, dir)
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.AddCookie(auth.NewSignInSessionCookie(token))
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusOK, result.StatusCode)
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Equal(t, authtest.TEST_USERNAME+"FooBarFooBar", string(body))
	})
	t.Run("With Edition", func(t *testing.T) {
		db := database.NewInMemory()
		ev := authtest.NewEmailVerifier()
		auth := authgo.NewAuthenticator(db, ev)
		authtest.NewTestAccount(t, auth)
		token, _ := authtest.SignIn(t, auth)
		cm := conveyearthgo.NewContentManager(db, fs)

		// Create Edition
		file := filepath.Join(dir, "Convey-Digest-2006-01.epub")
		_, err = os.Create(file)
		assert.NoError(t, err)
		defer os.Remove(file)

		mux := http.NewServeMux()
		handler.AttachIndexHandler(mux, auth, cm, tmpl, dir)
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.AddCookie(auth.NewSignInSessionCookie(token))
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusOK, result.StatusCode)
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Equal(t, authtest.TEST_USERNAME+"2006-01", string(body))
	})
}
