package handler_test

import (
	"aletheiaware.com/conveyearthgo/handler"
	"github.com/stretchr/testify/assert"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAbout(t *testing.T) {
	tmpl, err := template.New("about.go.html").Parse(`{{.Live}}`)
	assert.Nil(t, err)
	mux := http.NewServeMux()
	handler.AttachAboutHandler(mux, tmpl)
	request := httptest.NewRequest(http.MethodGet, "/about", nil)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	result := response.Result()
	assert.Equal(t, http.StatusOK, result.StatusCode)
	body, err := io.ReadAll(result.Body)
	assert.Nil(t, err)
	assert.Equal(t, "false", string(body))
}
