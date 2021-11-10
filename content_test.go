package conveyearthgo_test

import (
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

// TODO Test NewConversation
// TODO Test LookupConversation
// TODO Test LookupBestConversations
// TODO Test LookupRecentConversations
// TODO Test NewMessage
// TODO Test LookupMessage
// TODO Test LookupMessages
