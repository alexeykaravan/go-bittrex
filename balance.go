package bittrex

import "github.com/shopspring/decimal"

// Balance struct
type Balance struct {
	CurrencySymbol string          `json:"currencySymbol"`
	Total          decimal.Decimal `json:"total"`
	Available      decimal.Decimal `json:"available"`
	UpdatedAt      *jTime          `json:"updatedAt"`
}
