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

func TestDemo(t *testing.T) {
	tmpl, err := template.New("demo.go.html").Parse(`{{.Live}}`)
	assert.Nil(t, err)
	mux := http.NewServeMux()
	handler.AttachDemoHandler(mux, tmpl)
	request := httptest.NewRequest(http.MethodGet, "/demo", nil)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	result := response.Result()
	assert.Equal(t, http.StatusOK, result.StatusCode)
	body, err := io.ReadAll(result.Body)
	assert.Nil(t, err)
	assert.Equal(t, "false", string(body))
}
