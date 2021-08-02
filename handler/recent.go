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
		data := struct {
			Live          bool
			Account       *authgo.Account
			Conversations []*conveyearthgo.Conversation
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
		if err := cm.LookupRecentConversations(func(c *conveyearthgo.Conversation) error {
			data.Conversations = append(data.Conversations, c)
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
