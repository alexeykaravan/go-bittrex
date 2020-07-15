package bittrex

import "github.com/shopspring/decimal"

//Market struct
type Market struct {
	Symbol              string          `json:"symbol"`
	BaseCurrencySymbol  string          `json:"baseCurrencySymbol"`
	QuoteCurrencySymbol string          `json:"quoteCurrencySymbol"`
	MinTradeSize        decimal.Decimal `json:"minTradeSize"`
	Precision           int             `json:"precision"`
	Status              string          `json:"status"`
	CreatedAt           jTime           `json:"createdAt"`
	Notice              string          `json:"notice"`
	ProhibitedIn        []string        `json:"prohibitedIn"`
}
