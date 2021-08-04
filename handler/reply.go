package handler

import (
	"aletheiaware.com/authgo"
	authredirect "aletheiaware.com/authgo/redirect"
	"aletheiaware.com/conveyearthgo"
	"aletheiaware.com/conveyearthgo/redirect"
	"aletheiaware.com/netgo"
	"aletheiaware.com/netgo/handler"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func AttachReplyHandler(m *http.ServeMux, a authgo.Authenticator, am conveyearthgo.AccountManager, cm conveyearthgo.ContentManager, ts *template.Template) {
	m.Handle("/reply", handler.Log(Reply(a, am, cm, ts)))
}

func Reply(a authgo.Authenticator, am conveyearthgo.AccountManager, cm conveyearthgo.ContentManager, ts *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		account := a.CurrentAccount(w, r)
		if account == nil {
			authredirect.SignIn(w, r)
			return
		}
		var conversation int64
		if c := strings.TrimSpace(r.FormValue("conversation")); c != "" {
			if i, err := strconv.ParseInt(c, 10, 64); err != nil {
				log.Println(err)
			} else {
				conversation = int64(i)
			}
		}
		var message int64
		if m := strings.TrimSpace(r.FormValue("message")); m != "" {
			if i, err := strconv.ParseInt(m, 10, 64); err != nil {
				log.Println(err)
			} else {
				message = int64(i)
			}
		}
		if conversation == 0 || message == 0 {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		m, err := cm.LookupMessage(message)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		var content template.HTML
		if err := cm.LookupFiles(message, func(f *conveyearthgo.File) error {
			c, err := cm.ToHTML(f.Hash, f.Mime)
			if err != nil {
				return err
			}
			content += c
			return nil
		}); err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		data := &ReplyData{
			Live:         netgo.IsLive(),
			Account:      account,
			Conversation: conversation,
			Message:      message,
			User:         m.User,
			Cost:         m.Cost,
			Yield:        m.Yield,
			Content:      content,
			Created:      m.Created,
		}
		balance, err := am.AccountBalance(account.ID)
		if err != nil {
			log.Println(err)
			data.Error = err.Error()
			executeReplyTemplate(w, ts, data)
			return
		}
		data.Balance = balance
		switch r.Method {
		case "GET":
			executeReplyTemplate(w, ts, data)
		case "POST":
			r.ParseMultipartForm(MAXIMUM_PARSE_MEMORY)

			reply := strings.TrimSpace(r.FormValue("reply"))

			data.Reply = reply

			bytes := []byte(reply)

			// Check valid reply
			if err := conveyearthgo.ValidateContent(bytes); err != nil {
				log.Println(err)
				data.Error = err.Error()
				executeReplyTemplate(w, ts, data)
				return
			}

			var (
				hashes []string
				mimes  []string
				sizes  []int64
				cost   int64
			)

			// Store reply
			textHash, textSize, err := cm.AddText(bytes)
			if err != nil {
				log.Println(err)
				data.Error = err.Error()
				executeReplyTemplate(w, ts, data)
				return
			}

			hashes = append(hashes, textHash)
			mimes = append(mimes, conveyearthgo.MIME_TEXT_PLAIN)
			sizes = append(sizes, textSize)
			cost += textSize

			// Store attachment
			file, header, err := r.FormFile("attachment")
			if err != nil {
				if err != http.ErrMissingFile {
					log.Println(err)
					data.Error = err.Error()
					executeReplyTemplate(w, ts, data)
					return
				}
			} else {
				defer file.Close()
				fmt.Printf("Filename: %+v\n", header.Filename)
				fmt.Printf("Header: %+v\n", header.Header)
				fmt.Printf("Size: %+v\n", header.Size)
				fileMime, err := conveyearthgo.MimeTypeFromHeader(header)
				if err != nil {
					log.Println(err)
					data.Error = err.Error()
					executeReplyTemplate(w, ts, data)
					return
				}
				// Check valid mime
				if err := conveyearthgo.ValidateMime(fileMime); err != nil {
					log.Println(err)
					data.Error = err.Error()
					executeReplyTemplate(w, ts, data)
					return
				}
				fileHash, fileSize, err := cm.AddFile(file)
				if err != nil {
					log.Println(err)
					data.Error = err.Error()
					executeReplyTemplate(w, ts, data)
					return
				}
				hashes = append(hashes, fileHash)
				mimes = append(mimes, fileMime)
				sizes = append(sizes, fileSize)
				cost += fileSize
			}

			// Check account balance
			if cost > balance {
				err := conveyearthgo.ErrInsufficientBalance
				log.Println(err)
				data.Error = err.Error()
				executeReplyTemplate(w, ts, data)
				return
			}

			// Record message
			m, err := cm.NewMessage(account, conversation, message, hashes, mimes, sizes)
			if err != nil {
				log.Println(err)
				data.Error = err.Error()
				executeReplyTemplate(w, ts, data)
				return
			}

			redirect.Conversation(w, r, conversation, m.ID)
		}
	})
}

func executeReplyTemplate(w http.ResponseWriter, ts *template.Template, data *ReplyData) {
	if err := ts.ExecuteTemplate(w, "reply.go.html", data); err != nil {
		log.Println(err)
	}
}

type ReplyData struct {
	Live         bool
	Error        string
	Account      *authgo.Account
	Balance      int64
	Conversation int64
	Message      int64
	User         string
	Cost         int64
	Yield        int64
	Content      template.HTML
	Created      time.Time
	Reply        string
}
