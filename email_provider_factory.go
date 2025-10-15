package email

import (
	"errors"
	"time"

	"github.com/hivemindd/communication-service/internal/form"

	"github.com/hashicorp/go-retryablehttp"
)

func EmailProviderFactory(config form.EmailConfig) (EmailProvider, error) {
	switch config.Provider {
	case "sendgrid":
		return NewSendGrid(config), nil
	case "smtp2go":
		c := retryablehttp.NewClient()
		c.RetryWaitMin = time.Second
		c.RetryWaitMax = 3 * time.Second
		c.CheckRetry = retryablehttp.DefaultRetryPolicy
		c.RetryMax = 3

		return NewSmtp2Go(config, c), nil
	default:
		return nil, errors.New("invalid email provider specified")
	}
}
