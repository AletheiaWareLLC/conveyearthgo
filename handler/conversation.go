package handler

import (
	"aletheiaware.com/authgo"
	"aletheiaware.com/conveyearthgo"
	"aletheiaware.com/netgo"
	"aletheiaware.com/netgo/handler"
	"fmt"
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
		var id int64
		if c := strings.TrimSpace(r.FormValue("id")); c != "" {
			if i, err := strconv.ParseInt(c, 10, 64); err != nil {
				log.Println(err)
			} else {
				id = int64(i)
			}
		}
		type GiftData struct {
			Account        *authgo.Account
			GiftID         int64
			ConversationID int64
			MessageID      int64
			Author         *authgo.Account
			Amount         int64
			Created        time.Time
		}
		type MessageData struct {
			Account        *authgo.Account
			ConversationID int64
			MessageID      int64
			ParentID       int64
			Author         *authgo.Account
			Cost           int64
			Yield          int64
			Content        template.HTML
			ShareTitle     string
			ShareURL       string
			Replies        []*MessageData
			Gifts          []*GiftData
			Created        time.Time
		}
		data := struct {
			MessageData
			Live  bool
			Topic string
			Sort  string
		}{
			Live: netgo.IsLive(),
		}
		data.Account = a.CurrentAccount(w, r)
		data.ConversationID = id
		data.Sort = strings.TrimSpace(r.FormValue("sort"))
		c, err := cm.LookupConversation(id)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		data.Topic = c.Topic
		data.Created = c.Created

		scheme := conveyearthgo.Scheme()
		host := conveyearthgo.Host()
		data.ShareTitle = c.Topic
		data.ShareURL = fmt.Sprintf("%s://%s/conversation?id=%d", scheme, host, id)

		// Lookup Messages
		messages := make(map[int64]*MessageData)
		if err := cm.LookupMessages(id, func(m *conveyearthgo.Message) error {
			if m.ParentID == 0 {
				data.ConversationID = m.ConversationID
				data.MessageID = m.ID
				data.Created = m.Created
				data.ParentID = m.ParentID
				data.Author = m.Author
				data.Cost = m.Cost
				data.Yield = m.Yield
			} else {
				messages[m.ID] = &MessageData{
					Account:        data.Account,
					ConversationID: m.ConversationID,
					MessageID:      m.ID,
					Created:        m.Created,
					ParentID:       m.ParentID,
					Author:         m.Author,
					Cost:           m.Cost,
					Yield:          m.Yield,
					ShareTitle:     data.ShareTitle,
					ShareURL:       fmt.Sprintf("%s://%s/conversation?id=%d#message%d", scheme, host, m.ConversationID, m.ID),
				}
			}
			return nil
		}); err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		// Lookup Gifts
		if err := cm.LookupGifts(id, 0, func(g *conveyearthgo.Gift) error {
			gd := &GiftData{
				Account:        data.Account,
				GiftID:         g.ID,
				ConversationID: g.ConversationID,
				MessageID:      g.MessageID,
				Created:        g.Created,
				Author:         g.Author,
				Amount:         g.Amount,
			}
			if g.MessageID == data.MessageID {
				data.Gifts = append(data.Gifts, gd)
			} else {
				p := messages[g.MessageID]
				p.Gifts = append(p.Gifts, gd)
			}
			return nil
		}); err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		// Set Content
		if err := cm.LookupFiles(data.MessageID, func(f *conveyearthgo.File) error {
			c, err := cm.ToHTML(f.Hash, f.Mime)
			if err != nil {
				return err
			}
			data.Content += c
			return nil
		}); err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		// Set Content and Replies
		for _, m := range messages {
			if err := cm.LookupFiles(m.MessageID, func(f *conveyearthgo.File) error {
				c, err := cm.ToHTML(f.Hash, f.Mime)
				if err != nil {
					return err
				}
				m.Content += c
				return nil
			}); err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			if m.ParentID == data.MessageID {
				data.Replies = append(data.Replies, m)
			} else {
				p := messages[m.ParentID]
				p.Replies = append(p.Replies, m)
			}
		}
		var sorter func(*MessageData, *MessageData) bool
		switch data.Sort {
		case "time":
			sorter = func(a, b *MessageData) bool {
				return a.Created.Before(b.Created)
			}
		case "cost":
			sorter = func(a, b *MessageData) bool {
				return a.Cost > b.Cost
			}
		default:
			data.Sort = "yield"
			fallthrough
		case "yield":
			sorter = func(a, b *MessageData) bool {
				return a.Yield > b.Yield
			}
		}
		// Sort Replies
		sort.Slice(data.Replies, func(i, j int) bool {
			return sorter(data.Replies[i], data.Replies[j])
		})
		for _, m := range messages {
			sort.Slice(m.Replies, func(i, j int) bool {
				return sorter(m.Replies[i], m.Replies[j])
			})
		}
		if err := ts.ExecuteTemplate(w, "conversation.go.html", data); err != nil {
			log.Println(err)
			return
		}
	})
}
