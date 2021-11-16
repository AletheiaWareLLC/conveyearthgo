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
	"time"
)

func AttachBestHandler(m *http.ServeMux, a authgo.Authenticator, cm conveyearthgo.ContentManager, ts *template.Template) {
	m.Handle("/best", handler.Log(Best(a, cm, ts)))
}

func Best(a authgo.Authenticator, cm conveyearthgo.ContentManager, ts *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Live          bool
			Account       *authgo.Account
			Conversations []*conveyearthgo.Conversation
			Period        string
			Limit         int64
		}{
			Live: netgo.IsLive(),
		}
		now := time.Now()
		var since time.Time
		period := r.FormValue("period")
		switch period {
		case "all":
		case "year":
			since = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		case "month":
			since = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		default:
			period = "week"
			fallthrough
		case "week":
			since = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
			for since.Weekday() > time.Sunday {
				since = since.AddDate(0, 0, -1)
			}
		case "day":
			since = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
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
		account := a.CurrentAccount(w, r)
		if account == nil {
			if limit > 100 {
				limit = 100
			}
		} else {
			data.Account = account
		}
		if err := cm.LookupBestConversations(func(c *conveyearthgo.Conversation) error {
			data.Conversations = append(data.Conversations, c)
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
