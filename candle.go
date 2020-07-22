package bittrex

import "github.com/shopspring/decimal"

type candle struct {
	TimeStamp  candleTime      `json:"T"`
	Open       decimal.Decimal `json:"O"`
	Close      decimal.Decimal `json:"C"`
	High       decimal.Decimal `json:"H"`
	Low        decimal.Decimal `json:"L"`
	Volume     decimal.Decimal `json:"V"`
	BaseVolume decimal.Decimal `json:"BV"`
}

type newCandles struct {
	Ticks []candle `json:"ticks"`
}
