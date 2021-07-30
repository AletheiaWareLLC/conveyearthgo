package handler

import (
	"aletheiaware.com/authgo"
	"aletheiaware.com/conveyearthgo"
	"aletheiaware.com/netgo/handler"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

func AttachConversationHandler(m *http.ServeMux, a authgo.Authenticator, cm conveyearthgo.ContentManager, ts *template.Template) {
	m.Handle("/conversation", handler.Log(Conversation(a, cm, ts)))
}

func Conversation(a authgo.Authenticator, cm conveyearthgo.ContentManager, ts *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		account := a.CurrentAccount(w, r)
		var id int64
		if c := strings.TrimSpace(r.FormValue("id")); c != "" {
			if i, err := strconv.ParseInt(c, 10, 64); err != nil {
				log.Println(err)
			} else {
				id = int64(i)
			}
		}
		if id == 0 {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		type MessageData struct {
			Conversation int64
			Message      int64
			Parent       int64
			User         string
			Cost         int64
			Yield        int64
			Content      template.HTML
			Replies      []*MessageData
			Created      time.Time
		}
		data := struct {
			MessageData
			Account *authgo.Account
			Topic   string
		}{
			Account: account,
		}
		data.Conversation = id
		c, err := cm.LookupConversation(id)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		data.Topic = c.Topic
		data.Created = c.Created

		// Lookup Messages
		messages := make(map[int64]*MessageData)
		if err := cm.LookupMessages(id, func(m *conveyearthgo.Message) error {
			var content template.HTML
			if err := cm.LookupFiles(m.ID, func(f *conveyearthgo.File) error {
				c, err := cm.ToHTML(f.Hash, f.Mime)
				if err != nil {
					return err
				}
				content += c
				return nil
			}); err != nil {
				return err
			}
			if m.Parent == 0 {
				data.Conversation = m.Conversation
				data.Message = m.ID
				data.Parent = m.Parent
				data.User = m.User
				data.Cost = m.Cost
				data.Yield = m.Yield
				data.Content = content
			} else {
				messages[m.ID] = &MessageData{
					Conversation: m.Conversation,
					Message:      m.ID,
					Created:      m.Created,
					Parent:       m.Parent,
					User:         m.User,
					Cost:         m.Cost,
					Yield:        m.Yield,
					Content:      content,
				}
			}
			return nil
		}); err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		// Set Replies
		for _, m := range messages {
			if m.Parent == data.Message {
				data.Replies = append(data.Replies, m)
			} else {
				p := messages[m.Parent]
				p.Replies = append(p.Replies, m)
			}
		}
		// Sort Replies
		sort.Slice(data.Replies, func(i, j int) bool {
			return data.Replies[i].Yield > data.Replies[j].Yield
		})
		for _, m := range messages {
			sort.Slice(m.Replies, func(i, j int) bool {
				return m.Replies[i].Yield > m.Replies[j].Yield
			})
		}
		if err := ts.ExecuteTemplate(w, "conversation.go.html", data); err != nil {
			log.Println(err)
			return
		}
	})
}
