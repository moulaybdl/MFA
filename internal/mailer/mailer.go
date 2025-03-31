package mailer

import (
	"bytes"
	"embed"
	"text/template"
	"time"

	"github.com/go-mail/mail/v2"
)

//go:embed "templates"
var templatesFS embed.FS

type Mailer struct {
	dialer *mail.Dialer // this is used to connect to the SMTP server
	sender string // this is used to store sender information "Moulay Bouabdelli <moualy@gmail.com>"
}

func New(host string, port int, username, passoword, sender string) Mailer {
	dialer := mail.NewDialer(host, port, username, passoword)
	dialer.Timeout = time.Second * 5

	return Mailer{
		dialer: dialer,
		sender: sender,
	}
}

func (m Mailer) Send(receipent, templateFile string, data interface{}) error {
	tmpl, err := template.New("email").ParseFS(templatesFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}
	
	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}


	msg := mail.NewMessage()
	msg.SetHeader("To", receipent)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())

	err = m.dialer.DialAndSend(msg)
	if err != nil {
		return err
	}

	return nil

}