package conveytest

import (
	"aletheiaware.com/conveyearthgo"
	"log"
)

func NewNotificationSender() conveyearthgo.NotificationSender {
	return &notificationSender{}
}

type notificationSender struct{}

func (s *notificationSender) SendResponseNotification(email, responder, topic string, conversation, message int64) error {
	log.Println("Response Notification", email, responder, topic, conversation, message)
	return nil
}

func (s *notificationSender) SendMentionNotification(email, mentioner, topic string, conversation, message int64) error {
	log.Println("Mention Notification", email, mentioner, topic, conversation, message)
	return nil
}
