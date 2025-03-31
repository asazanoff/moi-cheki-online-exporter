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
	"sync"
	"time"
)

// Предположим, что mongoCollection имеет тип *mongo.Collection и инициализируется в main.go
var mongoCollection *mongo.Collection

func handleGenerate(c *gin.Context) {
	dateFrom := c.PostForm("dateFrom")
	dateTo := c.PostForm("dateTo")

	if debugMode {
		log.Printf("DEBUG: Received dates - DateFrom: %s, DateTo: %s", dateFrom, dateTo)
	}

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
			// Задержка перед следующим запросом
			time.Sleep(2 * time.Second) // Задержка 2 секунды
			defer wg.Done()
			defer func() { <-sem }()

			// Сначала запрашиваем fiscalData (независимо от наличия кэша)
			fiscalData, err := getFiscalData(receipt.Key)
			if err != nil {
				log.Printf("Error getting fiscal data for receipt %s: %v", receipt.Key, err)
				return
			}
			if debugMode {
				log.Printf("DEBUG: Received fiscal data for receipt %s", receipt.Key)
			}

			// Форматируем дату и генерируем qrText на основе fiscalData
			formattedDateTime := formatDateTime(fiscalData.DateTime)
			qrText := fmt.Sprintf("t=%s&s=%.2f&fn=%s&i=%d&fp=%s&n=%d",
				formattedDateTime,
				fiscalData.TotalSum,
				fiscalData.FiscalDriveNumber,
				fiscalData.FiscalDocumentNumber,
				fiscalData.FiscalSign,
				fiscalData.OperationType)

			var transformed TransformedReceipt

			if mongoCollection != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				filter := bson.M{"_id": qrText}
				var transformed TransformedReceipt
				err := mongoCollection.FindOne(ctx, filter).Decode(&transformed)
				cancel()
				if err == nil {
					log.Printf("DEBUG: Fetched from cache: %s", qrText)
				} else if err == mongo.ErrNoDocuments {
					log.Printf("DEBUG: Not found in cache: %s", qrText)
				} else {
					log.Printf("DEBUG: Error fetching from cache: %v", err)
				}
			}

			// Если кэш не найден (или ошибка – mongo.ErrNoDocuments), то выполняем преобразование
			if transformed.ID == "" {
				transformed = TransformedReceipt{
					ID:        qrText,
					CreatedAt: time.Now(),
				}
				// Заполняем структуру используя данные fiscalData и округление
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

				// Если MongoDB подключена — сохраняем (upsert) преобразованный чек по qrText
				if mongoCollection != nil {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					filter := bson.M{"_id": qrText}
					update := bson.M{"$set": transformed}
					opts := options.Update().SetUpsert(true)
					if _, err := mongoCollection.UpdateOne(ctx, filter, update, opts); err != nil {
						log.Printf("Error upserting transformed receipt with qrText %s: %v", qrText, err)
					}
					cancel()
				}
			}

			// Добавляем преобразованный чек в итоговый срез
			mu.Lock()
			transformedReceipts = append(transformedReceipts, transformed)
			mu.Unlock()

			// Генерируем QR-код используя сформированный qrText
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
