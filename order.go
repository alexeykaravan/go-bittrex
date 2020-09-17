package bittrex

import (
	"github.com/shopspring/decimal"
)

// Order struct
type Order struct {
	ID            string          `json:"id"`
	MarketSymbol  string          `json:"marketSymbol"`
	Direction     string          `json:"direction"`
	Type          string          `json:"type"`
	Quantity      decimal.Decimal `json:"quantity"`
	Limit         decimal.Decimal `json:"limit"`
	Ceiling       decimal.Decimal `json:"ceiling"`
	TimeInForce   string          `json:"timeInForce"`
	ClientOrderID string          `json:"clientOrderId"`
	FillQuantity  decimal.Decimal `json:"fillQuantity"`
	Commission    decimal.Decimal `json:"commission"`
	Proceeds      decimal.Decimal `json:"proceeds"`
	Status        string          `json:"status"`
	CreatedAt     jTime           `json:"createdAt"`
	UpdatedAt     *jTime          `json:"updatedAt"`
	ClosedAt      *jTime          `json:"closedAt"`
	OrderToCancel struct {
		Type string `json:"type"`
		ID   string `json:"id"`
	} `json:"orderToCancel"`
}

//OrderUpdate struct
type OrderUpdate struct {
	AccountID string `json:"accountId"`
	Sequence  int    `json:"sequence"`
	Delta     struct {
		ID           string `json:"id"`
		MarketSymbol string `json:"marketSymbol"`
		Direction    string `json:"direction"`
		Type         string `json:"type"`
		Quantity     string `json:"quantity"`
		Limit        string `json:"limit"`
		TimeInForce  string `json:"timeInForce"`
		FillQuantity string `json:"fillQuantity"`
		Commission   string `json:"commission"`
		Proceeds     string `json:"proceeds"`
		Status       string `json:"status"`
		CreatedAt    jTime  `json:"createdAt"`
		UpdatedAt    *jTime `json:"updatedAt"`
		ClosedAt     *jTime `json:"closedAt"`
	} `json:"delta"`
}
