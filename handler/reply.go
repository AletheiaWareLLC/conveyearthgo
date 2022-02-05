package handler

import (
	"aletheiaware.com/authgo"
	authredirect "aletheiaware.com/authgo/redirect"
	"aletheiaware.com/conveyearthgo"
	"aletheiaware.com/conveyearthgo/redirect"
	"aletheiaware.com/netgo"
	"aletheiaware.com/netgo/handler"
	"html/template"
	"log"
	"net/http"
	"strings"
)

func AttachReplyHandler(m *http.ServeMux, a authgo.Authenticator, am conveyearthgo.AccountManager, cm conveyearthgo.ContentManager, nm conveyearthgo.NotificationManager, ts *template.Template) {
	m.Handle("/reply", handler.Log(handler.Compress(Reply(a, am, cm, nm, ts))))
}

func Reply(a authgo.Authenticator, am conveyearthgo.AccountManager, cm conveyearthgo.ContentManager, nm conveyearthgo.NotificationManager, ts *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		account := a.CurrentAccount(w, r)
		if account == nil {
			authredirect.SignIn(w, r, r.URL.String())
			return
		}
		data := &ReplyData{
			Live:    netgo.IsLive(),
			Account: account,
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
			query := r.URL.Query()
			conversation := netgo.ParseInt(netgo.QueryParameter(query, "conversation"))
			message := netgo.ParseInt(netgo.QueryParameter(query, "message"))
			if err := populateReplyData(cm, conversation, message, data); err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			executeReplyTemplate(w, ts, data)
		case "POST":
			if err := r.ParseMultipartForm(MAXIMUM_PARSE_MEMORY); err != nil {
				log.Println(err)
				data.Error = err.Error()
				executeReplyTemplate(w, ts, data)
				return
			}

			conversation := netgo.ParseInt(r.FormValue("conversation"))
			message := netgo.ParseInt(r.FormValue("message"))
			if err := populateReplyData(cm, conversation, message, data); err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}

			reply := strings.ReplaceAll(strings.TrimSpace(r.FormValue("reply")), "\r\n", "\n")

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
			mimes = append(mimes, conveyearthgo.MIME_TEXT_MARKDOWN)
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
				log.Println("Filename:", header.Filename)
				log.Println("Header:", header.Header)
				log.Println("Size:", header.Size)
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
			response, _, err := cm.NewMessage(account, conversation, message, hashes, mimes, sizes)
			if err != nil {
				log.Println(err)
				data.Error = err.Error()
				executeReplyTemplate(w, ts, data)
				return
			}

			// Send Reply Notification
			if account.ID != data.Message.Author.ID {
				if err := nm.NotifyResponse(data.Message.Author, account, conversation, data.Conversation.Topic, response.ID); err != nil {
					log.Println(err)
				}
			}

			// Send Mention Notifications
			for _, username := range conveyearthgo.Mentions(reply) {
				a, err := am.Account(username)
				if err != nil {
					log.Println(err)
					continue
				}
				if err := nm.NotifyMention(a, account, conversation, data.Conversation.Topic, response.ID); err != nil {
					log.Println(err)
				}
			}

			redirect.Conversation(w, r, conversation, response.ID)
		}
	})
}

func executeReplyTemplate(w http.ResponseWriter, ts *template.Template, data *ReplyData) {
	if err := ts.ExecuteTemplate(w, "reply.go.html", data); err != nil {
		log.Println(err)
	}
}

func populateReplyData(cm conveyearthgo.ContentManager, conversation, message int64, data *ReplyData) error {
	c, err := cm.LookupConversation(conversation)
	if err != nil {
		return err
	}
	data.Conversation = c
	m, err := cm.LookupMessage(message)
	if err != nil {
		return err
	}
	data.Message = m
	var content template.HTML
	if err := cm.LookupFiles(message, func(f *conveyearthgo.File) error {
		c, err := cm.ToHTML(f.Hash, f.Mime)
		if err != nil {
			return err
		}
		content += c
		return nil
	}); err != nil {
		return err
	}
	data.Content = content
	return nil
}

type ReplyData struct {
	Live         bool
	Error        string
	Account      *authgo.Account
	Balance      int64
	Conversation *conveyearthgo.Conversation
	Message      *conveyearthgo.Message
	Content      template.HTML
	Reply        string
}
