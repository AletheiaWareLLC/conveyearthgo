package handler

import (
	"aletheiaware.com/netgo"
	"aletheiaware.com/netgo/handler"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
)

func AttachDigestHandler(m *http.ServeMux, ts *template.Template, dir string) {
	m.Handle("/digest/", handler.Log(http.StripPrefix("/digest/", http.FileServer(http.Dir(dir)))))
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
				// Scan dir for epubs
				files, err := ioutil.ReadDir(dir)
				if err != nil {
					log.Println(err)
					http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
					return
				}

				for _, f := range files {
					if n := f.Name(); strings.HasSuffix(n, ".epub") {
						data.Editions = append(data.Editions, strings.TrimSuffix(strings.TrimPrefix(n, "Convey-Digest-"), ".epub"))
					}
				}

				// Sort Editions Reverse-Chronologically (Newest First)
				sort.Sort(sort.Reverse(sort.StringSlice(data.Editions)))

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
