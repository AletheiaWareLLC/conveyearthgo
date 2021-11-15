package conveyearthgo_test

import (
	"aletheiaware.com/authgo"
	"aletheiaware.com/authgo/authtest"
	"aletheiaware.com/conveyearthgo"
	"aletheiaware.com/conveyearthgo/database"
	"aletheiaware.com/conveyearthgo/filesystem"
	"bytes"
	"github.com/stretchr/testify/assert"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"
)

func TestMentions(t *testing.T) {
	for name, tt := range map[string]struct {
		input    string
		expected []string
	}{
		"empty": {
			input: "",
		},
		"single": {
			input: "@alice",
			expected: []string{
				"alice",
			},
		},
		"single-start": {
			input: "@alice hello",
			expected: []string{
				"alice",
			},
		},
		"single-center": {
			input: "hello @alice!",
			expected: []string{
				"alice",
			},
		},
		"single-end": {
			input: "hello @alice",
			expected: []string{
				"alice",
			},
		},
		"double-comma": {
			input: "@alice, @bob",
			expected: []string{
				"alice",
				"bob",
			},
		},
		"double-space": {
			input: "@alice @bob",
			expected: []string{
				"alice",
				"bob",
			},
		},
		"email": {
			input: "test@example.com",
		},
	} {
		t.Run(name, func(t *testing.T) {
			mentions := conveyearthgo.Mentions(tt.input)
			assert.ElementsMatch(t, tt.expected, mentions)
		})
	}
}

func TestContentManager_AddText(t *testing.T) {
	db := database.NewInMemory()
	dir, err := ioutil.TempDir("", "test")
	assert.NoError(t, err)
	fs := filesystem.NewOnDisk(dir)
	defer os.RemoveAll(dir)
	cm := conveyearthgo.NewContentManager(db, fs)
	contents := []byte("this is a test")
	hash, size, err := cm.AddText(contents)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.Equal(t, int64(len(contents)), size)
}

func TestContentManager_AddFile(t *testing.T) {
	db := database.NewInMemory()
	dir, err := ioutil.TempDir("", "test")
	assert.NoError(t, err)
	fs := filesystem.NewOnDisk(dir)
	defer os.RemoveAll(dir)
	cm := conveyearthgo.NewContentManager(db, fs)
	contents := "this is a test"

	hash, size, err := cm.AddFile(strings.NewReader(contents))
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.Equal(t, int64(len([]byte(contents))), size)
}

func TestContentManager_Open(t *testing.T) {
	db := database.NewInMemory()
	dir, err := ioutil.TempDir("", "test")
	assert.NoError(t, err)
	fs := filesystem.NewOnDisk(dir)
	defer os.RemoveAll(dir)
	cm := conveyearthgo.NewContentManager(db, fs)
	contents := []byte("this is a test")
	hash, size, err := cm.AddText(contents)
	assert.NoError(t, err)
	assert.Equal(t, int64(len(contents)), size)
	t.Run("Exists", func(t *testing.T) {
		file, err := cm.Open(hash)
		assert.NoError(t, err)
		data, err := io.ReadAll(file)
		assert.NoError(t, err)
		assert.Equal(t, contents, data)
	})
	t.Run("Not Exists", func(t *testing.T) {
		file, err := cm.Open("hash1234")
		assert.NotNil(t, err)
		assert.Nil(t, file)
	})
}

func TestContentManager_ToHTML(t *testing.T) {
	db := database.NewInMemory()
	dir, err := ioutil.TempDir("", "test")
	assert.NoError(t, err)
	fs := filesystem.NewOnDisk(dir)
	defer os.RemoveAll(dir)
	cm := conveyearthgo.NewContentManager(db, fs)
	t.Run("Unrecognized", func(t *testing.T) {
		_, err := cm.ToHTML("", "")
		assert.Equal(t, conveyearthgo.ErrMimeUnrecognized, err)
	})
	class := `class="ucc"`
	uri := `/content/z4PhNX7vuL3xVChQ1m2AB9Yg5AULVxXcg_SpIdNs6c5H0NE8XYXysP-DGNKHfuwvY7kxvUdBeoGlODJ6-SfaPg`
	for name, tt := range map[string]struct {
		data     []byte
		mime     string
		expected template.HTML
	}{
		"pdf": {
			mime:     conveyearthgo.MIME_APPLICATION_PDF,
			expected: template.HTML(`<object ` + class + ` data="` + uri + `?mime=application%2Fpdf" type="application/pdf"><p><small><a href="` + uri + `?mime=application%2Fpdf" download>download</a></small></p></object>`),
		},
		"obj": {
			mime:     conveyearthgo.MIME_MODEL_OBJ,
			expected: template.HTML(`<object ` + class + ` data="` + uri + `?mime=model%2Fobj" type="model/obj"><p><small><a href="` + uri + `?mime=model%2Fobj" download>download</a></small></p></object>`),
		},
		"mtl": {
			mime:     conveyearthgo.MIME_MODEL_MTL,
			expected: template.HTML(`<object ` + class + ` data="` + uri + `?mime=model%2Fmtl" type="model/mtl"><p><small><a href="` + uri + `?mime=model%2Fmtl" download>download</a></small></p></object>`),
		},
		"stl": {
			mime:     conveyearthgo.MIME_MODEL_STL,
			expected: template.HTML(`<object ` + class + ` data="` + uri + `?mime=model%2Fstl" type="model/stl"><p><small><a href="` + uri + `?mime=model%2Fstl" download>download</a></small></p></object>`),
		},
		"gif": {
			mime:     conveyearthgo.MIME_IMAGE_GIF,
			expected: template.HTML(`<img ` + class + ` src="` + uri + `?mime=image%2Fgif" />`),
		},
		"jpg": {
			mime:     conveyearthgo.MIME_IMAGE_JPG,
			expected: template.HTML(`<img ` + class + ` src="` + uri + `?mime=image%2Fjpg" />`),
		},
		"jpeg": {
			mime:     conveyearthgo.MIME_IMAGE_JPEG,
			expected: template.HTML(`<img ` + class + ` src="` + uri + `?mime=image%2Fjpeg" />`),
		},
		"png": {
			mime:     conveyearthgo.MIME_IMAGE_PNG,
			expected: template.HTML(`<img ` + class + ` src="` + uri + `?mime=image%2Fpng" />`),
		},
		"svg": {
			mime:     conveyearthgo.MIME_IMAGE_SVG,
			expected: template.HTML(`<img ` + class + ` src="` + uri + `?mime=image%2Fsvg%2Bxml" />`),
		},
		"webp": {
			mime:     conveyearthgo.MIME_IMAGE_WEBP,
			expected: template.HTML(`<img ` + class + ` src="` + uri + `?mime=image%2Fwebp" />`),
		},
		"plain": {
			data: []byte(`This is a

test`),
			mime:     conveyearthgo.MIME_TEXT_PLAIN,
			expected: template.HTML(`<p ` + class + `>This is a</p><p ` + class + `>test</p>`),
		},
		"markdown": {
			data: []byte(`# Title
This is a

test`),
			mime: conveyearthgo.MIME_TEXT_MARKDOWN,
			expected: template.HTML(`<h1 ` + class + `>Title</h1>
<p ` + class + `>This is a</p>
<p ` + class + `>test</p>
`),
		},
	} {
		t.Run(name, func(t *testing.T) {
			hash, _, err := cm.AddFile(bytes.NewReader(tt.data))
			assert.NoError(t, err)
			html, err := cm.ToHTML(hash, tt.mime)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, html)
		})
	}
}

func Test_Conversation(t *testing.T) {
	db := database.NewInMemory()
	ev := authtest.NewEmailVerifier()
	auth := authgo.NewAuthenticator(db, ev)
	acc := authtest.NewTestAccount(t, auth)
	dir, err := ioutil.TempDir("", "test")
	assert.NoError(t, err)
	fs := filesystem.NewOnDisk(dir)
	defer os.RemoveAll(dir)
	cm := conveyearthgo.NewContentManager(db, fs)
	topic := "FooBar"
	content := "Hello World!"
	hash, size, err := cm.AddText([]byte(content))
	assert.NoError(t, err)
	mime := "text/plain"
	c, m, _, err := cm.NewConversation(acc, topic, []string{hash}, []string{mime}, []int64{size})
	assert.NoError(t, err)
	t.Run("Lookup", func(t *testing.T) {
		found, err := cm.LookupConversation(c.ID)
		assert.NoError(t, err)
		assert.Equal(t, c.ID, found.ID)
		assert.Equal(t, c.Author, found.Author)
		assert.Equal(t, c.Topic, found.Topic)
		assert.Equal(t, c.Created, found.Created)
	})
	t.Run("Best", func(t *testing.T) {
		var since time.Time
		cmap := make(map[int64]*conveyearthgo.Conversation)
		assert.NoError(t, cm.LookupBestConversations(func(c *conveyearthgo.Conversation) error {
			cmap[c.ID] = c
			return nil
		}, since, 10))
		assert.Equal(t, 0, len(cmap)) // Conversation has 0 yield, so can't be best

		// Add a Reply
		content := "Hi!"
		hash, size, err := cm.AddText([]byte(content))
		assert.NoError(t, err)
		mime := "text/plain"
		_, _, err = cm.NewMessage(acc, c.ID, m.ID, []string{hash}, []string{mime}, []int64{size})
		assert.NoError(t, err)

		assert.NoError(t, cm.LookupBestConversations(func(c *conveyearthgo.Conversation) error {
			cmap[c.ID] = c
			return nil
		}, since, 10))
		assert.Equal(t, 1, len(cmap)) // Conversation has non-zero yield, so can be best
		found := cmap[c.ID]
		assert.Equal(t, c.Author, found.Author)
		assert.Equal(t, c.Topic, found.Topic)
		assert.Equal(t, c.Created, found.Created)
	})
	t.Run("Recent", func(t *testing.T) {
		cmap := make(map[int64]*conveyearthgo.Conversation)
		assert.NoError(t, cm.LookupRecentConversations(func(c *conveyearthgo.Conversation) error {
			cmap[c.ID] = c
			return nil
		}, 10))
		assert.Equal(t, 1, len(cmap))
		found := cmap[c.ID]
		assert.Equal(t, c.Author, found.Author)
		assert.Equal(t, c.Topic, found.Topic)
		assert.Equal(t, c.Created, found.Created)
	})
}

func Test_Message(t *testing.T) {
	db := database.NewInMemory()
	ev := authtest.NewEmailVerifier()
	auth := authgo.NewAuthenticator(db, ev)
	acc := authtest.NewTestAccount(t, auth)
	dir, err := ioutil.TempDir("", "test")
	assert.NoError(t, err)
	fs := filesystem.NewOnDisk(dir)
	defer os.RemoveAll(dir)
	cm := conveyearthgo.NewContentManager(db, fs)
	topic := "FooBar"
	content := "Hello World!"
	hash, size, err := cm.AddText([]byte(content))
	assert.NoError(t, err)
	mime := "text/plain"
	c, m, _, err := cm.NewConversation(acc, topic, []string{hash}, []string{mime}, []int64{size})
	assert.NoError(t, err)
	t.Run("Lookup", func(t *testing.T) {
		found, err := cm.LookupMessage(m.ID)
		assert.NoError(t, err)
		assert.Equal(t, m.ID, found.ID)
		assert.Equal(t, m.Author, found.Author)
		assert.Equal(t, m.ConversationID, found.ConversationID)
		assert.Equal(t, m.ParentID, found.ParentID)
		assert.Equal(t, m.Cost, found.Cost)
		assert.Equal(t, m.Yield, found.Yield)
		assert.Equal(t, m.Created, found.Created)
	})
	t.Run("Lookups", func(t *testing.T) {
		mmap := make(map[int64]*conveyearthgo.Message)
		assert.NoError(t, cm.LookupMessages(c.ID, func(m *conveyearthgo.Message) error {
			mmap[m.ID] = m
			return nil
		}))
		assert.Equal(t, 1, len(mmap))
		found := mmap[m.ID]
		assert.Equal(t, m.Author, found.Author)
		assert.Equal(t, m.ConversationID, found.ConversationID)
		assert.Equal(t, m.ParentID, found.ParentID)
		assert.Equal(t, m.Cost, found.Cost)
		assert.Equal(t, m.Yield, found.Yield)
		assert.Equal(t, m.Created, found.Created)

		// Add a Reply
		content := "Hi!"
		hash, size, err := cm.AddText([]byte(content))
		assert.NoError(t, err)
		mime := "text/plain"
		r, _, err := cm.NewMessage(acc, c.ID, m.ID, []string{hash}, []string{mime}, []int64{size})
		assert.NoError(t, err)

		assert.NoError(t, cm.LookupMessages(c.ID, func(m *conveyearthgo.Message) error {
			mmap[m.ID] = m
			return nil
		}))
		assert.Equal(t, 2, len(mmap))
		found = mmap[m.ID]
		assert.Equal(t, m.Author, found.Author)
		assert.Equal(t, m.ConversationID, found.ConversationID)
		assert.Equal(t, m.ParentID, found.ParentID)
		assert.Equal(t, m.Cost, found.Cost)
		assert.Equal(t, m.Yield+(size/2), found.Yield)
		assert.Equal(t, m.Created, found.Created)

		found = mmap[r.ID]
		assert.Equal(t, r.Author, found.Author)
		assert.Equal(t, r.ConversationID, found.ConversationID)
		assert.Equal(t, r.ParentID, found.ParentID)
		assert.Equal(t, r.Cost, found.Cost)
		assert.Equal(t, r.Yield, found.Yield)
		assert.Equal(t, r.Created, found.Created)
	})
}

func Test_File(t *testing.T) {
	db := database.NewInMemory()
	ev := authtest.NewEmailVerifier()
	auth := authgo.NewAuthenticator(db, ev)
	acc := authtest.NewTestAccount(t, auth)
	dir, err := ioutil.TempDir("", "test")
	assert.NoError(t, err)
	fs := filesystem.NewOnDisk(dir)
	defer os.RemoveAll(dir)
	cm := conveyearthgo.NewContentManager(db, fs)
	topic := "FooBar"
	content1 := "Hello"
	hash1, size1, err := cm.AddText([]byte(content1))
	content2 := "World!"
	hash2, size2, err := cm.AddText([]byte(content2))
	assert.NoError(t, err)
	mime := "text/plain"
	c, m, files, err := cm.NewConversation(acc, topic, []string{hash1, hash2}, []string{mime, mime}, []int64{size1, size2})
	assert.NoError(t, err)
	t.Run("Lookup", func(t *testing.T) {
		found, err := cm.LookupFile(files[0].ID)
		assert.NoError(t, err)
		assert.Equal(t, m.ID, found.Message)
		assert.Equal(t, hash1, found.Hash)
		assert.Equal(t, mime, found.Mime)
		assert.Equal(t, c.Created, found.Created)

		found, err = cm.LookupFile(files[1].ID)
		assert.NoError(t, err)
		assert.Equal(t, m.ID, found.Message)
		assert.Equal(t, hash2, found.Hash)
		assert.Equal(t, mime, found.Mime)
		assert.Equal(t, c.Created, found.Created)
	})
	t.Run("Lookups", func(t *testing.T) {
		fmap := make(map[int64]*conveyearthgo.File)
		assert.NoError(t, cm.LookupFiles(m.ID, func(f *conveyearthgo.File) error {
			fmap[f.ID] = f
			return nil
		}))
		found := fmap[files[0].ID]
		assert.Equal(t, m.ID, found.Message)
		assert.Equal(t, hash1, found.Hash)
		assert.Equal(t, mime, found.Mime)
		assert.Equal(t, c.Created, found.Created)

		found = fmap[files[1].ID]
		assert.Equal(t, m.ID, found.Message)
		assert.Equal(t, hash2, found.Hash)
		assert.Equal(t, mime, found.Mime)
		assert.Equal(t, c.Created, found.Created)
	})
}
