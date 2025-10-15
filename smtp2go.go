package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/hivemindd/communication-service/internal/form"

	"github.com/hashicorp/go-retryablehttp"
)

type smtp2Go struct {
	config form.EmailConfig
	client *retryablehttp.Client
}

func NewSmtp2Go(config form.EmailConfig, client *retryablehttp.Client) EmailProvider {
	return smtp2Go{
		config: config,
		client: client,
	}
}

type payload struct {
	ApiKey        string         `json:"api_key"`
	To            []string       `json:"to"`
	Sender        string         `json:"sender"`
	Subject       string         `json:"subject"`
	HtmlBody      string         `json:"html_body"`
	CustomHeaders []customHeader `json:"custom_headers"`
}

type customHeader struct {
	Header string `json:"header"`
	Value  string `json:"value"`
}

func (s smtp2Go) Send(email form.Email, template *template.Template) error {

	var tpl bytes.Buffer
	if err := template.Execute(&tpl, email); err != nil {
		return err
	}
	htmlContent := tpl.String()

	customHeaders := []customHeader{
		{Header: "No-reply",
			Value: fmt.Sprintf("No-reply <%s>", s.config.From),
		},
	}

	toList := []string{email.To}

	pl := payload{
		ApiKey:        s.config.Key,
		Sender:        fmt.Sprintf("No-reply %s <%s>", s.config.FromName, s.config.From),
		To:            toList,
		Subject:       email.Subject,
		HtmlBody:      htmlContent,
		CustomHeaders: customHeaders,
	}
	bodyStr, err := json.Marshal(pl)
	if err != nil {
		return err
	}

	req, err := retryablehttp.NewRequest(http.MethodPost, s.config.ServerUrl, bodyStr)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unable to send email to smtp2go, %s, status %s", s.config.ServerUrl, resp.Status)
	}

	return err
}
