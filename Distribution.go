package bittrex

import "github.com/shopspring/decimal"

//Distribution struct
type Distribution struct {
	Distribution   []BalanceD      `json:"Distribution"`
	Balances       decimal.Decimal `json:"Balances"`
	AverageBalance decimal.Decimal `json:"AverageBalance"`
}

//BalanceD struct
type BalanceD struct {
	BalanceD decimal.Decimal `json:"Balance"`
}
