package handler

import (
	"aletheiaware.com/authgo"
	"aletheiaware.com/authgo/redirect"
	"aletheiaware.com/conveyearthgo"
	"aletheiaware.com/netgo"
	"aletheiaware.com/netgo/handler"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func AttachRecentHandler(m *http.ServeMux, a authgo.Authenticator, cm conveyearthgo.ContentManager, ts *template.Template) {
	m.Handle("/recent", handler.Log(Recent(a, cm, ts)))
}

func Recent(a authgo.Authenticator, cm conveyearthgo.ContentManager, ts *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		account := a.CurrentAccount(w, r)
		if account == nil {
			redirect.SignIn(w, r)
			return
		}
		type ConversationData struct {
			ID      int64
			Topic   string
			User    string
			Cost    int64
			Yield   int64
			Created time.Time
		}
		data := struct {
			Live          bool
			Account       *authgo.Account
			Conversations []*ConversationData
			Limit         int64
		}{
			Live:    netgo.IsLive(),
			Account: account,
		}
		limit := int64(8)
		if l := strings.TrimSpace(r.FormValue("limit")); l != "" {
			if i, err := strconv.ParseInt(l, 10, 64); err != nil {
				log.Println(err)
			} else {
				limit = int64(i)
			}
		}
		data.Limit = limit * 2
		if err := cm.LookupRecentConversations(func(c *conveyearthgo.Conversation, cost, yield int64) error {
			data.Conversations = append(data.Conversations, &ConversationData{
				ID:      c.ID,
				Topic:   c.Topic,
				User:    c.User,
				Cost:    cost,
				Yield:   yield,
				Created: c.Created,
			})
			return nil
		}, limit); err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		if err := ts.ExecuteTemplate(w, "recent.go.html", data); err != nil {
			log.Println(err)
			return
		}
	})
}
