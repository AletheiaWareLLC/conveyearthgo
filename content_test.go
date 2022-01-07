package conveyearthgo_test

import (
	"aletheiaware.com/authgo"
	"aletheiaware.com/authgo/authtest"
	"aletheiaware.com/conveyearthgo"
	"aletheiaware.com/conveyearthgo/conveytest"
	"aletheiaware.com/conveyearthgo/database"
	"aletheiaware.com/conveyearthgo/filesystem"
	"bytes"
	"github.com/stretchr/testify/assert"
	"html/template"
	"io"
	iofs "io/fs"
	"os"
	"strings"
	"testing"
	"time"
)

func TestValidateContent(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		assert.NoError(t, conveyearthgo.ValidateContent([]byte("Test")))
	})
	t.Run("Short", func(t *testing.T) {
		content := []byte(strings.Repeat("x", conveyearthgo.MINIMUM_TOPIC_LENGTH-1))
		assert.Error(t, conveyearthgo.ErrContentTooShort, conveyearthgo.ValidateContent(content))
	})
}

func TestValidateMime(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		for name, mime := range map[string]string{
			"pdf":      conveyearthgo.MIME_APPLICATION_PDF,
			"jpeg":     conveyearthgo.MIME_IMAGE_JPEG,
			"gif":      conveyearthgo.MIME_IMAGE_GIF,
			"jpg":      conveyearthgo.MIME_IMAGE_JPG,
			"png":      conveyearthgo.MIME_IMAGE_PNG,
			"svg":      conveyearthgo.MIME_IMAGE_SVG,
			"webp":     conveyearthgo.MIME_IMAGE_WEBP,
			"plain":    conveyearthgo.MIME_TEXT_PLAIN,
			"markdown": conveyearthgo.MIME_TEXT_MARKDOWN,
			"mp4":      conveyearthgo.MIME_VIDEO_MP4,
			"ogg":      conveyearthgo.MIME_VIDEO_OGG,
			"webm":     conveyearthgo.MIME_VIDEO_WEBM,
		} {
			t.Run(name, func(t *testing.T) {
				assert.NoError(t, conveyearthgo.ValidateMime(mime))
			})
		}
	})
	t.Run("Invalid", func(t *testing.T) {
		for name, mime := range map[string]string{
			"empty": "",
			"slash": "/",
			"wild":  "image/*",
			"any":   "*/*",
			"obj":   conveyearthgo.MIME_MODEL_OBJ,
			"mtl":   conveyearthgo.MIME_MODEL_MTL,
			"stl":   conveyearthgo.MIME_MODEL_STL,
		} {
			t.Run(name, func(t *testing.T) {
				assert.Error(t, conveyearthgo.ErrMimeUnrecognized, conveyearthgo.ValidateMime(mime))
			})
		}
	})
}

func TestMimeTypeFromHeader(t *testing.T) {
	// TODO
}

func TestMentions(t *testing.T) {
	for name, tt := range map[string]struct {
		input    string
		expected []string
	}{
		"empty": {},
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

func TestContentManager_Open(t *testing.T) {
	db := database.NewInMemory()
	dir, err := os.MkdirTemp("", "test")
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
		assert.Error(t, iofs.ErrNotExist, err)
		assert.Nil(t, file)
	})
	t.Run("Directory", func(t *testing.T) {
		file, err := cm.Open("/")
		assert.Error(t, iofs.ErrNotExist, err)
		assert.Nil(t, file)
	})
}

func TestContentManager_AddText(t *testing.T) {
	db := database.NewInMemory()
	dir, err := os.MkdirTemp("", "test")
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
	dir, err := os.MkdirTemp("", "test")
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

func TestContentManager_ToHTML(t *testing.T) {
	db := database.NewInMemory()
	dir, err := os.MkdirTemp("", "test")
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
		"mp4": {
			mime:     conveyearthgo.MIME_VIDEO_MP4,
			expected: template.HTML(`<video ` + class + ` controls><source src="` + uri + `?mime=video%2Fmp4" type="video/mp4" /><p><small><a href="` + uri + `?mime=video%2Fmp4" download>download</a></small></p></video>`),
		},
		"ogg": {
			mime:     conveyearthgo.MIME_VIDEO_OGG,
			expected: template.HTML(`<video ` + class + ` controls><source src="` + uri + `?mime=video%2Fogg" type="video/ogg" /><p><small><a href="` + uri + `?mime=video%2Fogg" download>download</a></small></p></video>`),
		},
		"webm": {
			mime:     conveyearthgo.MIME_VIDEO_WEBM,
			expected: template.HTML(`<video ` + class + ` controls><source src="` + uri + `?mime=video%2Fwebm" type="video/webm" /><p><small><a href="` + uri + `?mime=video%2Fwebm" download>download</a></small></p></video>`),
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

func TestContentManager_NewConversation(t *testing.T) {
	db := database.NewInMemory()
	ev := authtest.NewEmailVerifier()
	auth := authgo.NewAuthenticator(db, ev)
	acc := authtest.NewTestAccount(t, auth)
	dir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err)
	fs := filesystem.NewOnDisk(dir)
	defer os.RemoveAll(dir)
	cm := conveyearthgo.NewContentManager(db, fs)
	c, m, _ := conveytest.NewConversation(t, cm, acc)
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

		conveytest.NewReply(t, cm, acc, c, m)

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

func TestContentManager_DeleteConversation(t *testing.T) {
	db := database.NewInMemory()
	ev := authtest.NewEmailVerifier()
	auth := authgo.NewAuthenticator(db, ev)
	acc := authtest.NewTestAccount(t, auth)
	dir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err)
	fs := filesystem.NewOnDisk(dir)
	defer os.RemoveAll(dir)
	cm := conveyearthgo.NewContentManager(db, fs)
	c, m, _ := conveytest.NewConversation(t, cm, acc)

	// Cannot delete a message you did not create
	acc2, err := auth.NewAccount("2"+authtest.TEST_EMAIL, authtest.TEST_USERNAME+"2", []byte(authtest.TEST_PASSWORD))
	assert.Nil(t, err)
	err = cm.DeleteMessage(acc2, m)
	assert.Error(t, conveyearthgo.ErrDeletionNotPermitted, err)

	// Can delete a message you created
	err = cm.DeleteMessage(acc, m)
	assert.NoError(t, err)
	t.Run("Lookup", func(t *testing.T) {
		_, err := cm.LookupConversation(c.ID)
		assert.Error(t, conveyearthgo.ErrConversationNotFound, err)
	})
	t.Run("Best", func(t *testing.T) {
		var since time.Time
		cmap := make(map[int64]*conveyearthgo.Conversation)
		assert.NoError(t, cm.LookupBestConversations(func(c *conveyearthgo.Conversation) error {
			cmap[c.ID] = c
			return nil
		}, since, 10))
		assert.Equal(t, 0, len(cmap))
	})
	t.Run("Recent", func(t *testing.T) {
		cmap := make(map[int64]*conveyearthgo.Conversation)
		assert.NoError(t, cm.LookupRecentConversations(func(c *conveyearthgo.Conversation) error {
			cmap[c.ID] = c
			return nil
		}, 10))
		assert.Equal(t, 0, len(cmap))
	})
}

func TestContentManager_NewMessage(t *testing.T) {
	db := database.NewInMemory()
	ev := authtest.NewEmailVerifier()
	auth := authgo.NewAuthenticator(db, ev)
	acc := authtest.NewTestAccount(t, auth)
	dir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err)
	fs := filesystem.NewOnDisk(dir)
	defer os.RemoveAll(dir)
	cm := conveyearthgo.NewContentManager(db, fs)
	c, m1, f1 := conveytest.NewConversation(t, cm, acc)
	t.Run("LookupMessage", func(t *testing.T) {
		found, err := cm.LookupMessage(m1.ID)
		assert.NoError(t, err)
		assert.Equal(t, m1.ID, found.ID)
		assert.Equal(t, m1.Author, found.Author)
		assert.Equal(t, m1.ConversationID, found.ConversationID)
		assert.Equal(t, m1.ParentID, found.ParentID)
		assert.Equal(t, m1.Cost, found.Cost)
		assert.Equal(t, m1.Yield, found.Yield)
		assert.Equal(t, m1.Created, found.Created)
	})
	t.Run("LookupMessages", func(t *testing.T) {
		mmap := make(map[int64]*conveyearthgo.Message)
		assert.NoError(t, cm.LookupMessages(c.ID, func(m *conveyearthgo.Message) error {
			mmap[m.ID] = m
			return nil
		}))
		assert.Equal(t, 1, len(mmap))
		found := mmap[m1.ID]
		assert.Equal(t, m1.Author, found.Author)
		assert.Equal(t, m1.ConversationID, found.ConversationID)
		assert.Equal(t, m1.ParentID, found.ParentID)
		assert.Equal(t, m1.Cost, found.Cost)
		assert.Equal(t, m1.Yield, found.Yield)
		assert.Equal(t, m1.Created, found.Created)

		m2, _ := conveytest.NewReply(t, cm, acc, c, m1)

		assert.NoError(t, cm.LookupMessages(c.ID, func(m *conveyearthgo.Message) error {
			mmap[m.ID] = m
			return nil
		}))
		assert.Equal(t, 2, len(mmap))
		found = mmap[m1.ID]
		assert.Equal(t, m1.Author, found.Author)
		assert.Equal(t, m1.ConversationID, found.ConversationID)
		assert.Equal(t, m1.ParentID, found.ParentID)
		assert.Equal(t, m1.Cost, found.Cost)
		assert.Equal(t, m1.Yield+(m2.Cost/2), found.Yield)
		assert.Equal(t, m1.Created, found.Created)

		found = mmap[m2.ID]
		assert.Equal(t, m2.Author, found.Author)
		assert.Equal(t, m2.ConversationID, found.ConversationID)
		assert.Equal(t, m2.ParentID, found.ParentID)
		assert.Equal(t, m2.Cost, found.Cost)
		assert.Equal(t, m2.Yield, found.Yield)
		assert.Equal(t, m2.Created, found.Created)
	})
	t.Run("LookupFile", func(t *testing.T) {
		found, err := cm.LookupFile(f1[0].ID)
		assert.NoError(t, err)
		assert.Equal(t, m1.ID, found.Message)
		assert.Equal(t, m1.Created, found.Created)
	})
	t.Run("LookupFiles", func(t *testing.T) {
		fmap := make(map[int64]*conveyearthgo.File)
		assert.NoError(t, cm.LookupFiles(m1.ID, func(f *conveyearthgo.File) error {
			fmap[f.ID] = f
			return nil
		}))
		found := fmap[f1[0].ID]
		assert.Equal(t, m1.ID, found.Message)
		assert.Equal(t, m1.Created, found.Created)
	})
}

func TestContentManager_DeleteMessage(t *testing.T) {
	db := database.NewInMemory()
	ev := authtest.NewEmailVerifier()
	auth := authgo.NewAuthenticator(db, ev)
	acc := authtest.NewTestAccount(t, auth)
	dir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err)
	fs := filesystem.NewOnDisk(dir)
	defer os.RemoveAll(dir)
	cm := conveyearthgo.NewContentManager(db, fs)

	c, m1, _ := conveytest.NewConversation(t, cm, acc)
	m2, f2 := conveytest.NewReply(t, cm, acc, c, m1)

	// Cannot delete a message you did not create
	acc2, err := auth.NewAccount("2"+authtest.TEST_EMAIL, authtest.TEST_USERNAME+"2", []byte(authtest.TEST_PASSWORD))
	assert.Nil(t, err)
	err = cm.DeleteMessage(acc2, m2)
	assert.Error(t, conveyearthgo.ErrDeletionNotPermitted, err)

	// Can delete a message you created
	err = cm.DeleteMessage(acc, m2)
	assert.NoError(t, err)
	t.Run("LookupMessage", func(t *testing.T) {
		_, err := cm.LookupMessage(m2.ID)
		assert.Error(t, conveyearthgo.ErrMessageNotFound, err)
	})
	t.Run("LookupMessages", func(t *testing.T) {
		mmap := make(map[int64]*conveyearthgo.Message)
		assert.NoError(t, cm.LookupMessages(c.ID, func(m *conveyearthgo.Message) error {
			mmap[m.ID] = m
			return nil
		}))
		assert.Equal(t, 1, len(mmap))
	})
	t.Run("LookupFile", func(t *testing.T) {
		_, err := cm.LookupFile(f2[0].ID)
		assert.Error(t, conveyearthgo.ErrFileNotFound, err)
	})
	t.Run("LookupFiles", func(t *testing.T) {
		fmap := make(map[int64]*conveyearthgo.File)
		assert.NoError(t, cm.LookupFiles(m2.ID, func(f *conveyearthgo.File) error {
			fmap[f.ID] = f
			return nil
		}))
		assert.Equal(t, 0, len(fmap))
	})
}

func TestContentManager_DeleteMessage_WithReply(t *testing.T) {
	db := database.NewInMemory()
	ev := authtest.NewEmailVerifier()
	auth := authgo.NewAuthenticator(db, ev)
	acc := authtest.NewTestAccount(t, auth)
	dir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err)
	fs := filesystem.NewOnDisk(dir)
	defer os.RemoveAll(dir)
	cm := conveyearthgo.NewContentManager(db, fs)
	c, m1, f1 := conveytest.NewConversation(t, cm, acc)
	m2, f2 := conveytest.NewReply(t, cm, acc, c, m1)

	// Cannot delete a message you did not create
	acc2, err := auth.NewAccount("2"+authtest.TEST_EMAIL, authtest.TEST_USERNAME+"2", []byte(authtest.TEST_PASSWORD))
	assert.Nil(t, err)
	err = cm.DeleteMessage(acc2, m1)
	assert.Error(t, conveyearthgo.ErrDeletionNotPermitted, err)

	// Cannot delete a message you created if it has received a reply
	err = cm.DeleteMessage(acc, m1)
	assert.Error(t, conveyearthgo.ErrDeletionNotPermitted, err)
	t.Run("LookupMessage", func(t *testing.T) {
		found, err := cm.LookupMessage(m1.ID)
		assert.NoError(t, err)
		assert.Equal(t, m1.Author, found.Author)
		assert.Equal(t, m1.ConversationID, found.ConversationID)
		assert.Equal(t, m1.ParentID, found.ParentID)
		assert.Equal(t, m1.Cost, found.Cost)
		assert.Equal(t, m1.Yield+m2.Cost/2, found.Yield)
		assert.Equal(t, m1.Created, found.Created)

		found, err = cm.LookupMessage(m2.ID)
		assert.NoError(t, err)
		assert.Equal(t, m2.Author, found.Author)
		assert.Equal(t, m2.ConversationID, found.ConversationID)
		assert.Equal(t, m2.ParentID, found.ParentID)
		assert.Equal(t, m2.Cost, found.Cost)
		assert.Equal(t, m2.Yield, found.Yield)
		assert.Equal(t, m2.Created, found.Created)
	})
	t.Run("LookupMessages", func(t *testing.T) {
		mmap := make(map[int64]*conveyearthgo.Message)
		assert.NoError(t, cm.LookupMessages(c.ID, func(m *conveyearthgo.Message) error {
			mmap[m.ID] = m
			return nil
		}))
		assert.Equal(t, 2, len(mmap))
		found := mmap[m1.ID]
		assert.Equal(t, m1.Author, found.Author)
		assert.Equal(t, m1.ConversationID, found.ConversationID)
		assert.Equal(t, m1.ParentID, found.ParentID)
		assert.Equal(t, m1.Cost, found.Cost)
		assert.Equal(t, m1.Yield+m2.Cost/2, found.Yield)
		assert.Equal(t, m1.Created, found.Created)

		found = mmap[m2.ID]
		assert.Equal(t, m2.Author, found.Author)
		assert.Equal(t, m2.ConversationID, found.ConversationID)
		assert.Equal(t, m2.ParentID, found.ParentID)
		assert.Equal(t, m2.Cost, found.Cost)
		assert.Equal(t, m2.Yield, found.Yield)
		assert.Equal(t, m2.Created, found.Created)
	})
	t.Run("LookupFile", func(t *testing.T) {
		found, err := cm.LookupFile(f1[0].ID)
		assert.NoError(t, err)
		assert.Equal(t, m1.ID, found.Message)
		assert.Equal(t, m1.Created, found.Created)

		found, err = cm.LookupFile(f2[0].ID)
		assert.NoError(t, err)
		assert.Equal(t, m2.ID, found.Message)
		assert.Equal(t, m2.Created, found.Created)
	})
	t.Run("LookupFiles", func(t *testing.T) {
		fmap := make(map[int64]*conveyearthgo.File)
		assert.NoError(t, cm.LookupFiles(m1.ID, func(f *conveyearthgo.File) error {
			fmap[f.ID] = f
			return nil
		}))
		found := fmap[f1[0].ID]
		assert.Equal(t, m1.ID, found.Message)
		assert.Equal(t, m1.Created, found.Created)

		fmap = make(map[int64]*conveyearthgo.File)
		assert.NoError(t, cm.LookupFiles(m2.ID, func(f *conveyearthgo.File) error {
			fmap[f.ID] = f
			return nil
		}))
		found = fmap[f2[0].ID]
		assert.Equal(t, m2.ID, found.Message)
		assert.Equal(t, m2.Created, found.Created)
	})
}

func TestContentManager_DeleteMessage_WithGift(t *testing.T) {
	db := database.NewInMemory()
	ev := authtest.NewEmailVerifier()
	auth := authgo.NewAuthenticator(db, ev)
	acc := authtest.NewTestAccount(t, auth)
	dir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err)
	fs := filesystem.NewOnDisk(dir)
	defer os.RemoveAll(dir)
	cm := conveyearthgo.NewContentManager(db, fs)
	c, m, f := conveytest.NewConversation(t, cm, acc)
	conveytest.NewGift(t, cm, acc, c, m)

	// Cannot delete a message you did not create
	acc2, err := auth.NewAccount("2"+authtest.TEST_EMAIL, authtest.TEST_USERNAME+"2", []byte(authtest.TEST_PASSWORD))
	assert.Nil(t, err)
	err = cm.DeleteMessage(acc2, m)
	assert.Error(t, conveyearthgo.ErrDeletionNotPermitted, err)

	// Cannot delete a message you created if it has received a gift
	err = cm.DeleteMessage(acc, m)
	assert.Error(t, conveyearthgo.ErrDeletionNotPermitted, err)
	t.Run("LookupMessage", func(t *testing.T) {
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
	t.Run("LookupMessages", func(t *testing.T) {
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
	})
	t.Run("LookupFile", func(t *testing.T) {
		found, err := cm.LookupFile(f[0].ID)
		assert.NoError(t, err)
		assert.Equal(t, m.ID, found.Message)
		assert.Equal(t, m.Created, found.Created)
	})
	t.Run("LookupFiles", func(t *testing.T) {
		fmap := make(map[int64]*conveyearthgo.File)
		assert.NoError(t, cm.LookupFiles(m.ID, func(f *conveyearthgo.File) error {
			fmap[f.ID] = f
			return nil
		}))
		found := fmap[f[0].ID]
		assert.Equal(t, m.ID, found.Message)
		assert.Equal(t, m.Created, found.Created)
	})
}

func TestContentManager_NewGift(t *testing.T) {
	db := database.NewInMemory()
	ev := authtest.NewEmailVerifier()
	auth := authgo.NewAuthenticator(db, ev)
	acc := authtest.NewTestAccount(t, auth)
	dir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err)
	fs := filesystem.NewOnDisk(dir)
	defer os.RemoveAll(dir)
	cm := conveyearthgo.NewContentManager(db, fs)
	c, m, _ := conveytest.NewConversation(t, cm, acc)
	g := conveytest.NewGift(t, cm, acc, c, m)
	t.Run("LookupGift", func(t *testing.T) {
		found, err := cm.LookupGift(g.ID)
		assert.NoError(t, err)
		assert.Equal(t, g.ID, found.ID)
		assert.Equal(t, g.Author, found.Author)
		assert.Equal(t, g.ConversationID, found.ConversationID)
		assert.Equal(t, g.MessageID, found.MessageID)
		assert.Equal(t, g.Amount, found.Amount)
		assert.Equal(t, g.Created, found.Created)
	})
	t.Run("LookupGifts", func(t *testing.T) {
		gmap := make(map[int64]*conveyearthgo.Gift)
		assert.NoError(t, cm.LookupGifts(c.ID, m.ID, func(g *conveyearthgo.Gift) error {
			gmap[g.ID] = g
			return nil
		}))
		assert.Equal(t, 1, len(gmap))
		found := gmap[g.ID]
		assert.Equal(t, g.Author, found.Author)
		assert.Equal(t, g.ConversationID, found.ConversationID)
		assert.Equal(t, g.MessageID, found.MessageID)
		assert.Equal(t, g.Amount, found.Amount)
		assert.Equal(t, g.Created, found.Created)
	})
}

func TestContentManager_DeleteGift(t *testing.T) {
	db := database.NewInMemory()
	ev := authtest.NewEmailVerifier()
	auth := authgo.NewAuthenticator(db, ev)
	acc := authtest.NewTestAccount(t, auth)
	dir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err)
	fs := filesystem.NewOnDisk(dir)
	defer os.RemoveAll(dir)
	cm := conveyearthgo.NewContentManager(db, fs)
	c, m, _ := conveytest.NewConversation(t, cm, acc)
	g := conveytest.NewGift(t, cm, acc, c, m)

	// Cannot delete a gift you did not create
	acc2, err := auth.NewAccount("2"+authtest.TEST_EMAIL, authtest.TEST_USERNAME+"2", []byte(authtest.TEST_PASSWORD))
	assert.Nil(t, err)
	err = cm.DeleteGift(acc2, g)
	assert.Error(t, conveyearthgo.ErrDeletionNotPermitted, err)

	// Can delete a gift you created
	err = cm.DeleteGift(acc, g)
	assert.NoError(t, err)
	t.Run("LookupGift", func(t *testing.T) {
		_, err := cm.LookupGift(g.ID)
		assert.Error(t, conveyearthgo.ErrGiftNotFound, err)
	})
	t.Run("LookupGifts", func(t *testing.T) {
		gmap := make(map[int64]*conveyearthgo.Gift)
		assert.NoError(t, cm.LookupGifts(c.ID, m.ID, func(g *conveyearthgo.Gift) error {
			gmap[g.ID] = g
			return nil
		}))
		assert.Equal(t, 0, len(gmap))
	})
}

func TestContentManager_Lookup_Zero(t *testing.T) {
	db := database.NewInMemory()
	dir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err)
	fs := filesystem.NewOnDisk(dir)
	defer os.RemoveAll(dir)
	cm := conveyearthgo.NewContentManager(db, fs)
	_, err = cm.LookupConversation(0)
	assert.Error(t, conveyearthgo.ErrConversationNotFound, err)
	_, err = cm.LookupMessage(0)
	assert.Error(t, conveyearthgo.ErrMessageNotFound, err)
	_, err = cm.LookupFile(0)
	assert.Error(t, conveyearthgo.ErrFileNotFound, err)
	_, err = cm.LookupGift(0)
	assert.Error(t, conveyearthgo.ErrGiftNotFound, err)
}
