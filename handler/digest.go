package handler

import (
	"aletheiaware.com/authgo"
	"aletheiaware.com/conveyearthgo"
	"aletheiaware.com/netgo"
	"aletheiaware.com/netgo/handler"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strings"
)

func AttachDigestHandler(m *http.ServeMux, a authgo.Authenticator, ts *template.Template, dir, cache string) {
	m.Handle("/digest/", handler.Log(handler.Compress(handler.CacheControl(http.StripPrefix("/digest/", http.FileServer(http.Dir(dir))), cache))))
	m.Handle("/digest", handler.Log(handler.Compress(Digest(a, ts, dir))))
}

func Digest(a authgo.Authenticator, ts *template.Template, dir string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := &DigestData{
			Live:    netgo.IsLive(),
			Account: a.CurrentAccount(w, r),
		}

		switch r.Method {
		case "GET":
			template := "digest-viewer.go.html"

			data.Edition = strings.TrimSpace(r.FormValue("edition"))

			if data.Edition == "" {
				template = "digest.go.html"

				editions, err := conveyearthgo.ReadDigests(dir)
				if err != nil {
					log.Println(err)
					http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
					return
				}

				// Sort Editions Reverse-Chronologically (Newest First)
				sort.Sort(sort.Reverse(sort.StringSlice(editions)))

				data.Editions = editions
			}
			executeDigestTemplate(w, ts, template, data)
		}
	})
}

func executeDigestTemplate(w http.ResponseWriter, ts *template.Template, template string, data *DigestData) {
	if err := ts.ExecuteTemplate(w, template, data); err != nil {
		log.Println(err)
	}
}

type DigestData struct {
	Live     bool
	Account  *authgo.Account
	Edition  string
	Editions []string
}
