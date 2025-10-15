package email

import (
	"bytes"
	"fmt"
	"html/template"
	"log"

	"github.com/hivemindd/communication-service/internal/form"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type sendGrid struct {
	config form.EmailConfig
}

func NewSendGrid(config form.EmailConfig) EmailProvider {
	return sendGrid{
		config: config,
	}
}

func (s sendGrid) Send(email form.Email, template *template.Template) error {
	from := mail.NewEmail(fmt.Sprintf("No-reply (%s)", s.config.FromName), s.config.From)
	subject := email.Subject
	to := mail.NewEmail(email.Name, email.To)

	var tpl bytes.Buffer
	if err := template.Execute(&tpl, email); err != nil {
		log.Println("Template error: ", err)
	}
	htmlContent := tpl.String()

	mailSettings := mail.NewMailSettings()

	message := mail.NewSingleEmail(from, subject, to, "", htmlContent)
	message.SetMailSettings(mailSettings)

	client := sendgrid.NewSendClient(s.config.Key)
	_, err := client.Send(message)

	return err
}
