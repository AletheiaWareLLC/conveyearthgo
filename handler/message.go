package handler

import (
	"aletheiaware.com/authgo"
	"aletheiaware.com/conveyearthgo"
	"aletheiaware.com/netgo/handler"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func AttachMessageHandler(m *http.ServeMux, a authgo.Authenticator, cm conveyearthgo.ContentManager, ts *template.Template) {
	m.Handle("/message", handler.Log(Message(a, cm, ts)))
}

func Message(a authgo.Authenticator, cm conveyearthgo.ContentManager, ts *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		account := a.CurrentAccount(w, r)
		var id int64
		if m := strings.TrimSpace(r.FormValue("id")); m != "" {
			if i, err := strconv.ParseInt(m, 10, 64); err != nil {
				log.Println(err)
			} else {
				id = int64(i)
			}
		}
		if id == 0 {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		m, err := cm.LookupMessage(id)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		var content template.HTML
		if err := cm.LookupFiles(id, func(f *conveyearthgo.File) error {
			c, err := cm.ToHTML(f.Hash, f.Mime)
			if err != nil {
				return err
			}
			content += c
			return nil
		}); err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		data := struct {
			Account      *authgo.Account
			Conversation int64
			Message      int64
			Parent       int64
			User         string
			Cost         int64
			Yield        int64
			Content      template.HTML
			Created      time.Time
		}{
			Account:      account,
			Conversation: m.Conversation,
			Message:      id,
			Parent:       m.Parent,
			User:         m.User,
			Cost:         m.Cost,
			Yield:        m.Yield,
			Content:      content,
			Created:      m.Created,
		}
		if err := ts.ExecuteTemplate(w, "message.go.html", data); err != nil {
			log.Println(err)
			return
		}
	})
}
