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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestRecent(t *testing.T) {
	dir, err := ioutil.TempDir("", "test")
	assert.Nil(t, err)
	fs := filesystem.NewOnDisk(dir)
	defer os.RemoveAll(dir)
	tmpl, err := template.New("recent.go.html").Parse(`{{range .Conversations}}{{.Topic}}{{end}}`)
	assert.Nil(t, err)
	t.Run("Returns 200 With No Conversations", func(t *testing.T) {
		db := database.NewInMemory()
		cm := conveyearthgo.NewContentManager(db, fs)
		mux := http.NewServeMux()
		handler.AttachRecentHandler(mux, cm, tmpl)
		request := httptest.NewRequest(http.MethodGet, "/recent", nil)
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusOK, result.StatusCode)
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Equal(t, "", string(body))
	})
	t.Run("Returns 200 With One Conversation", func(t *testing.T) {
		db := database.NewInMemory()
		ev := authtest.NewEmailVerifier()
		auth := authgo.NewAuthenticator(db, ev)
		acc := authtest.NewTestAccount(t, auth)
		_, err := db.CreateConversation(acc.ID, "FooBar", time.Now())
		assert.Nil(t, err)
		cm := conveyearthgo.NewContentManager(db, fs)
		mux := http.NewServeMux()
		handler.AttachRecentHandler(mux, cm, tmpl)
		request := httptest.NewRequest(http.MethodGet, "/recent", nil)
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusOK, result.StatusCode)
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Equal(t, "FooBar", string(body))
	})
	// TODO Many Conversations (Limit/More button)
}
