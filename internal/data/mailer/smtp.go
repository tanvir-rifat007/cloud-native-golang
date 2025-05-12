package mailer

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
	"strings"
)

// type Mailer struct {
// 	host     string
// 	port     string
// 	username string
// 	password string
// 	from     string
// }

// func NewMailer(host, port, username, password, from string) *Mailer {
// 	return &Mailer{host, port, username, password, from}
// }
func Send(to []string, subject string, templateFile string, data any) error {
	auth := smtp.PlainAuth("", os.Getenv("FROM_EMAIL"), os.Getenv("FROM_EMAIL_PASSWORD"), os.Getenv("FROM_EMAIL_SMTP"))

	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, data)
	if err != nil {
		return err
	}

	headers := map[string]string{
		"From":         os.Getenv("FROM_EMAIL"),
		"To":           strings.Join(to, ","),
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/html; charset=\"UTF-8\"",
	}

	var message strings.Builder
	for k, v := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	message.WriteString("\r\n" + buf.String())

	return smtp.SendMail(
		os.Getenv("SMTP_ADDR"),
		auth,
		os.Getenv("FROM_EMAIL"),
		to,
		[]byte(message.String()),
	)
}


