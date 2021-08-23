package conveyearthgo

import "errors"

const (
	MINIMUM_TOPIC_LENGTH = 1
	MAXIMUM_TOPIC_LENGTH = 100
)

var (
	ErrTopicTooShort = errors.New("Topic Too Short")
	ErrTopicTooLong  = errors.New("Topic Too Long")
)

func ValidateTopic(topic string) error {
	// TODO ensure topic has no newline characters
	length := len(topic)
	if length < MINIMUM_TOPIC_LENGTH {
		return ErrTopicTooShort
	}
	if length > MAXIMUM_TOPIC_LENGTH {
		return ErrTopicTooLong
	}
	return nil
}
