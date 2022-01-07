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

func TestConversation(t *testing.T) {
	dir, err := os.MkdirTemp("", "test")
	assert.Nil(t, err)
	fs := filesystem.NewOnDisk(dir)
	defer os.RemoveAll(dir)
	tmpl, err := template.New("conversation.go.html").Parse(`{{with .Account}}{{.Username}}{{end}}{{.Topic}}{{.Author.Username}}{{.Cost}}{{.Yield}}{{.Content}}{{range .Gifts}}{{.Amount}}{{end}}{{range .Replies}}{{.Cost}}{{end}}`)
	assert.Nil(t, err)
	t.Run("Returns 200 When Not Signed In And Conversation Exists", func(t *testing.T) {
		db := database.NewInMemory()
		ev := authtest.NewEmailVerifier()
		auth := authgo.NewAuthenticator(db, ev)
		acc := authtest.NewTestAccount(t, auth)
		cm := conveyearthgo.NewContentManager(db, fs)
		c, _, _ := conveytest.NewConversation(t, cm, acc)
		mux := http.NewServeMux()
		handler.AttachConversationHandler(mux, auth, cm, tmpl)
		request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/conversation?id=%d", c.ID), nil)
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusOK, result.StatusCode)
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Equal(t, conveytest.TEST_TOPIC+authtest.TEST_USERNAME+strconv.FormatInt(c.Cost, 10)+"0"+`<p class="ucc">`+conveytest.TEST_CONTENT+`</p>`, string(body))
	})
	t.Run("Returns 200 When Signed In And Conversation Exists", func(t *testing.T) {
		db := database.NewInMemory()
		ev := authtest.NewEmailVerifier()
		auth := authgo.NewAuthenticator(db, ev)
		acc := authtest.NewTestAccount(t, auth)
		token, _ := authtest.SignIn(t, auth)
		cm := conveyearthgo.NewContentManager(db, fs)
		c, _, _ := conveytest.NewConversation(t, cm, acc)
		mux := http.NewServeMux()
		handler.AttachConversationHandler(mux, auth, cm, tmpl)
		request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/conversation?id=%d", c.ID), nil)
		request.AddCookie(auth.NewSignInSessionCookie(token))
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusOK, result.StatusCode)
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Equal(t, authtest.TEST_USERNAME+conveytest.TEST_TOPIC+authtest.TEST_USERNAME+strconv.FormatInt(c.Cost, 10)+"0"+`<p class="ucc">`+conveytest.TEST_CONTENT+`</p>`, string(body))
	})
	t.Run("Returns 404 When Conversation Does Not Exist", func(t *testing.T) {
		db := database.NewInMemory()
		ev := authtest.NewEmailVerifier()
		auth := authgo.NewAuthenticator(db, ev)
		authtest.NewTestAccount(t, auth)
		token, _ := authtest.SignIn(t, auth)
		cm := conveyearthgo.NewContentManager(db, fs)
		mux := http.NewServeMux()
		handler.AttachConversationHandler(mux, auth, cm, tmpl)
		request := httptest.NewRequest(http.MethodGet, "/conversation", nil)
		request.AddCookie(auth.NewSignInSessionCookie(token))
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusNotFound, result.StatusCode)
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusText(http.StatusNotFound)+"\n", string(body))
	})
	t.Run("Conversation with Reply", func(t *testing.T) {
		db := database.NewInMemory()
		ev := authtest.NewEmailVerifier()
		auth := authgo.NewAuthenticator(db, ev)
		acc := authtest.NewTestAccount(t, auth)
		token, _ := authtest.SignIn(t, auth)
		cm := conveyearthgo.NewContentManager(db, fs)
		c, m, _ := conveytest.NewConversation(t, cm, acc)
		conveytest.NewReply(t, cm, acc, c, m)
		mux := http.NewServeMux()
		handler.AttachConversationHandler(mux, auth, cm, tmpl)
		request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/conversation?id=%d", c.ID), nil)
		request.AddCookie(auth.NewSignInSessionCookie(token))
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusOK, result.StatusCode)
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Equal(t, authtest.TEST_USERNAME+conveytest.TEST_TOPIC+authtest.TEST_USERNAME+strconv.FormatInt(c.Cost, 10)+"1"+`<p class="ucc">`+conveytest.TEST_CONTENT+`</p>3`, string(body))
	})
	t.Run("Conversation with Gift", func(t *testing.T) {
		db := database.NewInMemory()
		ev := authtest.NewEmailVerifier()
		auth := authgo.NewAuthenticator(db, ev)
		acc := authtest.NewTestAccount(t, auth)
		token, _ := authtest.SignIn(t, auth)
		cm := conveyearthgo.NewContentManager(db, fs)
		c, m, _ := conveytest.NewConversation(t, cm, acc)
		conveytest.NewGift(t, cm, acc, c, m)
		mux := http.NewServeMux()
		handler.AttachConversationHandler(mux, auth, cm, tmpl)
		request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/conversation?id=%d", c.ID), nil)
		request.AddCookie(auth.NewSignInSessionCookie(token))
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusOK, result.StatusCode)
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Equal(t, authtest.TEST_USERNAME+conveytest.TEST_TOPIC+authtest.TEST_USERNAME+strconv.FormatInt(c.Cost, 10)+"0"+`<p class="ucc">`+conveytest.TEST_CONTENT+`</p>100`, string(body))
	})
}
