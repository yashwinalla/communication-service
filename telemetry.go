package main

import (
	"context"

	"github.com/hivemindd/kit/env"
	"github.com/hivemindd/kit/telemetry"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
)

func newTracerProvider(serviceName string, logger *zap.SugaredLogger) (*tracesdk.TracerProvider, func()) {
	tp, err := telemetry.NewTracerProvider(serviceName, string(env.Get()), telemetry.DefaultURL())
	if err != nil {
		logger.Errorf("Unable to connect to trace provider for telemetry: %v", err)
	}

	return tp, func() {
		if tp != nil {
			if err := tp.Shutdown(context.Background()); err != nil {
				logger.Errorf("Unable to shutdown trace provider for telemetry: %v", err)
			}
		}
	}
}
