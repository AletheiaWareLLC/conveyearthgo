package handler_test

import (
	"aletheiaware.com/conveyearthgo"
	"aletheiaware.com/conveyearthgo/database"
	"aletheiaware.com/conveyearthgo/filesystem"
	"aletheiaware.com/conveyearthgo/handler"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

const notFound = "404 page not found\n"

func TestContent(t *testing.T) {
	dir, err := os.MkdirTemp("", "test")
	assert.Nil(t, err)
	fs := filesystem.NewOnDisk(dir)
	defer os.RemoveAll(dir)
	t.Run("Returns 404 For Empty Hash", func(t *testing.T) {
		db := database.NewInMemory()
		cm := conveyearthgo.NewContentManager(db, fs)
		mux := http.NewServeMux()
		handler.AttachContentHandler(mux, cm, "")
		request := httptest.NewRequest(http.MethodGet, "/content", nil)
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusNotFound, result.StatusCode)
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Equal(t, notFound, string(body))
		request = httptest.NewRequest(http.MethodGet, "/content/", nil)
		response = httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result = response.Result()
		assert.Equal(t, http.StatusNotFound, result.StatusCode)
		body, err = io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Equal(t, notFound, string(body))
	})
	t.Run("Returns 404 When Content Does Not Exist", func(t *testing.T) {
		db := database.NewInMemory()
		cm := conveyearthgo.NewContentManager(db, fs)
		mux := http.NewServeMux()
		handler.AttachContentHandler(mux, cm, "")
		request := httptest.NewRequest(http.MethodGet, "/content/foobar", nil)
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusNotFound, result.StatusCode)
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Equal(t, notFound, string(body))
	})
	t.Run("Returns 200 When Content Exists", func(t *testing.T) {
		db := database.NewInMemory()
		cm := conveyearthgo.NewContentManager(db, fs)
		contents := []byte("this is a test")
		hash, size, err := cm.AddText(contents)
		assert.Nil(t, err)
		mux := http.NewServeMux()
		handler.AttachContentHandler(mux, cm, "")
		request := httptest.NewRequest(http.MethodGet, "/content/"+hash, nil)
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusOK, result.StatusCode)
		assert.Equal(t, conveyearthgo.MIME_TEXT_PLAIN+"; charset=utf-8", response.Header().Get("Content-Type"))
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Equal(t, size, int64(len(body)))
		assert.Equal(t, contents, body)
	})
	t.Run("Content-Type set by URL Query", func(t *testing.T) {
		db := database.NewInMemory()
		cm := conveyearthgo.NewContentManager(db, fs)
		contents := []byte("this is a test")
		hash, size, err := cm.AddText(contents)
		assert.Nil(t, err)
		mux := http.NewServeMux()
		handler.AttachContentHandler(mux, cm, "")
		request := httptest.NewRequest(http.MethodGet, "/content/"+hash+"?mime="+url.QueryEscape(conveyearthgo.MIME_IMAGE_PNG), nil)
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusOK, result.StatusCode)
		assert.Equal(t, conveyearthgo.MIME_IMAGE_PNG, response.Header().Get("Content-Type"))
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Equal(t, size, int64(len(body)))
		assert.Equal(t, contents, body)
	})
}
