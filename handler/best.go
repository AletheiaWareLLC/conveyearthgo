package handler

import (
	"aletheiaware.com/authgo"
	"aletheiaware.com/authgo/redirect"
	"aletheiaware.com/conveyearthgo"
	"aletheiaware.com/netgo/handler"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	PERIOD_DAILY  = time.Hour * 24
	PERIOD_WEEKLY = time.Hour * 168  // (24 * 7)
	PERIOD_YEARLY = time.Hour * 8766 // (24 * 365.25)
)

func AttachBestHandler(m *http.ServeMux, a authgo.Authenticator, cm conveyearthgo.ContentManager, ts *template.Template) {
	m.Handle("/best", handler.Log(Best(a, cm, ts)))
}

func Best(a authgo.Authenticator, cm conveyearthgo.ContentManager, ts *template.Template) http.Handler {
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
			Account       *authgo.Account
			Conversations []*ConversationData
			Period        string
			Limit         int64
		}{
			Account: account,
		}
		now := time.Now()
		var since time.Time
		period := r.FormValue("period")
		switch period {
		case "all":
		case "year":
			since = now.Truncate(PERIOD_YEARLY)
		case "week":
			since = now.Truncate(PERIOD_WEEKLY)
		default:
			period = "day"
			fallthrough
		case "day":
			since = now.Truncate(PERIOD_DAILY)
		}
		data.Period = period
		limit := int64(8)
		if l := strings.TrimSpace(r.FormValue("limit")); l != "" {
			if i, err := strconv.ParseInt(l, 10, 64); err != nil {
				log.Println(err)
			} else {
				limit = int64(i)
			}
		}
		data.Limit = limit * 2
		if err := cm.LookupBestConversations(func(c *conveyearthgo.Conversation, cost, yield int64) error {
			data.Conversations = append(data.Conversations, &ConversationData{
				ID:      c.ID,
				Topic:   c.Topic,
				User:    c.User,
				Cost:    cost,
				Yield:   yield,
				Created: c.Created,
			})
			return nil
		}, since, limit); err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		if err := ts.ExecuteTemplate(w, "best.go.html", data); err != nil {
			log.Println(err)
			return
		}
	})
}
