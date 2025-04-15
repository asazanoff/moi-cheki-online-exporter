package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

var mongoCollection *mongo.Collection // Инициализируется в main.go

func handleGenerate(c *gin.Context) {
	dateFrom := c.PostForm("dateFrom")
	dateTo := c.PostForm("dateTo")

	if debugMode {
		log.Printf("DEBUG: Received dates - DateFrom: %s, DateTo: %s", dateFrom, dateTo)
	}

	// Считываем задержку из переменной окружения RECEIVE_DELAY (в секундах)
	receiveDelayStr := os.Getenv("RECEIVE_DELAY")
	receiveDelaySeconds, err := strconv.Atoi(receiveDelayStr)
	if err != nil || receiveDelaySeconds < 0 {
		receiveDelaySeconds = 0 // по умолчанию отключаем задержку
	}
	receiveDelay := time.Duration(receiveDelaySeconds) * time.Second

	// Получаем чеки через API (исходные данные, без трансформации)
	receipts, err := getReceipts(dateFrom, dateTo)
	if err != nil {
		if apiErr, ok := err.(*APIResponseError); ok {
			c.String(apiErr.StatusCode, "%s", apiErr.Message)
			return
		}
		c.String(http.StatusInternalServerError, "Error getting receipts: %v", err)
		return
	}

	if debugMode {
		log.Printf("DEBUG: Received %d receipts", len(receipts))
	}
	if len(receipts) == 0 {
		log.Printf("DEBUG: No receipts found for the given dates")
	}

	var qrCodes []map[string]string
	var transformedReceipts []TransformedReceipt
	var mu sync.Mutex
	var wg sync.WaitGroup
	maxGoroutines := 5
	sem := make(chan struct{}, maxGoroutines)

	for _, receipt := range receipts {
		wg.Add(1)
		sem <- struct{}{}
		go func(receipt Receipt) {
			defer wg.Done()
			defer func() { <-sem }()

			// Сначала пытаемся получить преобразованный чек (transformed receipt) из кеша (MongoDB)
			var transformed TransformedReceipt
			if mongoCollection != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				filter := bson.M{"_id": receipt.Key}
				err := mongoCollection.FindOne(ctx, filter).Decode(&transformed)
				cancel()
				if err == nil {
					if debugMode {
						log.Printf("DEBUG: Fetched from cache: %s", receipt.Key)
					}

					mu.Lock()
					transformedReceipts = append(transformedReceipts, transformed)
					mu.Unlock()

					qrCode, err := generateQRCode(receipt.Key)
					if err != nil {
						log.Printf("Error generating QR code for cached receipt %s: %v", receipt.Key, err)
					} else {
						qrBase64 := base64.StdEncoding.EncodeToString(qrCode)
						mu.Lock()
						qrCodes = append(qrCodes, map[string]string{
							"image": fmt.Sprintf("data:image/png;base64,%s", qrBase64),
							"text":  receipt.Key,
						})
						mu.Unlock()
					}
					return // Чек найден в кеше – дальнейшие запросы не выполняем.
				} else if err != mongo.ErrNoDocuments {
					if debugMode {
						log.Printf("DEBUG: Error fetching from cache for receipt %s: %v", receipt.Key, err)
					}
				} else {
					if debugMode {
						log.Printf("DEBUG: Not found in cache: %s", receipt.Key)
					}
				}
			}

			// Если чек не найден в кеше, добавляем задержку перед запросом к API.
			time.Sleep(receiveDelay)

			// Запрашиваем fiscalData, так как преобразованного чека в кеше нет.
			fiscalData, err := getFiscalData(receipt.Key)
			if err != nil {
				log.Printf("Error getting fiscal data for receipt %s: %v", receipt.Key, err)
				return
			}
			if debugMode {
				log.Printf("DEBUG: Received fiscal data for receipt %s", receipt.Key)
			}

			// Используем receipt.Key в качестве уникального ключа;
			// таким образом qrText = receipt.Key.
			qrText := receipt.Key

			// Преобразуем чек и заполняем необходимые поля
			transformed = TransformedReceipt{
				ID:        qrText,
				CreatedAt: time.Now(),
			}
			transformed.Ticket.Document.Receipt = *fiscalData
			transformed.Ticket.Document.Receipt.TotalSum = RoundToFloat64(fiscalData.TotalSum)
			transformed.Ticket.Document.Receipt.CashTotalSum = RoundToFloat64(fiscalData.CashTotalSum)
			transformed.Ticket.Document.Receipt.ECashTotalSum = RoundToFloat64(fiscalData.ECashTotalSum)
			transformed.Ticket.Document.Receipt.CreditSum = RoundToFloat64(fiscalData.CreditSum)
			transformed.Ticket.Document.Receipt.PrepaidSum = RoundToFloat64(fiscalData.PrepaidSum)
			transformed.Ticket.Document.Receipt.ProvisionSum = RoundToFloat64(fiscalData.ProvisionSum)
			transformed.Ticket.Document.Receipt.NDS10 = RoundToFloat64(fiscalData.NDS10)
			transformed.Ticket.Document.Receipt.NDS18 = RoundToFloat64(fiscalData.NDS18)
			for i := range transformed.Ticket.Document.Receipt.Items {
				transformed.Ticket.Document.Receipt.Items[i].Sum = RoundToFloat64(transformed.Ticket.Document.Receipt.Items[i].Sum)
				transformed.Ticket.Document.Receipt.Items[i].Price = RoundToFloat64(transformed.Ticket.Document.Receipt.Items[i].Price)
			}

			// Если MongoDB подключена – сохраняем (upsert) преобразованный чек в кеш
			if mongoCollection != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				filter := bson.M{"_id": qrText}
				update := bson.M{"$set": transformed}
				opts := options.Update().SetUpsert(true)
				if _, err := mongoCollection.UpdateOne(ctx, filter, update, opts); err != nil {
					log.Printf("Error upserting transformed receipt with key %s: %v", qrText, err)
				}
				cancel()
			}

			mu.Lock()
			transformedReceipts = append(transformedReceipts, transformed)
			mu.Unlock()

			// Генерация QR-кода для преобразованного чека
			qrCode, err := generateQRCode(qrText)
			if err != nil {
				log.Printf("Error generating QR code for receipt %s: %v", receipt.Key, err)
				return
			}
			if debugMode {
				log.Printf("DEBUG: Generated QR code for receipt %s", receipt.Key)
			}
			qrBase64 := base64.StdEncoding.EncodeToString(qrCode)
			mu.Lock()
			qrCodes = append(qrCodes, map[string]string{
				"image": fmt.Sprintf("data:image/png;base64,%s", qrBase64),
				"text":  qrText,
			})
			mu.Unlock()
		}(receipt)
	}

	wg.Wait()

	if debugMode {
		log.Printf("DEBUG: Generated %d QR codes", len(qrCodes))
	}

	transformedReceiptsJSON, err := json.MarshalIndent(transformedReceipts, "", "    ")
	if err != nil {
		c.String(http.StatusInternalServerError, "Error marshaling transformed receipts: %v", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"qrCodes":             qrCodes,
		"transformedReceipts": string(transformedReceiptsJSON),
	})
}
