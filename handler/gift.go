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
)

func AttachGiftHandler(m *http.ServeMux, a authgo.Authenticator, am conveyearthgo.AccountManager, cm conveyearthgo.ContentManager, nm conveyearthgo.NotificationManager, ts *template.Template) {
	m.Handle("/gift", handler.Log(handler.Compress(Gift(a, am, cm, nm, ts))))
}

func Gift(a authgo.Authenticator, am conveyearthgo.AccountManager, cm conveyearthgo.ContentManager, nm conveyearthgo.NotificationManager, ts *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		account := a.CurrentAccount(w, r)
		if account == nil {
			authredirect.SignIn(w, r, r.URL.String())
			return
		}
		data := &GiftData{
			Live:    netgo.IsLive(),
			Account: account,
			Gift:    1,
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
			query := r.URL.Query()
			conversation := netgo.ParseInt(netgo.QueryParameter(query, "conversation"))
			message := netgo.ParseInt(netgo.QueryParameter(query, "message"))

			if err := populateGiftData(cm, conversation, message, data); err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}

			executeGiftTemplate(w, ts, data)
		case "POST":
			conversation := netgo.ParseInt(r.FormValue("conversation"))
			message := netgo.ParseInt(r.FormValue("message"))
			gift := netgo.ParseInt(r.FormValue("gift"))

			data.Gift = gift

			if err := populateGiftData(cm, conversation, message, data); err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}

			// Check account balance
			if gift > balance {
				err := conveyearthgo.ErrInsufficientBalance
				log.Println(err)
				data.Error = err.Error()
				executeGiftTemplate(w, ts, data)
				return
			}

			// Cannot gift to self
			if account.ID == data.Message.Author.ID {
				err := conveyearthgo.ErrSelfGiftingNotPermitted
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
			if err := nm.NotifyGift(data.Message.Author, account, conversation, data.Conversation.Topic, message, gift); err != nil {
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

func populateGiftData(cm conveyearthgo.ContentManager, conversation, message int64, data *GiftData) error {
	c, err := cm.LookupConversation(conversation)
	if err != nil {
		return err
	}
	data.Conversation = c
	m, err := cm.LookupMessage(message)
	if err != nil {
		return err
	}
	data.Message = m
	var content template.HTML
	if err := cm.LookupFiles(message, func(f *conveyearthgo.File) error {
		c, err := cm.ToHTML(f.Hash, f.Mime)
		if err != nil {
			return err
		}
		content += c
		return nil
	}); err != nil {
		return err
	}
	data.Content = content
	return nil
}

type GiftData struct {
	Live         bool
	Error        string
	Account      *authgo.Account
	Balance      int64
	Conversation *conveyearthgo.Conversation
	Message      *conveyearthgo.Message
	Content      template.HTML
	Gift         int64
}
