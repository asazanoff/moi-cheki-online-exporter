package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.Use(checkTokenExpiration())

	router.POST("/generate", handleGenerate)

	router.Run(":8080")
}
