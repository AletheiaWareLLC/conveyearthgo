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
	"strings"
)

func AttachNotificationPreferencesHandler(m *http.ServeMux, a authgo.Authenticator, nm conveyearthgo.NotificationManager, ts *template.Template) {
	m.Handle("/account-notification-preferences", handler.Log(NotificationPreferences(a, nm, ts)))
}

func NotificationPreferences(a authgo.Authenticator, nm conveyearthgo.NotificationManager, ts *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		account := a.CurrentAccount(w, r)
		if account == nil {
			redirect.SignIn(w, r, r.URL.String())
			return
		}
		data := &NotificationPreferencesData{
			Account: account,
			Live:    netgo.IsLive(),
		}
		id, responses, mentions, digests, err := nm.NotificationPreferences(account.ID)
		if err != nil {
			log.Println(err)
			data.Error = err.Error()
			executeNotificationPreferencesTemplate(w, ts, data)
			return
		}
		data.NotificationResponses = responses
		data.NotificationMentions = mentions
		data.NotificationDigests = digests
		switch r.Method {
		case "GET":
			executeNotificationPreferencesTemplate(w, ts, data)
		case "POST":
			responses := strings.TrimSpace(r.FormValue("responses")) == "yes"
			mentions := strings.TrimSpace(r.FormValue("mentions")) == "yes"
			digests := strings.TrimSpace(r.FormValue("digests")) == "yes"

			data.NotificationResponses = responses
			data.NotificationMentions = mentions
			data.NotificationDigests = digests

			if err := nm.SetNotificationPreferences(id, account.ID, responses, mentions, digests); err != nil {
				log.Println(err)
				data.Error = err.Error()
				executeNotificationPreferencesTemplate(w, ts, data)
			}
			redirect.Account(w, r)
		}
	})
}

func executeNotificationPreferencesTemplate(w http.ResponseWriter, ts *template.Template, data *NotificationPreferencesData) {
	if err := ts.ExecuteTemplate(w, "account-notification-preferences.go.html", data); err != nil {
		log.Println(err)
	}
}

type NotificationPreferencesData struct {
	Live                  bool
	Error                 string
	Account               *authgo.Account
	NotificationResponses bool
	NotificationMentions  bool
	NotificationDigests   bool
}
