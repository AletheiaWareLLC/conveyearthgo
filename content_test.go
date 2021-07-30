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
