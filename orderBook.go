package bittrex

import "github.com/shopspring/decimal"

//OrderBook struct
type OrderBook struct {
	MarketSymbol string       `json:"marketSymbol"`
	Depth        int          `json:"depth"`
	Sequence     int          `json:"sequence"`
	BidDeltas    []OrderDelta `json:"bidDeltas"`
	AskDeltas    []OrderDelta `json:"askDeltas"`
}

//OrderDelta struct
type OrderDelta struct {
	Quantity decimal.Decimal `json:"quantity"`
	Rate     decimal.Decimal `json:"rate"`
}

//OrderBook2 struct
type OrderBook2 struct {
	BidDeltas []OrderDelta `json:"bid"`
	AskDeltas []OrderDelta `json:"ask"`
}
