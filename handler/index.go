package handler

import (
	"aletheiaware.com/authgo"
	"aletheiaware.com/conveyearthgo"
	"aletheiaware.com/netgo"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"
)

func AttachIndexHandler(m *http.ServeMux, a authgo.Authenticator, cm conveyearthgo.ContentManager, ts *template.Template, dir string) {
	m.Handle("/", Index(a, cm, ts, dir))
}

func Index(a authgo.Authenticator, cm conveyearthgo.ContentManager, ts *template.Template, dir string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p := strings.TrimSuffix(r.URL.Path, "index.html"); p != "/" {
			log.Println(r.RemoteAddr, r.Proto, r.Method, r.Host, r.URL, r.Header, "not found")
			http.NotFound(w, r)
			return
		}
		netgo.LogRequest(r)
		limit := int64(8)
		data := struct {
			Live     bool
			Account  *authgo.Account
			Best     []*conveyearthgo.Conversation
			Recent   []*conveyearthgo.Conversation
			Editions []string
			Limit    int64
		}{
			Live:  netgo.IsLive(),
			Limit: limit * 2,
		}
		data.Account = a.CurrentAccount(w, r)

		// Query best of the year posts
		now := time.Now()
		since := time.Date(now.Year()-1, 1, 1, 0, 0, 0, 0, time.UTC)
		if err := cm.LookupBestConversations(func(c *conveyearthgo.Conversation) error {
			data.Best = append(data.Best, c)
			return nil
		}, since, limit); err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		// Query most recent posts
		if err := cm.LookupRecentConversations(func(c *conveyearthgo.Conversation) error {
			data.Recent = append(data.Recent, c)
			return nil
		}, limit); err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		// Query most recent editions
		editions, err := conveyearthgo.ReadDigests(dir)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		// Sort Editions Reverse-Chronologically (Newest First)
		sort.Sort(sort.Reverse(sort.StringSlice(editions)))
		if int64(len(editions)) > limit {
			editions = editions[:limit]
		}
		data.Editions = editions

		if err := ts.ExecuteTemplate(w, "index.go.html", data); err != nil {
			log.Println(err)
			return
		}
	})
}
