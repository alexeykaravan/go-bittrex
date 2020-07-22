package bittrex

import "github.com/shopspring/decimal"

//OrderBook struct
type OrderBook struct {
	Buy  []Orderb `json:"buy"`
	Sell []Orderb `json:"sell"`
}

//Orderb struct
type Orderb struct {
	Quantity decimal.Decimal `json:"Quantity"`
	Rate     decimal.Decimal `json:"Rate"`
}
