package main

import (
	"github.com/gin-gonic/gin"
	"github.com/hivemindd/kit/telemetry"
	"go.opentelemetry.io/otel/sdk/trace"
)

var excludeTrace = []string{"/health"}

func setupRouter(svr *Service, serviceName string, tp *trace.TracerProvider) *gin.Engine {
	// Create service instance. Main instance shared by all http handlers.
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Telemetry
	if tp != nil {
		router.Use(telemetry.Trace(serviceName, tp, nil, excludeTrace))
	}
	router.GET("/communications/health", svr.Health)
	return router
}
