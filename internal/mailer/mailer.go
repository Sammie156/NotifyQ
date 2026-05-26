package mailer

import (
	"fmt"
	"net/smtp"

	"github.com/Sammie156/NotifyQ/internal/job"
)

type Mailer struct {
	Host     string
	Port     int
	Username string
	Password string
}

func NewMailer(host string, port int, username string, password string) *Mailer {
	return &Mailer{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
	}
}

func (m *Mailer) SendJob(j *job.Job) error {
	message := []byte(
		"From: notifyq@test.com\r\n" +
			"To: " + j.Recipient + "\r\n" +
			"Subject: " + j.Subject + "\r\n" +
			"\r\n" +
			j.Body + "\r\n",
	)

	addr := fmt.Sprintf("%s:%d", m.Host, m.Port)

	auth := smtp.PlainAuth(
		"",
		m.Username,
		m.Password,
		m.Host,
	)

	err := smtp.SendMail(
		addr,
		auth,
		"notifyq@test.com",
		[]string{j.Recipient},
		message,
	)

	if err != nil {
		return fmt.Errorf("mailer: Error to send mail to %s: %w", j.Recipient, err)
	}

	return nil
}
