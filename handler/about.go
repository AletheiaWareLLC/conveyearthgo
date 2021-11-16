package handler

import (
	"aletheiaware.com/authgo"
	"aletheiaware.com/netgo"
	"aletheiaware.com/netgo/handler"
	"html/template"
	"log"
	"net/http"
)

func AttachAboutHandler(m *http.ServeMux, a authgo.Authenticator, ts *template.Template) {
	m.Handle("/about", handler.Log(About(a, ts)))
}

func About(a authgo.Authenticator, ts *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Live bool
		}{
			Live: netgo.IsLive(),
		}
		if err := ts.ExecuteTemplate(w, "about.go.html", data); err != nil {
			log.Println(err)
			return
		}
	})
}
