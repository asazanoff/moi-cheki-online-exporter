package main

import (
	"time"
)

type Receipt struct {
	Key           string `json:"key"`
	CreatedDate   string `json:"createdDate"`
	TotalSum      string `json:"totalSum"`
	FiscalNumber  string `json:"fiscalDriveNumber"`
	DocumentNum   string `json:"fiscalDocumentNumber"`
	FiscalSign    string `json:"fiscalSign"`
	OperationType int    `json:"operationType"`
}

type ReceiptResponse struct {
	Receipts []Receipt `json:"receipts"`
	HasMore  bool      `json:"hasMore"`
}

type Item struct {
	Name        string  `json:"name"`
	NDS         int     `json:"nds"`
	NDSSum      int64   `json:"ndsSum"` // Изменено на int64
	PaymentType int     `json:"paymentType"`
	Price       int64   `json:"price"` // Изменено на int64
	ProductType int     `json:"productType"`
	ProviderINN string  `json:"providerInn"`
	Quantity    float64 `json:"quantity"` // Оставлено как float64, так как количество может быть дробным
	Sum         int64   `json:"sum"`      // Изменено на int64
}

type FiscalDataResponse struct {
	BuyerPhoneOrAddress     string `json:"buyerAddress"`
	CashTotalSum            int64  `json:"cashTotalSum"` // Изменено на int64
	Code                    int    `json:"code"`
	CreditSum               int64  `json:"creditSum"` // Изменено на int64
	DateTime                string `json:"dateTime"`
	ECashTotalSum           int64  `json:"ecashTotalSum"` // Изменено на int64
	FiscalDocumentFormatVer string `json:"fiscalDocumentFormatVer"`
	FiscalDocumentNumber    int    `json:"fiscalDocumentNumber"`
	FiscalDriveNumber       string `json:"fiscalDriveNumber"`
	FiscalSign              string `json:"fiscalSign"`
	Items                   []Item `json:"items"`
	KKTRegID                string `json:"kktRegId"`
	NDS10                   int64  `json:"nds10"` // Изменено на int64
	NDS18                   int64  `json:"nds18"` // Изменено на int64
	OperationType           int    `json:"operationType"`
	Operator                string `json:"operator"`
	OperatorINN             string `json:"operatorInn"`
	PrepaidSum              int64  `json:"prepaidSum"`   // Изменено на int64
	ProvisionSum            int64  `json:"provisionSum"` // Изменено на int64
	RequestNumber           int    `json:"requestNumber"`
	RetailPlace             string `json:"retailPlace"`
	RetailPlaceAddress      string `json:"retailPlaceAddress"`
	ShiftNumber             int    `json:"shiftNumber"`
	TaxationType            int    `json:"taxationType"`
	AppliedTaxationType     int    `json:"appliedTaxationType"`
	TotalSum                int64  `json:"totalSum"` // Изменено на int64
	User                    string `json:"user"`
	UserINN                 string `json:"userInn"`
}

type TransformedReceipt struct {
	ID        string    `json:"_id"`
	CreatedAt time.Time `json:"createdAt"`
	Ticket    struct {
		Document struct {
			Receipt FiscalDataResponse `json:"receipt"`
		} `json:"document"`
	} `json:"ticket"`
}
