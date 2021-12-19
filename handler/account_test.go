package handler_test

import (
	"aletheiaware.com/authgo"
	"aletheiaware.com/authgo/authtest"
	"aletheiaware.com/conveyearthgo"
	"aletheiaware.com/conveyearthgo/conveytest"
	"aletheiaware.com/conveyearthgo/database"
	"aletheiaware.com/conveyearthgo/handler"
	"github.com/stretchr/testify/assert"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAccount(t *testing.T) {
	tmpl, err := template.New("account.go.html").Parse(`{{.Error}}{{with .Account}}{{.Username}}{{end}}`)
	assert.Nil(t, err)
	t.Run("Returns 200 When Signed In", func(t *testing.T) {
		db := database.NewInMemory()
		ev := authtest.NewEmailVerifier()
		auth := authgo.NewAuthenticator(db, ev)
		authtest.NewTestAccount(t, auth)
		token, _ := authtest.SignIn(t, auth)
		am := conveyearthgo.NewAccountManager(db)
		nm := conveyearthgo.NewNotificationManager(db, conveytest.NewNotificationSender())
		mux := http.NewServeMux()
		handler.AttachAccountHandler(mux, auth, am, nm, tmpl)
		request := httptest.NewRequest(http.MethodGet, "/account", nil)
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
		nm := conveyearthgo.NewNotificationManager(db, conveytest.NewNotificationSender())
		mux := http.NewServeMux()
		handler.AttachAccountHandler(mux, auth, am, nm, tmpl)
		request := httptest.NewRequest(http.MethodGet, "/account", nil)
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusFound, result.StatusCode)
		u, err := result.Location()
		assert.Nil(t, err)
		assert.Equal(t, "/sign-in?next=%2Faccount", u.String())
	})
}
