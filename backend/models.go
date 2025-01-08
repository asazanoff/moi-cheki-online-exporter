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
	NDSSum      float32 `json:"ndsSum"`
	PaymentType int     `json:"paymentType"`
	Price       float32 `json:"price"`
	ProductType int     `json:"productType"`
	ProviderINN string  `json:"providerInn"`
	Quantity    float32 `json:"quantity"`
	Sum         float32 `json:"sum"`
}

type FiscalDataResponse struct {
	BuyerPhoneOrAddress     string  `json:"buyerAddress"`
	CashTotalSum            float32 `json:"cashTotalSum"`
	Code                    int     `json:"code"`
	CreditSum               float32 `json:"creditSum"`
	DateTime                string  `json:"dateTime"`
	ECashTotalSum           float32 `json:"ecashTotalSum"`
	FiscalDocumentFormatVer string  `json:"fiscalDocumentFormatVer"`
	FiscalDocumentNumber    int     `json:"fiscalDocumentNumber"`
	FiscalDriveNumber       string  `json:"fiscalDriveNumber"`
	FiscalSign              string  `json:"fiscalSign"`
	Items                   []Item  `json:"items"`
	KKTRegID                string  `json:"kktRegId"`
	NDS10                   float32 `json:"nds10"`
	NDS18                   float32 `json:"nds18"`
	OperationType           int     `json:"operationType"`
	Operator                string  `json:"operator"`
	OperatorINN             string  `json:"operatorInn"`
	PrepaidSum              float32 `json:"prepaidSum"`
	ProvisionSum            float32 `json:"provisionSum"`
	RequestNumber           int     `json:"requestNumber"`
	RetailPlace             string  `json:"retailPlace"`
	RetailPlaceAddress      string  `json:"retailPlaceAddress"`
	ShiftNumber             int     `json:"shiftNumber"`
	TaxationType            int     `json:"taxationType"`
	AppliedTaxationType     int     `json:"appliedTaxationType"`
	TotalSum                float32 `json:"totalSum"`
	User                    string  `json:"user"`
	UserINN                 string  `json:"userInn"`
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
