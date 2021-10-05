package handler

import (
	"aletheiaware.com/authgo"
	"aletheiaware.com/authgo/redirect"
	"aletheiaware.com/conveyearthgo"
	"aletheiaware.com/netgo"
	"aletheiaware.com/netgo/handler"
	"html/template"
	"log"
	"net/http"
)

func AttachAccountHandler(m *http.ServeMux, a authgo.Authenticator, am conveyearthgo.AccountManager, nm conveyearthgo.NotificationManager, ts *template.Template) {
	m.Handle("/account", handler.Log(Account(a, am, nm, ts)))
}

func Account(a authgo.Authenticator, am conveyearthgo.AccountManager, nm conveyearthgo.NotificationManager, ts *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		account := a.CurrentAccount(w, r)
		if account == nil {
			redirect.SignIn(w, r, r.URL.String())
			return
		}
		data := &AccountData{
			Account: account,
			Live:    netgo.IsLive(),
		}
		balance, err := am.AccountBalance(account.ID)
		if err != nil {
			log.Println(err)
			data.Error = err.Error()
			executeAccountTemplate(w, ts, data)
			return
		}
		data.Balance = balance
		_, responses, mentions, digests, err := nm.NotificationPreferences(account.ID)
		if err != nil {
			log.Println(err)
			data.Error = err.Error()
			executeAccountTemplate(w, ts, data)
			return
		}
		data.NotificationResponses = responses
		data.NotificationMentions = mentions
		data.NotificationDigests = digests
		executeAccountTemplate(w, ts, data)
	})
}

func executeAccountTemplate(w http.ResponseWriter, ts *template.Template, data *AccountData) {
	if err := ts.ExecuteTemplate(w, "account.go.html", data); err != nil {
		log.Println(err)
	}
}

type AccountData struct {
	Live                  bool
	Error                 string
	Account               *authgo.Account
	Balance               int64
	NotificationResponses bool
	NotificationMentions  bool
	NotificationDigests   bool
}
