package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"sync"
	"time"
)

func handleGenerate(c *gin.Context) {
	dateFrom := c.PostForm("dateFrom")
	dateTo := c.PostForm("dateTo")

	if debugMode {
		log.Printf("DEBUG: Received dates - DateFrom: %s, DateTo: %s", dateFrom, dateTo)
	}

	receipts, err := getReceipts(dateFrom, dateTo)
	if err != nil {
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

			fiscalData, err := getFiscalData(receipt.Key)
			if err != nil {
				log.Printf("Error getting fiscal data for receipt %s: %v", receipt.Key, err)
				return
			}

			if debugMode {
				log.Printf("DEBUG: Received fiscal data for receipt %s", receipt.Key)
			}

			transformedReceipt := TransformedReceipt{
				ID:        generateRandomID(),
				CreatedAt: time.Now(),
			}
			transformedReceipt.Ticket.Document.Receipt = *fiscalData
			transformedReceipt.Ticket.Document.Receipt.TotalSum = float64(int(fiscalData.TotalSum * 100))
			transformedReceipt.Ticket.Document.Receipt.CashTotalSum = float64(int(fiscalData.CashTotalSum * 100))
			transformedReceipt.Ticket.Document.Receipt.ECashTotalSum = float64(int(fiscalData.ECashTotalSum * 100))
			transformedReceipt.Ticket.Document.Receipt.CreditSum = float64(int(fiscalData.CreditSum * 100))
			transformedReceipt.Ticket.Document.Receipt.PrepaidSum = float64(int(fiscalData.PrepaidSum * 100))
			transformedReceipt.Ticket.Document.Receipt.ProvisionSum = float64(int(fiscalData.ProvisionSum * 100))
			transformedReceipt.Ticket.Document.Receipt.NDS10 = float64(int(fiscalData.NDS10 * 100))
			transformedReceipt.Ticket.Document.Receipt.NDS18 = float64(int(fiscalData.NDS18 * 100))

			for i := range transformedReceipt.Ticket.Document.Receipt.Items {
				transformedReceipt.Ticket.Document.Receipt.Items[i].Sum = float64(int(transformedReceipt.Ticket.Document.Receipt.Items[i].Sum * 100))
				transformedReceipt.Ticket.Document.Receipt.Items[i].Price = float64(int(transformedReceipt.Ticket.Document.Receipt.Items[i].Price * 100))
			}

			mu.Lock()
			transformedReceipts = append(transformedReceipts, transformedReceipt)
			mu.Unlock()

			formattedDateTime := formatDateTime(fiscalData.DateTime)
			qrText := fmt.Sprintf("t=%s&s=%.2f&fn=%s&i=%d&fp=%s&n=%d",
				formattedDateTime,
				fiscalData.TotalSum,
				fiscalData.FiscalDriveNumber,
				fiscalData.FiscalDocumentNumber,
				fiscalData.FiscalSign,
				fiscalData.OperationType)

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
