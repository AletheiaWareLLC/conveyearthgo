package handler_test

import (
	"aletheiaware.com/authgo"
	"aletheiaware.com/authgo/authtest"
	"aletheiaware.com/conveyearthgo/handler"
	"github.com/stretchr/testify/assert"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDemo(t *testing.T) {
	tmpl, err := template.New("demo.go.html").Parse(``)
	assert.Nil(t, err)
	t.Run("Returns 200 When Signed In", func(t *testing.T) {
		auth := authtest.NewAuthenticator(t)
		authtest.NewTestAccount(t, auth)
		token, _ := authtest.SignIn(t, auth)
		mux := http.NewServeMux()
		handler.AttachDemoHandler(mux, auth, tmpl)
		request := httptest.NewRequest(http.MethodGet, "/demo", nil)
		request.AddCookie(authgo.NewSignInSessionCookie(token))
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusOK, result.StatusCode)
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Empty(t, string(body))
	})
	t.Run("Returns 200 When Not Signed In", func(t *testing.T) {
		auth := authtest.NewAuthenticator(t)
		mux := http.NewServeMux()
		handler.AttachDemoHandler(mux, auth, tmpl)
		request := httptest.NewRequest(http.MethodGet, "/demo", nil)
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusOK, result.StatusCode)
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Empty(t, string(body))
	})
}
