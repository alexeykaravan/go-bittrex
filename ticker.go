package bittrex

import "github.com/shopspring/decimal"

// Ticker struct
type Ticker struct {
	Symbol        string
	LastTradeRate decimal.Decimal `json:"lastTradeRate"`
	BidRate       decimal.Decimal `json:"bidRate"`
	AskRate       decimal.Decimal `json:"askRate"`
}
