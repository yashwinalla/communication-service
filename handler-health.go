package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (svr *Service) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "ok",
		"error":   false,
	})
}
