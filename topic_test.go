package conveyearthgo_test

import (
	"aletheiaware.com/conveyearthgo"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func Test_ValidateTopic(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		assert.NoError(t, conveyearthgo.ValidateTopic("Test"))
	})
	t.Run("Short", func(t *testing.T) {
		topic := strings.Repeat("x", conveyearthgo.MINIMUM_TOPIC_LENGTH-1)
		assert.Error(t, conveyearthgo.ErrTopicTooShort, conveyearthgo.ValidateTopic(topic))
	})
	t.Run("Long", func(t *testing.T) {
		topic := strings.Repeat("x", conveyearthgo.MAXIMUM_TOPIC_LENGTH+1)
		assert.Error(t, conveyearthgo.ErrTopicTooLong, conveyearthgo.ValidateTopic(topic))
	})
}
