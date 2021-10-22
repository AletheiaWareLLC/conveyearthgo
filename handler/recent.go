package handler

import (
	"aletheiaware.com/conveyearthgo"
	"aletheiaware.com/netgo"
	"aletheiaware.com/netgo/handler"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func AttachRecentHandler(m *http.ServeMux, cm conveyearthgo.ContentManager, ts *template.Template) {
	m.Handle("/recent", handler.Log(Recent(cm, ts)))
}

func Recent(cm conveyearthgo.ContentManager, ts *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Live          bool
			Conversations []*conveyearthgo.Conversation
			Limit         int64
		}{
			Live: netgo.IsLive(),
		}
		limit := int64(8)
		// TODO limit may only be greater than 8 if user is signed in
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
