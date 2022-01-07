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
	"strconv"
	"testing"
)

func TestMessage(t *testing.T) {
	dir, err := os.MkdirTemp("", "test")
	assert.Nil(t, err)
	fs := filesystem.NewOnDisk(dir)
	defer os.RemoveAll(dir)
	tmpl, err := template.New("message.go.html").Parse(`{{with .Account}}{{.Username}}{{end}}{{.Author.Username}}{{.Cost}}{{.Yield}}{{.Content}}`)
	assert.Nil(t, err)
	t.Run("Returns 200 When Not Signed In And Message Exists", func(t *testing.T) {
		db := database.NewInMemory()
		ev := authtest.NewEmailVerifier()
		auth := authgo.NewAuthenticator(db, ev)
		acc := authtest.NewTestAccount(t, auth)
		cm := conveyearthgo.NewContentManager(db, fs)
		c, m, _ := conveytest.NewConversation(t, cm, acc)
		mux := http.NewServeMux()
		handler.AttachMessageHandler(mux, auth, cm, tmpl)
		request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/message?id=%d", m.ID), nil)
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusOK, result.StatusCode)
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Equal(t, authtest.TEST_USERNAME+strconv.FormatInt(c.Cost, 10)+"0"+`<p class="ucc">`+conveytest.TEST_CONTENT+`</p>`, string(body))
	})
	t.Run("Returns 200 When Signed In And Message Exists", func(t *testing.T) {
		db := database.NewInMemory()
		ev := authtest.NewEmailVerifier()
		auth := authgo.NewAuthenticator(db, ev)
		acc := authtest.NewTestAccount(t, auth)
		token, _ := authtest.SignIn(t, auth)
		cm := conveyearthgo.NewContentManager(db, fs)
		c, m, _ := conveytest.NewConversation(t, cm, acc)
		mux := http.NewServeMux()
		handler.AttachMessageHandler(mux, auth, cm, tmpl)
		request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/message?id=%d", m.ID), nil)
		request.AddCookie(auth.NewSignInSessionCookie(token))
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusOK, result.StatusCode)
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Equal(t, authtest.TEST_USERNAME+authtest.TEST_USERNAME+strconv.FormatInt(c.Cost, 10)+"0"+`<p class="ucc">`+conveytest.TEST_CONTENT+`</p>`, string(body))
	})
	t.Run("Returns 404 When Message Does Not Exist", func(t *testing.T) {
		db := database.NewInMemory()
		ev := authtest.NewEmailVerifier()
		auth := authgo.NewAuthenticator(db, ev)
		authtest.NewTestAccount(t, auth)
		token, _ := authtest.SignIn(t, auth)
		cm := conveyearthgo.NewContentManager(db, fs)
		mux := http.NewServeMux()
		handler.AttachMessageHandler(mux, auth, cm, tmpl)
		request := httptest.NewRequest(http.MethodGet, "/message", nil)
		request.AddCookie(auth.NewSignInSessionCookie(token))
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusNotFound, result.StatusCode)
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusText(http.StatusNotFound)+"\n", string(body))
	})
	// TODO Message with Reply shows Yield
}
