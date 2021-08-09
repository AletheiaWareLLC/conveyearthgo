package conveyearthgo_test

import (
	"aletheiaware.com/conveyearthgo"
	"aletheiaware.com/conveyearthgo/database"
	"aletheiaware.com/conveyearthgo/filesystem"
	"github.com/stretchr/testify/assert"
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
			assert.Equal(t, len(tt.expected), len(mentions))
			assert.Equal(t, tt.expected, mentions)
		})
	}
}

func TestContentManager_AddText(t *testing.T) {
	db := database.NewInMemory()
	dir, err := ioutil.TempDir("", "test")
	assert.Nil(t, err)
	fs := filesystem.NewOnDisk(dir)
	defer os.RemoveAll(dir)
	cm := conveyearthgo.NewContentManager(db, fs)
	contents := []byte("this is a test")
	hash, size, err := cm.AddText(contents)
	assert.Nil(t, err)
	assert.NotEmpty(t, hash)
	assert.Equal(t, int64(len(contents)), size)
}

func TestContentManager_AddFile(t *testing.T) {
	db := database.NewInMemory()
	dir, err := ioutil.TempDir("", "test")
	assert.Nil(t, err)
	fs := filesystem.NewOnDisk(dir)
	defer os.RemoveAll(dir)
	cm := conveyearthgo.NewContentManager(db, fs)
	contents := "this is a test"

	hash, size, err := cm.AddFile(strings.NewReader(contents))
	assert.Nil(t, err)
	assert.NotEmpty(t, hash)
	assert.Equal(t, int64(len([]byte(contents))), size)
}

func TestContentManager_Open(t *testing.T) {
	db := database.NewInMemory()
	dir, err := ioutil.TempDir("", "test")
	assert.Nil(t, err)
	fs := filesystem.NewOnDisk(dir)
	defer os.RemoveAll(dir)
	cm := conveyearthgo.NewContentManager(db, fs)
	contents := []byte("this is a test")
	hash, size, err := cm.AddText(contents)
	assert.Nil(t, err)
	assert.Equal(t, int64(len(contents)), size)
	t.Run("Exists", func(t *testing.T) {
		file, err := cm.Open(hash)
		assert.Nil(t, err)
		data, err := io.ReadAll(file)
		assert.Nil(t, err)
		assert.Equal(t, contents, data)
	})
	t.Run("Not Exists", func(t *testing.T) {
		file, err := cm.Open("hash1234")
		assert.NotNil(t, err)
		assert.Nil(t, file)
	})
}

// TODO Test NewConversation
// TODO Test LookupConversation
// TODO Test LookupBestConversations
// TODO Test LookupRecentConversations
// TODO Test NewMessage
// TODO Test LookupMessage
// TODO Test LookupMessages
