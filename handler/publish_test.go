package handler_test

import (
	"aletheiaware.com/authgo"
	"aletheiaware.com/authgo/authtest"
	"aletheiaware.com/conveyearthgo"
	"aletheiaware.com/conveyearthgo/conveytest"
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

func TestPublish(t *testing.T) {
	dir, err := ioutil.TempDir("", "test")
	assert.Nil(t, err)
	fs := filesystem.NewOnDisk(dir)
	defer os.RemoveAll(dir)
	tmpl, err := template.New("publish.go.html").Parse(`{{.Error}}{{with .Account}}{{.Username}}{{end}}`)
	assert.Nil(t, err)
	t.Run("Returns 200 When Signed In", func(t *testing.T) {
		db := database.NewInMemory()
		ev := authtest.NewEmailVerifier()
		auth := authgo.NewAuthenticator(db, ev)
		authtest.NewTestAccount(t, auth)
		token, _ := authtest.SignIn(t, auth)
		am := conveyearthgo.NewAccountManager(db)
		cm := conveyearthgo.NewContentManager(db, fs)
		nm := conveyearthgo.NewNotificationManager(db, conveytest.NewNotificationSender())
		mux := http.NewServeMux()
		handler.AttachPublishHandler(mux, auth, am, cm, nm, tmpl)
		request := httptest.NewRequest(http.MethodGet, "/publish", nil)
		request.AddCookie(auth.NewSignInSessionCookie(token))
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusOK, result.StatusCode)
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Equal(t, authtest.TEST_USERNAME, string(body))
	})
	t.Run("Redirects When Not Signed In", func(t *testing.T) {
		db := database.NewInMemory()
		ev := authtest.NewEmailVerifier()
		auth := authgo.NewAuthenticator(db, ev)
		authtest.NewTestAccount(t, auth)
		am := conveyearthgo.NewAccountManager(db)
		cm := conveyearthgo.NewContentManager(db, fs)
		nm := conveyearthgo.NewNotificationManager(db, conveytest.NewNotificationSender())
		mux := http.NewServeMux()
		handler.AttachPublishHandler(mux, auth, am, cm, nm, tmpl)
		request := httptest.NewRequest(http.MethodGet, "/publish", nil)
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusFound, result.StatusCode)
		u, err := result.Location()
		assert.Nil(t, err)
		assert.Equal(t, "/sign-in", u.String())
	})
	// TODO Publish Success
	// TODO Publish Topic Too Short
	// TODO Publish Topic Too Long
	// TODO Publish Content Too Short
	// TODO Publish Insufficient Balance
}
