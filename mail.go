package argus

import (
	"github.com/jordan-wright/email"
	"net/smtp"
	"strconv"
)

type mailMessage struct {
	from       string
	to         []string
	subject    string
	body       string
	attachment string
}

type mailSender interface {
	send(msg mailMessage) error
}

func newMailSender(config Configuration) mailSender {
	return &defaultMailSender{
		serverHost:     config.MailConfig.ServerHost,
		serverPort:     config.MailConfig.ServerPort,
		serverUser:     config.MailConfig.ServerUser,
		serverPassword: config.MailConfig.ServerPassword}
}

type defaultMailSender struct {
	serverHost     string
	serverPort     int
	serverUser     string
	serverPassword string
}

func (m *defaultMailSender) send(msg mailMessage) error {
	e := email.NewEmail()
	e.From = msg.from
	e.To = msg.to
	e.Subject = msg.subject
	e.Text = []byte(msg.body)
	if _, err := e.AttachFile(msg.attachment); err != nil {
		return err
	}
	serverHostPort := m.serverHost + ":" + strconv.Itoa(m.serverPort)
	auth := smtp.PlainAuth("", m.serverUser, m.serverPassword, m.serverHost)
	return e.Send(serverHostPort, auth)
}
