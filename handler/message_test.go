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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

func TestMessage(t *testing.T) {
	dir, err := ioutil.TempDir("", "test")
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
		topic := "FooBar"
		content := "Hello World!"
		hash, size, err := cm.AddText([]byte(content))
		assert.Nil(t, err)
		mime := "text/plain"
		cost := strconv.FormatInt(size, 10)
		yield := "0"
		_, m, err := cm.NewConversation(acc, topic, []string{hash}, []string{mime}, []int64{size})
		assert.Nil(t, err)
		mux := http.NewServeMux()
		handler.AttachMessageHandler(mux, auth, cm, tmpl)
		request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/message?id=%d", m.ID), nil)
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusOK, result.StatusCode)
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Equal(t, authtest.TEST_USERNAME+cost+yield+"<p>"+content+"</p>", string(body))
	})
	t.Run("Returns 200 When Signed In And Message Exists", func(t *testing.T) {
		db := database.NewInMemory()
		ev := authtest.NewEmailVerifier()
		auth := authgo.NewAuthenticator(db, ev)
		acc := authtest.NewTestAccount(t, auth)
		token, _ := authtest.SignIn(t, auth)
		cm := conveyearthgo.NewContentManager(db, fs)
		topic := "FooBar"
		content := "Hello World!"
		hash, size, err := cm.AddText([]byte(content))
		assert.Nil(t, err)
		mime := "text/plain"
		cost := strconv.FormatInt(size, 10)
		yield := "0"
		_, m, err := cm.NewConversation(acc, topic, []string{hash}, []string{mime}, []int64{size})
		assert.Nil(t, err)
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
		assert.Equal(t, authtest.TEST_USERNAME+authtest.TEST_USERNAME+cost+yield+"<p>"+content+"</p>", string(body))
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
