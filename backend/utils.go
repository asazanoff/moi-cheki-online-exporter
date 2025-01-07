package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/skip2/go-qrcode"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"
)

// Function to format the datetime string
func formatDateTime(dateTime string) string {
	t, err := time.Parse("2006-01-02T15:04:05", dateTime)
	if debugMode {
		if err != nil {
			log.Printf("DEBUG: error parsing time: %v", err)
		} else {
			log.Printf("DEBUG: time is: %s", t)
			log.Printf("DEBUG: time formatted is: %s", t.Format("20060102T1504"))
		}
	}

	if err != nil {
		return dateTime
	}
	return t.Format("20060102T1504")
}

// Function to generate random ID
func generateRandomID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 24)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// Function to get receipts
func getReceipts(dateFrom, dateTo string) ([]Receipt, error) {
	client := &http.Client{}
	data := fmt.Sprintf(`{"limit":1000,"offset":0,"dateFrom":"%s","dateTo":"%s","orderBy":"CREATED_DATE:DESC"}`, dateFrom, dateTo)
	req, err := http.NewRequest("POST", fnsApiUrl+"/api/v1/receipt", bytes.NewBuffer([]byte(data)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	if debugMode {
		log.Printf("DEBUG: Request Type: POST")
		log.Printf("DEBUG: Request URL: %s", req.URL)
		log.Printf("DEBUG: Request Body: %s", data)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if debugMode {
		log.Printf("DEBUG: Response Status: %s", resp.Status)
		log.Printf("DEBUG: Response Body: %s", string(body))
	}

	var receiptResponse ReceiptResponse
	err = json.Unmarshal(body, &receiptResponse)
	if err != nil {
		return nil, err
	}
	return receiptResponse.Receipts, nil
}

// Function to get fiscal data for a receipt
func getFiscalData(key string) (*FiscalDataResponse, error) {
	client := &http.Client{}
	data := fmt.Sprintf(`{"key":"%s"}`, key)
	req, err := http.NewRequest("POST", fnsApiUrl+"/api/v1/receipt/fiscal_data", bytes.NewBuffer([]byte(data)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	if debugMode {
		log.Printf("DEBUG: Request Type: POST")
		log.Printf("DEBUG: Request URL: %s", req.URL)
		log.Printf("DEBUG: Request Body: %s", data)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if debugMode {
		log.Printf("DEBUG: Response Status: %s", resp.Status)
		log.Printf("DEBUG: Response Body: %s", string(body))
	}

	var fiscalDataResponse FiscalDataResponse
	err = json.Unmarshal(body, &fiscalDataResponse)
	if err != nil {
		return nil, err
	}
	return &fiscalDataResponse, nil
}

// Function to generate QR code
func generateQRCode(text string) ([]byte, error) {
	qrCode, err := qrcode.Encode(text, qrcode.Medium, 256)
	if err != nil {
		return nil, err
	}
	return qrCode, nil
}
