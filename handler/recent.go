package handler

import (
	"aletheiaware.com/authgo"
	"aletheiaware.com/conveyearthgo"
	"aletheiaware.com/netgo"
	"aletheiaware.com/netgo/handler"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func AttachRecentHandler(m *http.ServeMux, a authgo.Authenticator, cm conveyearthgo.ContentManager, ts *template.Template, count, maximum int64) {
	m.Handle("/recent", handler.Log(Recent(a, cm, ts, count, maximum)))
}

func Recent(a authgo.Authenticator, cm conveyearthgo.ContentManager, ts *template.Template, count, maximum int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Live          bool
			Account       *authgo.Account
			Conversations []*conveyearthgo.Conversation
			Limit         int64
		}{
			Live: netgo.IsLive(),
		}
		limit := count
		if l := strings.TrimSpace(r.FormValue("limit")); l != "" {
			if i, err := strconv.ParseInt(l, 10, 64); err != nil {
				log.Println(err)
			} else {
				limit = int64(i)
			}
		}
		data.Limit = limit * 2
		data.Account = a.CurrentAccount(w, r)
		if data.Account == nil {
			if limit > maximum {
				limit = maximum
			}
		}
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
