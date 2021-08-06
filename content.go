package conveyearthgo

import (
	"aletheiaware.com/authgo"
	"aletheiaware.com/cryptogo"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"html/template"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/url"
	"regexp"
	"time"
)

const (
	MIME_APPLICATION_PDF = "application/pdf"
	MIME_IMAGE_JPEG      = "image/jpeg"
	MIME_IMAGE_GIF       = "image/gif"
	MIME_IMAGE_JPG       = "image/jpg"
	MIME_IMAGE_PNG       = "image/png"
	MIME_IMAGE_SVG       = "image/svg+xml"
	MIME_IMAGE_WEBP      = "image/webp"
	MIME_MODEL_OBJ       = "model/obj"
	MIME_MODEL_MTL       = "model/mtl"
	MIME_MODEL_STL       = "model/stl"
	MIME_TEXT_PLAIN      = "text/plain"

	MINIMUM_CONTENT_LENGTH = 1
)

var (
	ErrContentTooShort  = errors.New("Content Too Short")
	ErrMimeUnrecognized = errors.New("Unrecognized MIME")
)

func ValidateContent(content []byte) error {
	if len(content) < MINIMUM_TOPIC_LENGTH {
		return ErrContentTooShort
	}
	return nil
}

func ValidateMime(mime string) error {
	switch mime {
	case MIME_APPLICATION_PDF,
		MIME_IMAGE_JPEG,
		MIME_IMAGE_GIF,
		MIME_IMAGE_JPG,
		MIME_IMAGE_PNG,
		MIME_IMAGE_SVG,
		MIME_IMAGE_WEBP,
		MIME_TEXT_PLAIN:
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

var (
	newlines = regexp.MustCompile(`\r?\n\r?\n`)
	anchors  = regexp.MustCompile(`\b(file|ftp|https?):\/\/\S+[\/\w]`)
)

type ContentDatabase interface {
	CreateConversation(int64, string, time.Time) (int64, error)
	SelectConversation(int64) (int64, string, string, time.Time, error)
	LookupBestConversations(func(int64, int64, string, string, time.Time, int64, int64) error, time.Time, int64) error
	LookupRecentConversations(func(int64, int64, string, string, time.Time, int64, int64) error, int64) error

	CreateMessage(int64, int64, int64, time.Time) (int64, error)
	SelectMessage(int64) (int64, string, int64, int64, time.Time, int64, int64, error)
	LookupMessage(int64, int64) (int64, int64, time.Time, int64, int64, error)
	LookupMessages(int64, func(int64, int64, string, int64, time.Time, int64, int64) error) error
	LookupMessageParent(int64) (int64, error)

	CreateFile(int64, string, string, time.Time) (int64, error)
	SelectFile(int64) (int64, string, string, time.Time, error)
	LookupFiles(int64, func(int64, string, string, time.Time) error) error

	CreateCharge(int64, int64, int64, int64, time.Time) (int64, error)
	CreateYield(int64, int64, int64, int64, int64, time.Time) (int64, error)
}

type ContentManager interface {
	fs.FS
	AddText([]byte) (string, int64, error)
	AddFile(io.Reader) (string, int64, error)
	ToHTML(string, string) (template.HTML, error)
	NewConversation(*authgo.Account, string, []string, []string, []int64) (*Conversation, *Message, error)
	LookupConversation(int64) (*Conversation, error)
	LookupBestConversations(func(*Conversation) error, time.Time, int64) error
	LookupRecentConversations(func(*Conversation) error, int64) error
	NewMessage(*authgo.Account, int64, int64, []string, []string, []int64) (*Message, error)
	LookupMessage(int64) (*Message, error)
	LookupMessages(int64, func(*Message) error) error
	LookupFile(int64) (*File, error)
	LookupFiles(int64, func(*File) error) error
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
		return template.HTML("<object class=\"user-content\" data=\"/content/" + hash + "?mime=" + url.QueryEscape(mime) + "\" type=\"" + mime + "\"><p><small><a href=\"/content/" + hash + "?mime=" + url.QueryEscape(mime) + "\" download>download</a></small></p></object>"), nil
	case MIME_IMAGE_JPEG,
		MIME_IMAGE_GIF,
		MIME_IMAGE_JPG,
		MIME_IMAGE_PNG,
		MIME_IMAGE_SVG,
		MIME_IMAGE_WEBP:
		return template.HTML("<img class=\"user-content\" src=\"/content/" + hash + "?mime=" + url.QueryEscape(mime) + "\" />"), nil
	case MIME_TEXT_PLAIN:
		file, err := m.Open(hash)
		if err != nil {
			return "", err
		}
		data, err := ioutil.ReadAll(file)
		if err != nil {
			return "", err
		}
		safe := template.HTMLEscapeString(string(data))
		safe = anchors.ReplaceAllString(safe, `<a href="$0">$0</a>`)
		safe = newlines.ReplaceAllString(safe, `</p><p>`)
		return template.HTML(`<p>` + safe + `</p>`), nil
	default:
		return "", ErrMimeUnrecognized
	}
}

func (m *contentManager) NewConversation(account *authgo.Account, topic string, hashes, mimes []string, sizes []int64) (*Conversation, *Message, error) {
	created := time.Now()
	conversation, err := m.database.CreateConversation(account.ID, topic, created)
	if err != nil {
		return nil, nil, err
	}
	log.Println("Created Conversation", conversation)
	message, err := m.database.CreateMessage(account.ID, conversation, 0, created)
	if err != nil {
		return nil, nil, err
	}
	log.Println("Created Message", message)
	var cost int64
	for i := 0; i < len(hashes); i++ {
		cost += sizes[i]
		file, err := m.database.CreateFile(message, hashes[i], mimes[i], created)
		if err != nil {
			return nil, nil, err
		}
		log.Println("Created File", file)
	}
	charge, err := m.database.CreateCharge(account.ID, conversation, message, cost, created)
	if err != nil {
		return nil, nil, err
	}
	log.Println("Created Charge", charge)
	return &Conversation{
			ID:      conversation,
			User:    account.Username,
			Topic:   topic,
			Created: created,
		}, &Message{
			ID:           message,
			User:         account.Username,
			Conversation: conversation,
			Cost:         cost,
			Created:      created,
		}, nil
}

func (m *contentManager) LookupConversation(id int64) (*Conversation, error) {
	_, username, topic, created, err := m.database.SelectConversation(id)
	if err != nil {
		return nil, err
	}
	return &Conversation{
		ID:      id,
		User:    username,
		Topic:   topic,
		Created: created,
	}, nil
}

func (m *contentManager) LookupBestConversations(callback func(*Conversation) error, since time.Time, limit int64) error {
	return m.database.LookupBestConversations(func(id, user int64, username, topic string, created time.Time, cost, yield int64) error {
		return callback(&Conversation{
			ID:      id,
			User:    username,
			Topic:   topic,
			Cost:    cost,
			Yield:   yield,
			Created: created,
		})
	}, since, limit)
}

func (m *contentManager) LookupRecentConversations(callback func(*Conversation) error, limit int64) error {
	return m.database.LookupRecentConversations(func(id, user int64, username, topic string, created time.Time, cost, yield int64) error {
		return callback(&Conversation{
			ID:      id,
			User:    username,
			Topic:   topic,
			Cost:    cost,
			Yield:   yield,
			Created: created,
		})
	}, limit)
}

func (m *contentManager) NewMessage(account *authgo.Account, conversation, parent int64, hashes, mimes []string, sizes []int64) (*Message, error) {
	created := time.Now()
	message, err := m.database.CreateMessage(account.ID, conversation, parent, created)
	if err != nil {
		return nil, err
	}
	log.Println("Created Message", message)
	var cost int64
	for i := 0; i < len(hashes); i++ {
		cost += sizes[i]
		file, err := m.database.CreateFile(message, hashes[i], mimes[i], created)
		if err != nil {
			return nil, err
		}
		log.Println("Created File", file)
	}
	charge, err := m.database.CreateCharge(account.ID, conversation, message, cost, created)
	if err != nil {
		return nil, err
	}
	amount := cost
	log.Println("Created Charge", charge)
	for parent != 0 {
		half := amount / 2
		yield, err := m.database.CreateYield(account.ID, conversation, message, parent, half, created)
		if err != nil {
			return nil, err
		}
		log.Println("Created Yield", yield)
		amount = amount - half
		parent, err = m.database.LookupMessageParent(parent)
		if err != nil {
			return nil, err
		}
	}
	return &Message{
		ID:           message,
		User:         account.Username,
		Conversation: conversation,
		Parent:       parent,
		Cost:         cost,
		Created:      created,
	}, nil
}

func (m *contentManager) LookupMessage(id int64) (*Message, error) {
	_, username, conversation, parent, created, cost, yield, err := m.database.SelectMessage(id)
	if err != nil {
		return nil, err
	}
	return &Message{
		ID:           id,
		User:         username,
		Conversation: conversation,
		Parent:       parent,
		Cost:         cost,
		Yield:        yield,
		Created:      created,
	}, nil
}

func (m *contentManager) LookupMessages(conversation int64, callback func(*Message) error) error {
	return m.database.LookupMessages(conversation, func(id, user int64, username string, parent int64, created time.Time, cost, yield int64) error {
		return callback(&Message{
			ID:           id,
			User:         username,
			Conversation: conversation,
			Parent:       parent,
			Cost:         cost,
			Yield:        yield,
			Created:      created,
		})
	})
}

func (m *contentManager) LookupFile(id int64) (*File, error) {
	message, hash, mime, created, err := m.database.SelectFile(id)
	if err != nil {
		return nil, err
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
	return m.database.LookupFiles(message, func(id int64, hash, mime string, created time.Time) error {
		return callback(&File{
			ID:      id,
			Message: message,
			Hash:    hash,
			Mime:    mime,
			Created: created,
		})
	})
}
