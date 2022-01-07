package handler_test

import (
	"aletheiaware.com/conveyearthgo/handler"
	"github.com/stretchr/testify/assert"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDigest(t *testing.T) {
	dir, err := os.MkdirTemp("", "test")
	assert.Nil(t, err)
	defer os.RemoveAll(dir)
	_, err = os.Create(filepath.Join(dir, "Convey-Digest-2006-01.epub"))
	assert.NoError(t, err)
	_, err = os.Create(filepath.Join(dir, "Convey-Digest-2006-02.epub"))
	assert.NoError(t, err)

	tmpl, err := template.New("digest.go.html").Parse(`{{range .Editions}}{{.}}{{end}}`)
	assert.Nil(t, err)
	tmpl, err = tmpl.New("digest-viewer.go.html").Parse(`{{.Edition}}`)
	assert.Nil(t, err)

	mux := http.NewServeMux()
	handler.AttachDigestHandler(mux, tmpl, dir, "")

	request := httptest.NewRequest(http.MethodGet, "/digest", nil)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	result := response.Result()
	assert.Equal(t, http.StatusOK, result.StatusCode)
	body, err := io.ReadAll(result.Body)
	assert.Nil(t, err)
	assert.Equal(t, "2006-022006-01", string(body))

	request = httptest.NewRequest(http.MethodGet, "/digest?edition=2006-01", nil)
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	result = response.Result()
	assert.Equal(t, http.StatusOK, result.StatusCode)
	body, err = io.ReadAll(result.Body)
	assert.Nil(t, err)
	assert.Equal(t, "2006-01", string(body))
}
