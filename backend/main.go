package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	router := gin.Default()

	router.GET("/health/live", func(c *gin.Context) {
		c.String(http.StatusOK, "Live check passed")
	})

	router.GET("/health/ready", func(c *gin.Context) {
		ready := true

		if ready {
			c.String(http.StatusOK, "Ready check passed")
		} else {
			c.String(http.StatusServiceUnavailable, "Ready check failed")
		}
	})

	router.Use(checkTokenExpiration())

	router.POST("/generate", handleGenerate)

	router.Run(":8080")
}
