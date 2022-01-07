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

func AttachDeleteHandler(m *http.ServeMux, a authgo.Authenticator, am conveyearthgo.AccountManager, cm conveyearthgo.ContentManager, ts *template.Template) {
	m.Handle("/delete", handler.Log(Delete(a, am, cm, ts)))
}

func Delete(a authgo.Authenticator, am conveyearthgo.AccountManager, cm conveyearthgo.ContentManager, ts *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		account := a.CurrentAccount(w, r)
		if account == nil {
			authredirect.SignIn(w, r, r.URL.String())
			return
		}
		data := &DeleteData{
			Live:    netgo.IsLive(),
			Account: account,
		}
		switch r.Method {
		case "GET":
			query := r.URL.Query()
			conversation := netgo.ParseInt(netgo.QueryParameter(query, "conversation"))
			message := netgo.ParseInt(netgo.QueryParameter(query, "message"))
			gift := netgo.ParseInt(netgo.QueryParameter(query, "gift"))

			if err := populateDeleteData(cm, conversation, message, gift, data); err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			executeDeleteTemplate(w, ts, data)
		case "POST":
			conversation := netgo.ParseInt(r.FormValue("conversation"))
			message := netgo.ParseInt(r.FormValue("message"))
			gift := netgo.ParseInt(r.FormValue("gift"))

			if err := populateDeleteData(cm, conversation, message, gift, data); err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}

			var err error
			if message != 0 {
				err = cm.DeleteMessage(account, data.Message)
			} else {
				err = cm.DeleteGift(account, data.Gift)
			}
			if err != nil {
				log.Println(err)
				data.Error = err.Error()
				executeDeleteTemplate(w, ts, data)
				return
			}

			if data.Message != nil {
				if data.Message.ParentID == 0 {
					// Entire conversation was deleted
					authredirect.Index(w, r)
				} else {
					redirect.Conversation(w, r, conversation, data.Message.ParentID)
				}
			} else {
				redirect.Conversation(w, r, conversation, data.Gift.MessageID)
			}
		}
	})
}

func executeDeleteTemplate(w http.ResponseWriter, ts *template.Template, data *DeleteData) {
	if err := ts.ExecuteTemplate(w, "delete.go.html", data); err != nil {
		log.Println(err)
	}
}

func populateDeleteData(cm conveyearthgo.ContentManager, conversation, message, gift int64, data *DeleteData) error {
	c, err := cm.LookupConversation(conversation)
	if err != nil {
		return err
	}
	data.Conversation = c

	if message != 0 {
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
	} else if gift != 0 {
		g, err := cm.LookupGift(gift)
		if err != nil {
			return err
		}
		data.Gift = g
	} else {
		return conveyearthgo.ErrMessageNotFound
	}
	return nil
}

type DeleteData struct {
	Live         bool
	Error        string
	Account      *authgo.Account
	Conversation *conveyearthgo.Conversation
	Message      *conveyearthgo.Message
	Content      template.HTML
	Gift         *conveyearthgo.Gift
}
