package config

type Config struct {
	RabbitURI  string `envconfig:"RABBIT_URI" required:"true"`
	ServerPort string `envconfig:"COMMUNICATION_SERVICE_SERVER_PORT"`
	SentryDSN  string `envconfig:"SENTRY_DSN"`
	JaegerHost string `envconfig:"JAEGER_HOST"`

	EmailProvider  string `envconfig:"EMAIL_PROVIDER" required:"true"`
	EmailFrom      string `envconfig:"EMAIL_FROM" required:"true"`
	EmailSandbox   bool   `envconfig:"EMAIL_SANDBOX" required:"true"`
	EmailKey       string `envconfig:"EMAIL_KEY" required:"true"`
	EmailServerUrl string `envconfig:"EMAIL_URL" required:"true"`
	EmailName      string `envconfig:"EMAIL_NAME" required:"true"`

	UIBaseURL string `envconfig:"UI_BASE_URL" required:"false"`
}
