package bittrex

import "github.com/shopspring/decimal"

//Deposit struct
type Deposit struct {
	ID            int64           `json:"Id"`
	Amount        decimal.Decimal `json:"Amount"`
	Currency      string          `json:"Currency"`
	Confirmations int             `json:"Confirmations"`
	LastUpdated   jTime           `json:"LastUpdated"`
	TxID          string          `json:"TxId"`
	CryptoAddress string          `json:"CryptoAddress"`
}
