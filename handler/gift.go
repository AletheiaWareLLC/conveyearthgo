package handler

import (
	"aletheiaware.com/authgo"
	authredirect "aletheiaware.com/authgo/redirect"
	"aletheiaware.com/conveyearthgo"
	"aletheiaware.com/conveyearthgo/redirect"
	"aletheiaware.com/netgo"
	"aletheiaware.com/netgo/handler"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func AttachGiftHandler(m *http.ServeMux, a authgo.Authenticator, am conveyearthgo.AccountManager, cm conveyearthgo.ContentManager, nm conveyearthgo.NotificationManager, ts *template.Template) {
	m.Handle("/gift", handler.Log(Gift(a, am, cm, nm, ts)))
}

func Gift(a authgo.Authenticator, am conveyearthgo.AccountManager, cm conveyearthgo.ContentManager, nm conveyearthgo.NotificationManager, ts *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		account := a.CurrentAccount(w, r)
		if account == nil {
			authredirect.SignIn(w, r, r.URL.String())
			return
		}
		var conversation int64
		if c := strings.TrimSpace(r.FormValue("conversation")); c != "" {
			if i, err := strconv.ParseInt(c, 10, 64); err != nil {
				log.Println(err)
			} else {
				conversation = int64(i)
			}
		}
		var message int64
		if m := strings.TrimSpace(r.FormValue("message")); m != "" {
			if i, err := strconv.ParseInt(m, 10, 64); err != nil {
				log.Println(err)
			} else {
				message = int64(i)
			}
		}
		if conversation == 0 || message == 0 {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		c, err := cm.LookupConversation(conversation)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		m, err := cm.LookupMessage(message)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		var content template.HTML
		if err := cm.LookupFiles(message, func(f *conveyearthgo.File) error {
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
		data := &GiftData{
			Live:           netgo.IsLive(),
			Account:        account,
			Topic:          c.Topic,
			ConversationID: conversation,
			MessageID:      message,
			Author:         m.Author,
			Cost:           m.Cost,
			Yield:          m.Yield,
			Content:        content,
			Created:        m.Created,
			Gift:           1,
		}
		balance, err := am.AccountBalance(account.ID)
		if err != nil {
			log.Println(err)
			data.Error = err.Error()
			executeGiftTemplate(w, ts, data)
			return
		}
		data.Balance = balance
		switch r.Method {
		case "GET":
			executeGiftTemplate(w, ts, data)
		case "POST":
			var gift int64
			if g := strings.TrimSpace(r.FormValue("gift")); g != "" {
				if i, err := strconv.ParseInt(g, 10, 64); err != nil {
					log.Println(err)
				} else {
					gift = int64(i)
				}
			}

			data.Gift = gift

			// Check account balance
			if gift > balance {
				err := conveyearthgo.ErrInsufficientBalance
				log.Println(err)
				data.Error = err.Error()
				executeGiftTemplate(w, ts, data)
				return
			}

			// Cannot gift to self
			if account.ID == m.Author.ID {
				err := conveyearthgo.ErrSelfGift
				log.Println(err)
				data.Error = err.Error()
				executeGiftTemplate(w, ts, data)
				return
			}

			// Record gift
			_, err := cm.NewGift(account, conversation, message, gift)
			if err != nil {
				log.Println(err)
				data.Error = err.Error()
				executeGiftTemplate(w, ts, data)
				return
			}

			// Send Gift Notification
			if err := nm.NotifyGift(m.Author, account, conversation, c.Topic, message, gift); err != nil {
				log.Println(err)
			}

			redirect.Conversation(w, r, conversation, message)
		}
	})
}

func executeGiftTemplate(w http.ResponseWriter, ts *template.Template, data *GiftData) {
	if err := ts.ExecuteTemplate(w, "gift.go.html", data); err != nil {
		log.Println(err)
	}
}

type GiftData struct {
	Live           bool
	Error          string
	Account        *authgo.Account
	Balance        int64
	Topic          string
	ConversationID int64
	MessageID      int64
	Author         *authgo.Account
	Cost           int64
	Yield          int64
	Content        template.HTML
	Created        time.Time
	Gift           int64
}
