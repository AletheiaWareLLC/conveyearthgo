package conveyearthgo

import (
	"aletheiaware.com/authgo"
	authemail "aletheiaware.com/authgo/email"
	"fmt"
	"html/template"
	"log"
)

type NotificationDatabase interface {
	SelectNotificationPreferences(int64) (int64, bool, bool, bool, error)
	UpdateNotificationPreferences(int64, int64, bool, bool, bool) (int64, error)
}

type NotificationManager interface {
	NotificationPreferences(int64) (int64, bool, bool, bool, error)
	SetNotificationPreferences(int64, int64, bool, bool, bool) error
	NotifyResponse(*authgo.Account, *authgo.Account, int64, string, int64) error
	NotifyMention(*authgo.Account, *authgo.Account, int64, string, int64) error
}

type NotificationSender interface {
	SendResponseNotification(string, string, string, int64, int64) error
	SendMentionNotification(string, string, string, int64, int64) error
}

func NewNotificationManager(db NotificationDatabase, sender NotificationSender) NotificationManager {
	return &notificationManager{
		database: db,
		sender:   sender,
	}
}

type notificationManager struct {
	database NotificationDatabase
	sender   NotificationSender
}

func (m *notificationManager) NotificationPreferences(user int64) (int64, bool, bool, bool, error) {
	return m.database.SelectNotificationPreferences(user)
}

func (m *notificationManager) SetNotificationPreferences(id, user int64, responses, mentions, digests bool) error {
	_, err := m.database.UpdateNotificationPreferences(id, user, responses, mentions, digests)
	return err
}

func (m *notificationManager) NotifyResponse(author, responder *authgo.Account, conversation int64, topic string, message int64) error {
	_, responses, _, _, err := m.database.SelectNotificationPreferences(author.ID)
	if err != nil {
		return err
	}
	if !responses {
		// User disabled reponse notifications
		return nil
	}
	return m.sender.SendResponseNotification(author.Email, responder.Username, topic, conversation, message)
}

func (m *notificationManager) NotifyMention(author, mentioner *authgo.Account, conversation int64, topic string, message int64) error {
	_, _, mentions, _, err := m.database.SelectNotificationPreferences(author.ID)
	if err != nil {
		return err
	}
	if !mentions {
		// User disabled mention notifications
		return nil
	}
	return m.sender.SendMentionNotification(author.Email, mentioner.Username, topic, conversation, message)
}

func NewSmtpNotificationSender(scheme, host, server, identity, sender string, templates *template.Template) NotificationSender {
	return &smtpNotificationSender{
		scheme:    scheme,
		host:      host,
		server:    server,
		identity:  identity,
		sender:    sender,
		templates: templates,
	}
}

type smtpNotificationSender struct {
	scheme,
	host,
	server,
	identity,
	sender string
	templates *template.Template
}

func (s *smtpNotificationSender) SendResponseNotification(email, responder, topic string, conversation, message int64) error {
	log.Println("Notifying", email, "of response")
	data := struct {
		From      string
		To        string
		Topic     string
		Responder string
		Link      string
	}{
		From:      s.sender,
		To:        email,
		Topic:     topic,
		Responder: responder,
		Link:      createLink(s.scheme, s.host, conversation, message),
	}
	return authemail.SendEmail(s.server, s.identity, s.sender, email, s.templates.Lookup("email-notification-response.go.html"), data)
}

func (s *smtpNotificationSender) SendMentionNotification(email, mentioner, topic string, conversation, message int64) error {
	log.Println("Notifying", email, "of mention")
	data := struct {
		From      string
		To        string
		Topic     string
		Mentioner string
		Link      string
	}{
		From:      s.sender,
		To:        email,
		Topic:     topic,
		Mentioner: mentioner,
		Link:      createLink(s.scheme, s.host, conversation, message),
	}
	return authemail.SendEmail(s.server, s.identity, s.sender, email, s.templates.Lookup("email-notification-mention.go.html"), data)
}

func createLink(scheme, host string, conversation, message int64) string {
	if message == 0 {
		return fmt.Sprintf("%s://%s/conversation?id=%d", scheme, host, conversation)
	}
	return fmt.Sprintf("%s://%s/conversation?id=%d#message%d", scheme, host, conversation, message)
}