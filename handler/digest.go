package handler

import (
	"aletheiaware.com/conveyearthgo"
	"aletheiaware.com/netgo"
	"aletheiaware.com/netgo/handler"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strings"
)

func AttachDigestHandler(m *http.ServeMux, ts *template.Template, dir, cache string) {
	m.Handle("/digest/", handler.Log(handler.CacheControl(http.StripPrefix("/digest/", http.FileServer(http.Dir(dir))), cache)))
	m.Handle("/digest", handler.Log(Digest(ts, dir)))
}

func Digest(ts *template.Template, dir string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := &DigestData{
			Live: netgo.IsLive(),
		}

		switch r.Method {
		case "GET":
			data.Edition = strings.TrimSpace(r.FormValue("edition"))
			if data.Edition == "" {
				editions, err := conveyearthgo.ReadDigests(dir)
				if err != nil {
					log.Println(err)
					http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
					return
				}

				// Sort Editions Reverse-Chronologically (Newest First)
				sort.Sort(sort.Reverse(sort.StringSlice(editions)))

				data.Editions = editions

				// Render template
				executeDigestTemplate(w, ts, data)
			} else {
				if err := ts.ExecuteTemplate(w, "digest-viewer.go.html", data); err != nil {
					log.Println(err)
				}
			}
		}
	})
}

func executeDigestTemplate(w http.ResponseWriter, ts *template.Template, data *DigestData) {
	if err := ts.ExecuteTemplate(w, "digest.go.html", data); err != nil {
		log.Println(err)
	}
}

type DigestData struct {
	Live     bool
	Edition  string
	Editions []string
}
