package handler

import (
	"aletheiaware.com/authgo"
	"aletheiaware.com/netgo"
	"html/template"
	"log"
	"net/http"
	"strings"
)

func AttachIndexHandler(m *http.ServeMux, a authgo.Authenticator, ts *template.Template) {
	m.Handle("/", Index(a, ts))
}

func Index(a authgo.Authenticator, ts *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p := strings.TrimSuffix(r.URL.Path, "index.html"); p != "/" {
			log.Println(r.RemoteAddr, r.Proto, r.Method, r.Host, r.URL, r.Header, "not found")
			http.NotFound(w, r)
			return
		}
		log.Println(r.RemoteAddr, r.Proto, r.Method, r.Host, r.URL, r.Header)
		data := struct {
			Live    bool
			Account *authgo.Account
		}{
			Live: netgo.IsLive(),
		}
		if account := a.CurrentAccount(w, r); account != nil {
			data.Account = account
		}
		if err := ts.ExecuteTemplate(w, "index.go.html", data); err != nil {
			log.Println(err)
			return
		}
	})
}
