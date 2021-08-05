package handler

import (
	"aletheiaware.com/authgo"
	"aletheiaware.com/authgo/redirect"
	"aletheiaware.com/conveyearthgo"
	"aletheiaware.com/netgo"
	"aletheiaware.com/netgo/handler"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
)

func AttachDigestHandler(m *http.ServeMux, a authgo.Authenticator, dm conveyearthgo.DigestManager, ts *template.Template) {
	m.Handle("/digest", handler.Log(Digest(a, dm, ts)))
}

func Digest(a authgo.Authenticator, dm conveyearthgo.DigestManager, ts *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		account := a.CurrentAccount(w, r)
		if account == nil {
			redirect.SignIn(w, r)
			return
		}
		data := &DigestData{
			Live:    netgo.IsLive(),
			Account: account,
			Years:   conveyearthgo.Years,
			Months:  conveyearthgo.Months,
		}
		switch r.Method {
		case "GET":
			executeDigestTemplate(w, ts, data)
		case "POST":
			year := strings.TrimSpace(r.FormValue("year"))
			month := strings.TrimSpace(r.FormValue("month"))

			data.Year = year
			data.Month = month

			name, modified, file, err := dm.Digest(year, month)
			if err != nil {
				log.Println(err)
				data.Error = err.Error()
				executeDigestTemplate(w, ts, data)
				return
			}
			defer file.Close()
			http.ServeContent(w, r, name, modified, file.(io.ReadSeeker))
		}
	})
}

func executeDigestTemplate(w http.ResponseWriter, ts *template.Template, data *DigestData) {
	if err := ts.ExecuteTemplate(w, "digest.go.html", data); err != nil {
		log.Println(err)
	}
}

type DigestData struct {
	Live    bool
	Error   string
	Account *authgo.Account
	Years   []string
	Months  []string
	Year    string
	Month   string
}
