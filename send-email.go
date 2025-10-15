package main

import (
	"context"
	"encoding/json"
	"html/template"
	"time"

	"github.com/rabbitmq/amqp091-go"

	"github.com/hivemindd/communication-service/config"
	"github.com/hivemindd/communication-service/internal/email"
	"github.com/hivemindd/communication-service/internal/form"

	"github.com/hivemindd/kit/docid"
	"github.com/hivemindd/kit/queue"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

/**
* connect queue
* call queue consumer and start listening for new messages
* configure email templates
* configure email provider
* instantiate email provider instance to send email
* initiate goroutine to process incoming messages
* send email
 */
func sendEmailsFromQueue(ctx context.Context, conf *config.Config, q *queue.RabbitQueue, logger *zap.SugaredLogger) {
	// call queue consumer and start listening for new messages
	msgs := q.Consume("send_email", false)
	logger.Info("Waiting for messages, to exit press Ctrl+C")

	// configure email provider
	cfg := form.EmailConfig{
		Provider:  conf.EmailProvider,
		FromName:  conf.EmailName,
		From:      conf.EmailFrom,
		Sandbox:   conf.EmailSandbox,
		Key:       conf.EmailKey,
		ServerUrl: conf.EmailServerUrl,
	}

	// instantiate email provider instance to send email
	var emailProvider email.EmailProvider
	emailProvider, err := email.EmailProviderFactory(cfg)
	if err != nil {
		logger.Fatalf("Failed to setup email provider: %v", err.Error())
	}

	var templateMap = make(map[string]*template.Template)
	for _, tt := range templates {
		tmpl, err := template.ParseFiles(tt.Path)
		if err != nil {
			logger.Fatalf("Failed to parse templates: %q, error: %v", tt.Path, err.Error())
		}
		templateMap[tt.Name] = tmpl
	}

	go func() {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		for {
			for d := range msgs {
				handleMsg(ctx, conf, &amqpDelivery{d}, logger, emailProvider, templateMap)
			}

			logger.Info("Connection closed")

			if err := q.Retry(ctx, conf.RabbitURI, time.Second); err != nil {
				cancel()
				break
			}

			msgs = q.Consume("send_email", false)
		}
	}()
}

// Define Delivery interface
type Delivery interface {
	Ack(multiple bool) error
	Nack(multiple, requeue bool) error
	Reject(requeue bool) error
	GetBody() []byte
}

// amqpDelivery wraps amqp091.Delivery to implement Delivery
type amqpDelivery struct {
	amqp091.Delivery
}

func (d *amqpDelivery) GetBody() []byte {
	return d.Body
}

func handleMsg(ctx context.Context, conf *config.Config, d Delivery, logger *zap.SugaredLogger,
	emailProvider email.EmailProvider, templateMap map[string]*template.Template) {

	msgID := docid.New()
	logger.Infof("Received a message: %s", msgID)

	tracer := otel.Tracer("sendEmails")
	_, span := tracer.Start(ctx, "sendEmailsFromQueue", trace.WithAttributes(attribute.String("msgid", msgID)))
	defer span.End()
	// NOTE: 'span' has to be closed by calling .End() on it at the end of email processing.

	var email form.Email
	if err := json.Unmarshal(d.GetBody(), &email); err != nil {
		logger.Infof("logging incoming email msg upon unmarshal error", string(d.GetBody()))
		logger.Errorf("unable to unmarshall  incoming email message, %v", err.Error())
		// discard message on queue and prevent replay
		d.Reject(false)
		return
	}
	logger.Infof("Email: %+v", email)

	tpl := findTemplateByType(email.Type)
	if tpl.IsAbsent() {
		logger.Errorf("Error missing templates with ID: %v", email.Type)
		d.Reject(false)
		return
	}

	email.Subject = tpl.MustGet().Title
	email.BaseUrl = conf.UIBaseURL

	_, spanq := tracer.Start(ctx, "Send", trace.WithAttributes(attribute.String("msgid", msgID)))
	defer spanq.End()

	// send email
	err := emailProvider.Send(email, templateMap[email.Type])
	if err != nil {
		d.Nack(false, false)
		logger.Errorf("Error sending email: %v", err)
	} else {
		d.Ack(false)
		logger.Infof("Email sent %s, type %s, email %s", msgID, email.Type, email.To)
	}
}
