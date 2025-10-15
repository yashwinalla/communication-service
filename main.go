package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/hivemindd/communication-service/config"

	"github.com/gin-gonic/gin"

	"github.com/hivemindd/kit/env"
	"github.com/hivemindd/kit/queue"
	"github.com/hivemindd/kit/sentry"
	"github.com/hivemindd/kit/zaplog"

	"go.uber.org/zap"
)

// version is injected via makefile.
var version string

const serviceName = "communication-service"

func main() {
	flag.Parse()

	// IMPORTANT: default environment variable ENV has to be set via Makefile with values: dev, qa, prod.
	environ := os.Getenv("ENV")
	if environ == "" {
		panic("Failed to get environment variable ENV. Make sure it is set.")
	}
	// Used only when local env needs to load secrets from GCP Secret Manager
	// env.SetProjectID(os.Getenv("PROJECT_ID"))
	// env.SetCredentialsFile(os.Getenv("SECRET_CREDENTIALS_PATH"))

	var conf config.Config
	if err := env.Load(&conf); err != nil {
		panic("Failed to load environment variables:" + err.Error())
	}
	conf.RabbitURI = strings.Trim(conf.RabbitURI, "'")
	if !strings.HasPrefix(conf.ServerPort, ":") {
		conf.ServerPort = ":" + conf.ServerPort
	}

	// Setup zap logger.
	// Sentry DSN can be easily revoked and recreated inside Sentry UI.
	lg := zaplog.Setup(serviceName)
	logger := sentry.AttachLogger(lg, conf.SentryDSN, serviceName).Sugar()
	defer logger.Sync()

	startService(context.Background(), &conf, logger)
}

// startService sets up logging, connects to external databases, starts http server.
func startService(ctx context.Context, conf *config.Config, logger *zap.SugaredLogger) {
	// IMPORTANT: default environment variable ENV is set via Makefile with values: dev, stg, prod.
	logger.Infof("Starting %s in %q. Version: %s", serviceName, env.Get(), version)

	// connect queue
	q := queue.Connect(conf.RabbitURI)
	logger.Info("Connected to queue")

	// Telemetry
	tp, shutdown := newTracerProvider(serviceName, logger)
	defer shutdown()

	// initiate goroutine to process incoming emails
	sendEmailsFromQueue(ctx, conf, q, logger)

	var srv Service
	router := setupRouter(&srv, serviceName, tp)
	listenAndServe(router, conf.ServerPort, logger)
}

func listenAndServe(router *gin.Engine, port string, logger *zap.SugaredLogger) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Addr:    port,
		Handler: router,
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("listen: %s\n", err)
		}
	}()

	// Listen for the interrupt signal.
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	logger.Info("Shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exiting")
}
