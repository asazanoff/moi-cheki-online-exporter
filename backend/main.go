package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"os"
)

// initMongoDB пытается установить соединение с MongoDB по строке подключения из переменной окружения MONGO_URI.
func initMongoDB() (*mongo.Client, *mongo.Collection, error) {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Println("MONGO_URI не задана, кэширование отключено.")
		return nil, nil, nil
	}

	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Printf("Ошибка подключения к MongoDB: %v", err)
		return nil, nil, err
	}

	// Проверяем соединение (ping)
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Printf("Ошибка ping MongoDB: %v", err)
		return nil, nil, err
	}

	collection := client.Database("receiptsdb").Collection("receipts")
	log.Println("Успешное подключение к MongoDB!")
	return client, collection, nil
}

func main() {
	// Инициализируем подключение к MongoDB
	client, collection, err := initMongoDB()
	if client != nil {
		defer client.Disconnect(context.Background())
	}
	if err != nil {
		log.Printf("MongoDB не доступна: %v", err)
	}

	// Используем глобальную переменную mongoCollection, которая уже объявлена, например, в handlers.go
	mongoCollection = collection

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
