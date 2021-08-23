package conveyearthgo

import (
	"aletheiaware.com/authgo"
	"aletheiaware.com/conveyearthgo/content/markdown"
	"aletheiaware.com/conveyearthgo/content/plaintext"
	"aletheiaware.com/cryptogo"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"html/template"
	"io"
	"io/fs"
	"log"
	"mime"
	"mime/multipart"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	MIME_APPLICATION_PDF = "application/pdf"
	MIME_IMAGE_JPEG      = "image/jpeg"
	MIME_IMAGE_JPG       = "image/jpg"
	MIME_IMAGE_GIF       = "image/gif"
	MIME_IMAGE_PNG       = "image/png"
	MIME_IMAGE_SVG       = "image/svg+xml"
	MIME_IMAGE_WEBP      = "image/webp"
	MIME_MODEL_OBJ       = "model/obj"
	MIME_MODEL_MTL       = "model/mtl"
	MIME_MODEL_STL       = "model/stl"
	MIME_TEXT_PLAIN      = "text/plain"
	MIME_TEXT_MARKDOWN   = "text/markdown"
	MIME_VIDEO_MP4       = "video/mp4"
	MIME_VIDEO_OGG       = "video/ogg"
	MIME_VIDEO_WEBM      = "video/webm"

	MINIMUM_CONTENT_LENGTH = 1
)

var (
	ErrContentTooShort      = errors.New("Content Too Short")
	ErrMimeUnrecognized     = errors.New("Unrecognized MIME")
	ErrConversationNotFound = errors.New("Conversation Not Found")
	ErrMessageNotFound      = errors.New("Message Not Found")
	ErrFileNotFound         = errors.New("File Not Found")
	ErrGiftNotFound         = errors.New("Gift Not Found")
	ErrDeletionNotPermitted = errors.New("Deletion Not Permitted")
)

func ValidateContent(content []byte) error {
	if len(content) < MINIMUM_CONTENT_LENGTH {
		return ErrContentTooShort
	}
	return nil
}

func ValidateMime(mime string) error {
	switch mime {
	case MIME_APPLICATION_PDF,
		MIME_IMAGE_JPEG,
		MIME_IMAGE_JPG,
		MIME_IMAGE_GIF,
		MIME_IMAGE_PNG,
		MIME_IMAGE_SVG,
		MIME_IMAGE_WEBP,
		MIME_TEXT_PLAIN,
		MIME_TEXT_MARKDOWN,
		MIME_VIDEO_MP4,
		MIME_VIDEO_OGG,
		MIME_VIDEO_WEBM:
		return nil
	default:
		return ErrMimeUnrecognized
	}
}

func MimeTypeFromHeader(header *multipart.FileHeader) (string, error) {
	mediaType, params, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if err != nil {
		return "", err
	}
	log.Println("MediaType:", mediaType)
	log.Println("Params:", params)

	/* TODO if mime is generic, try infer from file extension
	switch header.Filename {
	}
	*/

	return mediaType, nil
}

var mentions = regexp.MustCompile(`(^|\s)@[[:alnum:]]{3,}`)

func Mentions(input string) []string {
	usernames := make(map[string]struct{})
	for _, u := range mentions.FindAllString(input, -1) {
		usernames[strings.TrimPrefix(strings.TrimSpace(u), "@")] = struct{}{}
	}
	var us []string
	for u := range usernames {
		us = append(us, u)
	}
	return us
}

type ContentDatabase interface {
	CreateConversation(int64, string, time.Time) (int64, error)
	DeleteConversation(int64, int64, time.Time) (int64, error)
	SelectConversation(int64) (*authgo.Account, string, time.Time, error)
	SelectBestConversations(func(int64, *authgo.Account, string, time.Time, int64, int64) error, time.Time, int64) error
	SelectRecentConversations(func(int64, *authgo.Account, string, time.Time, int64, int64) error, int64) error

	CreateMessage(int64, int64, int64, time.Time) (int64, error)
	DeleteMessage(int64, int64, time.Time) (int64, error)
	SelectMessage(int64) (*authgo.Account, int64, int64, time.Time, int64, int64, error)
	SelectMessages(int64, func(int64, *authgo.Account, int64, time.Time, int64, int64) error) error
	SelectMessageParent(int64) (int64, error)

	CreateFile(int64, string, string, time.Time) (int64, error)
	SelectFile(int64) (int64, string, string, time.Time, error)
	SelectFiles(int64, func(int64, string, string, time.Time) error) error

	CreateCharge(int64, int64, int64, int64, time.Time) (int64, error)
	CreateYield(int64, int64, int64, int64, int64, time.Time) (int64, error)

	CreateGift(int64, int64, int64, int64, time.Time) (int64, error)
	DeleteGift(int64, int64, time.Time) (int64, error)
	SelectGift(int64) (int64, int64, *authgo.Account, int64, time.Time, error)
	SelectGifts(int64, int64, func(int64, int64, int64, *authgo.Account, int64, time.Time) error) error
}

type ContentManager interface {
	fs.FS
	AddText([]byte) (string, int64, error)
	AddFile(io.Reader) (string, int64, error)
	ToHTML(string, string) (template.HTML, error)
	NewConversation(*authgo.Account, string, []string, []string, []int64) (*Conversation, *Message, []*File, error)
	LookupConversation(int64) (*Conversation, error)
	LookupBestConversations(func(*Conversation) error, time.Time, int64) error
	LookupRecentConversations(func(*Conversation) error, int64) error
	NewMessage(*authgo.Account, int64, int64, []string, []string, []int64) (*Message, []*File, error)
	DeleteMessage(*authgo.Account, *Message) error
	LookupMessage(int64) (*Message, error)
	LookupMessages(int64, func(*Message) error) error
	LookupFile(int64) (*File, error)
	LookupFiles(int64, func(*File) error) error
	NewGift(*authgo.Account, int64, int64, int64) (*Gift, error)
	DeleteGift(*authgo.Account, *Gift) error
	LookupGift(int64) (*Gift, error)
	LookupGifts(int64, int64, func(*Gift) error) error
}

func NewContentManager(db ContentDatabase, fs Filesystem) ContentManager {
	return &contentManager{
		database:   db,
		filesystem: fs,
	}
}

type contentManager struct {
	database   ContentDatabase
	filesystem Filesystem
}

func (m *contentManager) Open(path string) (fs.File, error) {
	file, err := m.filesystem.Open(path)
	if err != nil {
		return nil, err
	}

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if stat.IsDir() {
		// Directory Listings Disallowed
		return nil, fs.ErrNotExist
	}

	return file, nil
}

func (m *contentManager) AddText(text []byte) (string, int64, error) {
	length := len(text)

	sum := sha512.Sum512(text)

	hash := base64.RawURLEncoding.EncodeToString(sum[:])

	destination, err := m.filesystem.Create(hash)
	if err != nil {
		return "", 0, err
	}

	count, err := destination.Write(text)
	if err != nil {
		return "", 0, err
	}
	if count != length {
		return "", 0, io.ErrShortWrite
	}

	if err := destination.Close(); err != nil {
		return "", 0, err
	}
	return hash, int64(length), nil
}

func (m *contentManager) AddFile(reader io.Reader) (string, int64, error) {
	// Create new file with random name
	name, err := cryptogo.RandomString(20)
	if err != nil {
		return "", 0, err
	}

	destination, err := m.filesystem.Create(name)
	if err != nil {
		return "", 0, err
	}

	hasher := sha512.New()

	reader = io.TeeReader(reader, hasher)

	size, err := io.Copy(destination, reader)
	if err != nil {
		return "", 0, err
	}

	if err := destination.Close(); err != nil {
		return "", 0, err
	}

	hash := base64.RawURLEncoding.EncodeToString(hasher.Sum(nil))

	// Rename file to hash
	if err := m.filesystem.Rename(name, hash); err != nil {
		return "", 0, err
	}

	return hash, size, nil
}

func (m *contentManager) ToHTML(hash, mime string) (template.HTML, error) {
	switch mime {
	case MIME_APPLICATION_PDF,
		MIME_MODEL_OBJ,
		MIME_MODEL_MTL,
		MIME_MODEL_STL:
		return template.HTML(`<object class="ucc" data="/content/` + hash + `?mime=` + url.QueryEscape(mime) + `" type="` + mime + `"><p><small><a href="/content/` + hash + `?mime=` + url.QueryEscape(mime) + `" download>download</a></small></p></object>`), nil
	case MIME_IMAGE_GIF,
		MIME_IMAGE_JPG,
		MIME_IMAGE_JPEG,
		MIME_IMAGE_PNG,
		MIME_IMAGE_SVG,
		MIME_IMAGE_WEBP:
		return template.HTML(`<img class="ucc" src="/content/` + hash + `?mime=` + url.QueryEscape(mime) + `" />`), nil
	case MIME_TEXT_PLAIN:
		file, err := m.Open(hash)
		if err != nil {
			return "", err
		}
		return plaintext.ToHTML(file)
	case MIME_TEXT_MARKDOWN:
		file, err := m.Open(hash)
		if err != nil {
			return "", err
		}
		return markdown.ToHTML(file)
	case MIME_VIDEO_MP4,
		MIME_VIDEO_OGG,
		MIME_VIDEO_WEBM:
		return template.HTML(`<video class="ucc" controls><source src="/content/` + hash + `?mime=` + url.QueryEscape(mime) + `" type="` + mime + `" /><p><small><a href="/content/` + hash + `?mime=` + url.QueryEscape(mime) + `" download>download</a></small></p></video>`), nil
	default:
		return "", ErrMimeUnrecognized
	}
}

func (m *contentManager) NewConversation(account *authgo.Account, topic string, hashes, mimes []string, sizes []int64) (*Conversation, *Message, []*File, error) {
	created := time.Now()
	conversation, err := m.database.CreateConversation(account.ID, topic, created)
	if err != nil {
		return nil, nil, nil, err
	}
	log.Println("Created Conversation", conversation)
	message, err := m.database.CreateMessage(account.ID, conversation, 0, created)
	if err != nil {
		return nil, nil, nil, err
	}
	log.Println("Created Message", message)
	var (
		files []*File
		cost  int64
	)
	for i := 0; i < len(hashes); i++ {
		cost += sizes[i]
		file, err := m.database.CreateFile(message, hashes[i], mimes[i], created)
		if err != nil {
			return nil, nil, nil, err
		}
		log.Println("Created File", file)
		files = append(files, &File{
			ID:      file,
			Message: message,
			Hash:    hashes[i],
			Mime:    mimes[i],
			Created: created,
		})
	}
	charge, err := m.database.CreateCharge(account.ID, conversation, message, cost, created)
	if err != nil {
		return nil, nil, nil, err
	}
	log.Println("Created Charge", charge)
	return &Conversation{
			ID:      conversation,
			Author:  account,
			Topic:   topic,
			Cost:    cost,
			Created: created,
		}, &Message{
			ID:             message,
			Author:         account,
			ConversationID: conversation,
			Cost:           cost,
			Created:        created,
		}, files, nil
}

func (m *contentManager) LookupConversation(id int64) (*Conversation, error) {
	if id == 0 {
		return nil, ErrConversationNotFound
	}
	author, topic, created, err := m.database.SelectConversation(id)
	if err != nil {
		log.Println(err)
		return nil, ErrConversationNotFound
	}
	return &Conversation{
		ID:      id,
		Author:  author,
		Topic:   topic,
		Created: created,
	}, nil
}

func (m *contentManager) LookupBestConversations(callback func(*Conversation) error, since time.Time, limit int64) error {
	return m.database.SelectBestConversations(func(id int64, author *authgo.Account, topic string, created time.Time, cost, yield int64) error {
		return callback(&Conversation{
			ID:      id,
			Author:  author,
			Topic:   topic,
			Cost:    cost,
			Yield:   yield,
			Created: created,
		})
	}, since, limit)
}

func (m *contentManager) LookupRecentConversations(callback func(*Conversation) error, limit int64) error {
	return m.database.SelectRecentConversations(func(id int64, author *authgo.Account, topic string, created time.Time, cost, yield int64) error {
		return callback(&Conversation{
			ID:      id,
			Author:  author,
			Topic:   topic,
			Cost:    cost,
			Yield:   yield,
			Created: created,
		})
	}, limit)
}

func (m *contentManager) NewMessage(account *authgo.Account, conversation, parent int64, hashes, mimes []string, sizes []int64) (*Message, []*File, error) {
	created := time.Now()
	message, err := m.database.CreateMessage(account.ID, conversation, parent, created)
	if err != nil {
		return nil, nil, err
	}
	log.Println("Created Message", message)
	var (
		files []*File
		cost  int64
	)
	for i := 0; i < len(hashes); i++ {
		cost += sizes[i]
		file, err := m.database.CreateFile(message, hashes[i], mimes[i], created)
		if err != nil {
			return nil, nil, err
		}
		log.Println("Created File", file)
		files = append(files, &File{
			ID:      file,
			Message: message,
			Hash:    hashes[i],
			Mime:    mimes[i],
			Created: created,
		})
	}
	charge, err := m.database.CreateCharge(account.ID, conversation, message, cost, created)
	if err != nil {
		return nil, nil, err
	}
	amount := cost
	log.Println("Created Charge", charge)
	for p := parent; p != 0; {
		half := amount / 2
		yield, err := m.database.CreateYield(account.ID, conversation, message, p, half, created)
		if err != nil {
			return nil, nil, err
		}
		log.Println("Created Yield", yield)
		amount = amount - half
		p, err = m.database.SelectMessageParent(p)
		if err != nil {
			return nil, nil, err
		}
	}
	return &Message{
		ID:             message,
		Author:         account,
		ConversationID: conversation,
		ParentID:       parent,
		Cost:           cost,
		Created:        created,
	}, files, nil
}

func (m *contentManager) DeleteMessage(account *authgo.Account, message *Message) error {
	deleted := time.Now()
	count, err := m.database.DeleteMessage(account.ID, message.ID, deleted)
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrDeletionNotPermitted
	}
	log.Println("Deleted Message", message.ID)
	if message.ParentID == 0 {
		count, err := m.database.DeleteConversation(account.ID, message.ConversationID, deleted)
		if err != nil {
			return err
		}
		if count == 0 {
			return ErrDeletionNotPermitted
		}
		log.Println("Deleted Conversation", message.ConversationID)
	}
	return nil
}

func (m *contentManager) LookupMessage(id int64) (*Message, error) {
	if id == 0 {
		return nil, ErrMessageNotFound
	}
	author, conversation, parent, created, cost, yield, err := m.database.SelectMessage(id)
	if err != nil {
		log.Println(err)
		return nil, ErrMessageNotFound
	}
	return &Message{
		ID:             id,
		Author:         author,
		ConversationID: conversation,
		ParentID:       parent,
		Cost:           cost,
		Yield:          yield,
		Created:        created,
	}, nil
}

func (m *contentManager) LookupMessages(conversation int64, callback func(*Message) error) error {
	return m.database.SelectMessages(conversation, func(id int64, author *authgo.Account, parent int64, created time.Time, cost, yield int64) error {
		return callback(&Message{
			ID:             id,
			Author:         author,
			ConversationID: conversation,
			ParentID:       parent,
			Cost:           cost,
			Yield:          yield,
			Created:        created,
		})
	})
}

func (m *contentManager) LookupFile(id int64) (*File, error) {
	if id == 0 {
		return nil, ErrFileNotFound
	}
	message, hash, mime, created, err := m.database.SelectFile(id)
	if err != nil {
		log.Println(err)
		return nil, ErrFileNotFound
	}
	return &File{
		ID:      id,
		Message: message,
		Hash:    hash,
		Mime:    mime,
		Created: created,
	}, nil
}

func (m *contentManager) LookupFiles(message int64, callback func(*File) error) error {
	return m.database.SelectFiles(message, func(id int64, hash, mime string, created time.Time) error {
		return callback(&File{
			ID:      id,
			Message: message,
			Hash:    hash,
			Mime:    mime,
			Created: created,
		})
	})
}

func (m *contentManager) NewGift(account *authgo.Account, conversation, message int64, amount int64) (*Gift, error) {
	created := time.Now()
	gift, err := m.database.CreateGift(account.ID, conversation, message, amount, created)
	if err != nil {
		return nil, err
	}
	log.Println("Created Gift", gift)
	return &Gift{
		ID:             gift,
		Author:         account,
		ConversationID: conversation,
		MessageID:      message,
		Amount:         amount,
		Created:        created,
	}, nil
}

func (m *contentManager) DeleteGift(account *authgo.Account, gift *Gift) error {
	deleted := time.Now()
	count, err := m.database.DeleteGift(account.ID, gift.ID, deleted)
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrDeletionNotPermitted
	}
	log.Println("Deleted Gift", gift.ID)
	return nil
}

func (m *contentManager) LookupGift(id int64) (*Gift, error) {
	if id == 0 {
		return nil, ErrGiftNotFound
	}
	conversation, message, author, amount, created, err := m.database.SelectGift(id)
	if err != nil {
		log.Println(err)
		return nil, ErrGiftNotFound
	}
	return &Gift{
		ID:             id,
		ConversationID: conversation,
		MessageID:      message,
		Author:         author,
		Amount:         amount,
		Created:        created,
	}, nil
}

func (m *contentManager) LookupGifts(conversation, message int64, callback func(*Gift) error) error {
	return m.database.SelectGifts(conversation, message, func(id, conversation, message int64, author *authgo.Account, amount int64, created time.Time) error {
		return callback(&Gift{
			ID:             id,
			Author:         author,
			ConversationID: conversation,
			MessageID:      message,
			Amount:         amount,
			Created:        created,
		})
	})
}
