package handler_test

import (
	"aletheiaware.com/authgo"
	"aletheiaware.com/authgo/authtest"
	"aletheiaware.com/conveyearthgo"
	"aletheiaware.com/conveyearthgo/conveytest"
	"aletheiaware.com/conveyearthgo/database"
	"aletheiaware.com/conveyearthgo/filesystem"
	"aletheiaware.com/conveyearthgo/handler"
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestReply(t *testing.T) {
	dir, err := os.MkdirTemp("", "test")
	assert.Nil(t, err)
	fs := filesystem.NewOnDisk(dir)
	defer os.RemoveAll(dir)
	tmpl, err := template.New("reply.go.html").Parse(`{{.Error}}{{with .Account}}{{.Username}}{{end}}`)
	assert.Nil(t, err)
	t.Run("Returns 200 When Signed In And Conversation And Message Exist", func(t *testing.T) {
		db := database.NewInMemory()
		ev := authtest.NewEmailVerifier()
		auth := authgo.NewAuthenticator(db, ev)
		acc := authtest.NewTestAccount(t, auth)
		token, _ := authtest.SignIn(t, auth)
		am := conveyearthgo.NewAccountManager(db)
		cm := conveyearthgo.NewContentManager(db, fs)
		nm := conveyearthgo.NewNotificationManager(db, conveytest.NewNotificationSender())
		c, m, _ := conveytest.NewConversation(t, cm, acc)
		mux := http.NewServeMux()
		handler.AttachReplyHandler(mux, auth, am, cm, nm, tmpl)
		request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/reply?conversation=%d&message=%d", c.ID, m.ID), nil)
		request.AddCookie(auth.NewSignInSessionCookie(token))
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusOK, result.StatusCode)
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Equal(t, authtest.TEST_USERNAME, string(body))
	})
	t.Run("Returns 404 When Conversation Does Not Exist", func(t *testing.T) {
		db := database.NewInMemory()
		ev := authtest.NewEmailVerifier()
		auth := authgo.NewAuthenticator(db, ev)
		acc := authtest.NewTestAccount(t, auth)
		token, _ := authtest.SignIn(t, auth)
		am := conveyearthgo.NewAccountManager(db)
		cm := conveyearthgo.NewContentManager(db, fs)
		c, m, _ := conveytest.NewConversation(t, cm, acc)
		nm := conveyearthgo.NewNotificationManager(db, conveytest.NewNotificationSender())
		mux := http.NewServeMux()
		handler.AttachReplyHandler(mux, auth, am, cm, nm, tmpl)
		// Get
		request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/reply?conversation=%d&message=%d", c.ID+1, m.ID), nil)
		request.AddCookie(auth.NewSignInSessionCookie(token))
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusNotFound, result.StatusCode)
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusText(http.StatusNotFound)+"\n", string(body))
		// Post
		var buffer bytes.Buffer
		writer := multipart.NewWriter(&buffer)
		_ = writer.WriteField("conversation", strconv.FormatInt(c.ID+1, 10))
		_ = writer.WriteField("message", strconv.FormatInt(m.ID, 10))
		_ = writer.WriteField("reply", conveytest.TEST_REPLY)
		assert.NoError(t, writer.Close())
		request = httptest.NewRequest(http.MethodPost, "/reply", &buffer)
		request.Header.Set("Content-Type", writer.FormDataContentType())
		request.AddCookie(auth.NewSignInSessionCookie(token))
		response = httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result = response.Result()
		assert.Equal(t, http.StatusNotFound, result.StatusCode)
		body, err = io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusText(http.StatusNotFound)+"\n", string(body))
	})
	t.Run("Returns 404 When Message Does Not Exist", func(t *testing.T) {
		db := database.NewInMemory()
		ev := authtest.NewEmailVerifier()
		auth := authgo.NewAuthenticator(db, ev)
		acc := authtest.NewTestAccount(t, auth)
		token, _ := authtest.SignIn(t, auth)
		am := conveyearthgo.NewAccountManager(db)
		cm := conveyearthgo.NewContentManager(db, fs)
		c, m, _ := conveytest.NewConversation(t, cm, acc)
		nm := conveyearthgo.NewNotificationManager(db, conveytest.NewNotificationSender())
		mux := http.NewServeMux()
		handler.AttachReplyHandler(mux, auth, am, cm, nm, tmpl)
		// Get
		request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/reply?conversation=%d&message=%d", c.ID, m.ID+1), nil)
		request.AddCookie(auth.NewSignInSessionCookie(token))
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusNotFound, result.StatusCode)
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusText(http.StatusNotFound)+"\n", string(body))
		// Post
		var buffer bytes.Buffer
		writer := multipart.NewWriter(&buffer)
		_ = writer.WriteField("conversation", strconv.FormatInt(c.ID, 10))
		_ = writer.WriteField("message", strconv.FormatInt(m.ID+1, 10))
		_ = writer.WriteField("reply", conveytest.TEST_REPLY)
		assert.NoError(t, writer.Close())
		request = httptest.NewRequest(http.MethodPost, "/reply", &buffer)
		request.Header.Set("Content-Type", writer.FormDataContentType())
		request.AddCookie(auth.NewSignInSessionCookie(token))
		response = httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result = response.Result()
		assert.Equal(t, http.StatusNotFound, result.StatusCode)
		body, err = io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusText(http.StatusNotFound)+"\n", string(body))
	})
	t.Run("Redirects When Not Signed In", func(t *testing.T) {
		db := database.NewInMemory()
		ev := authtest.NewEmailVerifier()
		auth := authgo.NewAuthenticator(db, ev)
		am := conveyearthgo.NewAccountManager(db)
		cm := conveyearthgo.NewContentManager(db, fs)
		nm := conveyearthgo.NewNotificationManager(db, conveytest.NewNotificationSender())
		mux := http.NewServeMux()
		handler.AttachReplyHandler(mux, auth, am, cm, nm, tmpl)
		request := httptest.NewRequest(http.MethodGet, "/reply?conversation=10&message=10", nil)
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusFound, result.StatusCode)
		u, err := result.Location()
		assert.Nil(t, err)
		assert.Equal(t, "/sign-in?next=%2Freply%3Fconversation%3D10%26message%3D10", u.String())
	})
	t.Run("Too Short", func(t *testing.T) {
		db := database.NewInMemory()
		ev := authtest.NewEmailVerifier()
		auth := authgo.NewAuthenticator(db, ev)
		acc := authtest.NewTestAccount(t, auth)
		token, _ := authtest.SignIn(t, auth)
		am := conveyearthgo.NewAccountManager(db)
		cm := conveyearthgo.NewContentManager(db, fs)
		c, m, _ := conveytest.NewConversation(t, cm, acc)
		nm := conveyearthgo.NewNotificationManager(db, conveytest.NewNotificationSender())
		mux := http.NewServeMux()
		handler.AttachReplyHandler(mux, auth, am, cm, nm, tmpl)
		var buffer bytes.Buffer
		writer := multipart.NewWriter(&buffer)
		_ = writer.WriteField("conversation", strconv.FormatInt(c.ID, 10))
		_ = writer.WriteField("message", strconv.FormatInt(m.ID, 10))
		_ = writer.WriteField("reply", strings.Repeat("x", conveyearthgo.MINIMUM_CONTENT_LENGTH-1))
		assert.NoError(t, writer.Close())
		request := httptest.NewRequest(http.MethodPost, "/reply", &buffer)
		request.Header.Set("Content-Type", writer.FormDataContentType())
		request.AddCookie(auth.NewSignInSessionCookie(token))
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusOK, result.StatusCode)
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Equal(t, conveyearthgo.ErrContentTooShort.Error()+authtest.TEST_USERNAME, string(body))
	})
	// TODO Content Type Not Multipart Form
	// TODO Attachment Invalid Mime
	t.Run("Insufficient Balance", func(t *testing.T) {
		db := database.NewInMemory()
		ev := authtest.NewEmailVerifier()
		auth := authgo.NewAuthenticator(db, ev)
		acc := authtest.NewTestAccount(t, auth)
		token, _ := authtest.SignIn(t, auth)
		am := conveyearthgo.NewAccountManager(db)
		cm := conveyearthgo.NewContentManager(db, fs)
		c, m, _ := conveytest.NewConversation(t, cm, acc)
		nm := conveyearthgo.NewNotificationManager(db, conveytest.NewNotificationSender())
		mux := http.NewServeMux()
		handler.AttachReplyHandler(mux, auth, am, cm, nm, tmpl)
		var buffer bytes.Buffer
		writer := multipart.NewWriter(&buffer)
		_ = writer.WriteField("conversation", strconv.FormatInt(c.ID, 10))
		_ = writer.WriteField("message", strconv.FormatInt(m.ID, 10))
		_ = writer.WriteField("reply", conveytest.TEST_REPLY)
		assert.NoError(t, writer.Close())
		request := httptest.NewRequest(http.MethodPost, "/reply", &buffer)
		request.Header.Set("Content-Type", writer.FormDataContentType())
		request.AddCookie(auth.NewSignInSessionCookie(token))
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusOK, result.StatusCode)
		body, err := io.ReadAll(result.Body)
		assert.Nil(t, err)
		assert.Equal(t, conveyearthgo.ErrInsufficientBalance.Error()+authtest.TEST_USERNAME, string(body))
	})
	t.Run("Success", func(t *testing.T) {
		db := database.NewInMemory()
		ev := authtest.NewEmailVerifier()
		auth := authgo.NewAuthenticator(db, ev)
		acc := authtest.NewTestAccount(t, auth)
		token, _ := authtest.SignIn(t, auth)
		am := conveyearthgo.NewAccountManager(db)
		conveytest.NewPurchase(t, am, acc)
		cm := conveyearthgo.NewContentManager(db, fs)
		c, m, _ := conveytest.NewConversation(t, cm, acc)
		nm := conveyearthgo.NewNotificationManager(db, conveytest.NewNotificationSender())
		mux := http.NewServeMux()
		handler.AttachReplyHandler(mux, auth, am, cm, nm, tmpl)
		var buffer bytes.Buffer
		writer := multipart.NewWriter(&buffer)
		_ = writer.WriteField("conversation", strconv.FormatInt(c.ID, 10))
		_ = writer.WriteField("message", strconv.FormatInt(m.ID, 10))
		_ = writer.WriteField("reply", conveytest.TEST_REPLY)
		assert.NoError(t, writer.Close())
		request := httptest.NewRequest(http.MethodPost, "/reply", &buffer)
		request.Header.Set("Content-Type", writer.FormDataContentType())
		request.AddCookie(auth.NewSignInSessionCookie(token))
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		result := response.Result()
		assert.Equal(t, http.StatusFound, result.StatusCode)
		u, err := result.Location()
		assert.Nil(t, err)
		assert.True(t, strings.HasPrefix(u.String(), fmt.Sprintf("/conversation?id=%d#message", c.ID)))
	})
	// TODO Success Reply Notification
	// TODO Success Mention Notification
	// TODO Success Attachment
	// TODO Success Content Carriage Return Removed
}
