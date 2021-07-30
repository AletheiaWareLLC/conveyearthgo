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
)

func TestBest(t *testing.T) {
	dir, err := ioutil.TempDir("", "test")
	assert.Nil(t, err)
	fs := filesystem.NewOnDisk(dir)
	defer os.RemoveAll(dir)
	tmpl, err := template.New("best.go.html").Parse(`{{with .Account}}{{.Username}}{{end}}`)
	assert.Nil(t, err)
	t.Run("Returns 200 When Signed In", func(t *testing.T) {
		db := database.NewInMemory()
		ev := authtest.NewEmailVerifier()
		auth := authgo.NewAuthenticator(db, ev)
		authtest.NewTestAccount(t, auth)
		token, _ := authtest.SignIn(t, auth)
		cm := conveyearthgo.NewContentManager(db, fs)
		mux := http.NewServeMux()
		handler.AttachBestHandler(mux, auth, cm, tmpl)
		request := httptest.NewRequest(http.MethodGet, "/best", nil)
		request.AddCookie(authgo.NewSignInSessionCookie(token))
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, result.StatusCode)
		assert.Equal(t, authtest.TEST_USERNAME, string(body))
	})
	t.Run("Redirects When Not Signed In", func(t *testing.T) {
		db := database.NewInMemory()
		ev := authtest.NewEmailVerifier()
		auth := authgo.NewAuthenticator(db, ev)
		authtest.NewTestAccount(t, auth)
		cm := conveyearthgo.NewContentManager(db, fs)
		mux := http.NewServeMux()
		handler.AttachBestHandler(mux, auth, cm, tmpl)
		request := httptest.NewRequest(http.MethodGet, "/best", nil)
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusFound, result.StatusCode)
		u, err := result.Location()
		assert.Nil(t, err)
		assert.Equal(t, "/sign-in", u.String())
	})
	// TODO No Conversations
	// TODO Many Conversations (Limit/More button)
	// TODO Best of the Day
	// TODO Best of the Week
	// TODO Best of the Year
	// TODO Best of All
}
