package handler

import (
	"aletheiaware.com/netgo"
	"aletheiaware.com/netgo/handler"
	"html/template"
	"log"
	"net/http"
)

func AttachAboutHandler(m *http.ServeMux, ts *template.Template) {
	m.Handle("/about", handler.Log(handler.Compress(About(ts))))
}

func About(ts *template.Template) http.Handler {
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
