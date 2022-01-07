package conveytest

import (
	"aletheiaware.com/authgo"
	"aletheiaware.com/conveyearthgo"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	TEST_TOPIC   = "FooBar"
	TEST_CONTENT = "Hello World!"
	TEST_REPLY   = "Hi!"
)

func NewConversation(t *testing.T, cm conveyearthgo.ContentManager, acc *authgo.Account) (*conveyearthgo.Conversation, *conveyearthgo.Message, []*conveyearthgo.File) {
	t.Helper()
	hash, size, err := cm.AddText([]byte(TEST_CONTENT))
	assert.Nil(t, err)
	c, m, fs, err := cm.NewConversation(acc, TEST_TOPIC, []string{hash}, []string{conveyearthgo.MIME_TEXT_PLAIN}, []int64{size})
	assert.Nil(t, err)
	return c, m, fs
}

func NewReply(t *testing.T, cm conveyearthgo.ContentManager, acc *authgo.Account, c *conveyearthgo.Conversation, m *conveyearthgo.Message) (*conveyearthgo.Message, []*conveyearthgo.File) {
	t.Helper()
	hash, size, err := cm.AddText([]byte(TEST_REPLY))
	assert.NoError(t, err)
	m, fs, err := cm.NewMessage(acc, c.ID, m.ID, []string{hash}, []string{conveyearthgo.MIME_TEXT_PLAIN}, []int64{size})
	assert.NoError(t, err)
	return m, fs
}

func NewGift(t *testing.T, cm conveyearthgo.ContentManager, acc *authgo.Account, c *conveyearthgo.Conversation, m *conveyearthgo.Message) *conveyearthgo.Gift {
	t.Helper()
	g, err := cm.NewGift(acc, c.ID, m.ID, 100)
	assert.NoError(t, err)
	return g
}
