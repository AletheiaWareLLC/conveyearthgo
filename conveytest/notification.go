package conveytest

import (
	"aletheiaware.com/authgo"
	"aletheiaware.com/conveyearthgo"
	"log"
)

func NewNotificationSender() conveyearthgo.NotificationSender {
	return &notificationSender{}
}

type notificationSender struct{}

func (s *notificationSender) SendResponseNotification(account *authgo.Account, responder, topic string, conversation, message int64) error {
	log.Println("Response Notification", account.Email, account.Username, responder, topic, conversation, message)
	return nil
}

func (s *notificationSender) SendMentionNotification(account *authgo.Account, mentioner, topic string, conversation, message int64) error {
	log.Println("Mention Notification", account.Email, account.Username, mentioner, topic, conversation, message)
	return nil
}

func (s *notificationSender) SendGiftNotification(account *authgo.Account, gifter, topic string, conversation, message, amount int64) error {
	log.Println("Gift Notification", account.Email, account.Username, gifter, topic, conversation, message, amount)
	return nil
}
