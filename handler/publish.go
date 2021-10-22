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
	"strings"
)

const (
	MAXIMUM_ATTACHMENTS  = 10
	MAXIMUM_PARSE_MEMORY = 32 << 20
)

func AttachPublishHandler(m *http.ServeMux, a authgo.Authenticator, am conveyearthgo.AccountManager, cm conveyearthgo.ContentManager, nm conveyearthgo.NotificationManager, ts *template.Template) {
	m.Handle("/publish", handler.Log(Publish(a, am, cm, nm, ts)))
}

func Publish(a authgo.Authenticator, am conveyearthgo.AccountManager, cm conveyearthgo.ContentManager, nm conveyearthgo.NotificationManager, ts *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		account := a.CurrentAccount(w, r)
		if account == nil {
			authredirect.SignIn(w, r, r.URL.String())
			return
		}
		data := &PublishData{
			Live:    netgo.IsLive(),
			Account: account,
		}
		balance, err := am.AccountBalance(account.ID)
		if err != nil {
			log.Println(err)
			data.Error = err.Error()
			executePublishTemplate(w, ts, data)
			return
		}
		data.Balance = balance
		switch r.Method {
		case "GET":
			executePublishTemplate(w, ts, data)
		case "POST":
			if err := r.ParseMultipartForm(MAXIMUM_PARSE_MEMORY); err != nil {
				log.Println(err)
				data.Error = err.Error()
				executePublishTemplate(w, ts, data)
				return
			}

			topic := strings.TrimSpace(r.FormValue("topic"))
			content := strings.TrimSpace(r.FormValue("content"))

			// TODO replace \r\n with \n

			data.Topic = topic
			data.Content = content

			// Check valid topic
			if err := conveyearthgo.ValidateTopic(topic); err != nil {
				log.Println(err)
				data.Error = err.Error()
				executePublishTemplate(w, ts, data)
				return
			}

			bytes := []byte(content)

			// Check valid content
			if err := conveyearthgo.ValidateContent(bytes); err != nil {
				log.Println(err)
				data.Error = err.Error()
				executePublishTemplate(w, ts, data)
				return
			}

			var (
				hashes []string
				mimes  []string
				sizes  []int64
				cost   int64
			)

			// Store content
			textHash, textSize, err := cm.AddText(bytes)
			if err != nil {
				log.Println(err)
				data.Error = err.Error()
				executePublishTemplate(w, ts, data)
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
					executePublishTemplate(w, ts, data)
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
					executePublishTemplate(w, ts, data)
					return
				}
				// Check valid mime
				if err := conveyearthgo.ValidateMime(fileMime); err != nil {
					log.Println(err)
					data.Error = err.Error()
					executePublishTemplate(w, ts, data)
					return
				}
				fileHash, fileSize, err := cm.AddFile(file)
				if err != nil {
					log.Println(err)
					data.Error = err.Error()
					executePublishTemplate(w, ts, data)
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
				executePublishTemplate(w, ts, data)
				return
			}

			// Record conversation
			conversation, _, err := cm.NewConversation(account, topic, hashes, mimes, sizes)
			if err != nil {
				log.Println(err)
				data.Error = err.Error()
				executePublishTemplate(w, ts, data)
				return
			}

			// Send Mention Notifications
			for _, username := range conveyearthgo.Mentions(content) {
				a, err := am.Account(username)
				if err != nil {
					log.Println(err)
					continue
				}
				if err := nm.NotifyMention(a, account, conversation.ID, topic, 0); err != nil {
					log.Println(err)
				}
			}

			redirect.Conversation(w, r, conversation.ID, 0)
		}
	})
}

func executePublishTemplate(w http.ResponseWriter, ts *template.Template, data *PublishData) {
	if err := ts.ExecuteTemplate(w, "publish.go.html", data); err != nil {
		log.Println(err)
	}
}

type PublishData struct {
	Live    bool
	Error   string
	Account *authgo.Account
	Balance int64
	Topic   string
	Content string
}
