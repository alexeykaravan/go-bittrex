// Package bittrex is an implementation of the Biitrex API in Golang.
package bittrex

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

const (
	//APIBASE Bittrex API endpoint
	APIBASE = "https://api.bittrex.com/"
	//APIVERSION api version
	APIVERSION = "v3"
	//WSBASE Bittrex WS API endpoint
	WSBASE = "socket-v3.bittrex.com"
	//WSHUB SignalR main hub
	WSHUB = "CoreHub"
)

// New returns an instantiated bittrex struct
func New(apiKey, apiSecret string) *Bittrex {
	client := NewClient(apiKey, apiSecret)
	return &Bittrex{client}
}

// NewWithCustomHTTPClient returns an instantiated bittrex struct with custom http client
func NewWithCustomHTTPClient(apiKey, apiSecret string, httpClient *http.Client) *Bittrex {
	client := NewClientWithCustomHttpConfig(apiKey, apiSecret, httpClient)
	return &Bittrex{client}
}

// NewWithCustomTimeout returns an instantiated bittrex struct with custom timeout
func NewWithCustomTimeout(apiKey, apiSecret string, timeout time.Duration) *Bittrex {
	client := NewClientWithCustomTimeout(apiKey, apiSecret, timeout)
	return &Bittrex{client}
}

// handleErr gets JSON response from Bittrex API en deal with error
func handleErr(r jsonResponse) error {
	if !r.Success {
		return errors.New(r.Message)
	}
	return nil
}

// Bittrex represent a Bittrex client
type Bittrex struct {
	client *client
}

// SetDebug set enable/disable http request/response dump
func (b *Bittrex) SetDebug(enable bool) {
	b.client.debug = enable
}

// GetMarkets is used to get the open and available trading markets at Bittrex along with other meta data.
func (b *Bittrex) GetMarkets() (markets []Market, err error) {
	r, err := b.client.do("GET", "markets", "", false)
	if err != nil {
		return
	}

	err = json.Unmarshal(r, &markets)
	return
}

// GetTicker is used to get the current ticker values for a market.
func (b *Bittrex) GetTicker(market string) (ticker Ticker, err error) {
	r, err := b.client.do("GET", "markets/"+strings.ToUpper(market)+"/ticker", "", false)
	if err != nil {
		return
	}

	err = json.Unmarshal(r, &ticker)
	return
}

// Market

// BuyLimit is used to place a limited buy order in a specific market.
func (b *Bittrex) BuyLimit(market, force string, quantity, rate decimal.Decimal) (responce []byte, err error) {
	r, err := b.client.do("GET", fmt.Sprintf("market/buylimit?market=%s&quantity=%s&rate=%s&timeInForce=%s", market, quantity, rate, force), "", true)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// SellLimit is used to place a limited sell order in a specific market.
func (b *Bittrex) SellLimit(market, force string, quantity, rate decimal.Decimal) (response []byte, err error) {
	r, err := b.client.do("GET", fmt.Sprintf("market/selllimit?market=%s&quantity=%s&rate=%s&timeInForce=%s", market, quantity, rate, force), "", true)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// CancelOrder is used to cancel a buy or sell order.
func (b *Bittrex) CancelOrder(orderID string) (respone []byte, err error) {
	r, err := b.client.do("DELETE", "orders/"+orderID, "", true)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// GetOpenOrders returns orders that you currently have opened.
func (b *Bittrex) GetOpenOrders(market string) (openOrders []Order, err error) {
	resource := "orders/open"

	if market != "" {
		resource += "?marketSymbol=" + strings.ToUpper(market)
	}

	r, err := b.client.do("GET", resource, "", true)
	if err != nil {
		return
	}

	err = json.Unmarshal(r, &openOrders)
	return
}

// GetOrder func
func (b *Bittrex) GetOrder(orderUUID string) (order Order, err error) {

	resource := "orders/" + orderUUID

	r, err := b.client.do("GET", resource, "", true)
	if err != nil {
		return
	}

	err = json.Unmarshal(r, &order)
	return
}

// Account

// GetBalances is used to retrieve all balances from your account
func (b *Bittrex) GetBalances() (balances []Balance, err error) {
	r, err := b.client.do("GET", "balances", "", true)
	if err != nil {
		return
	}

	err = json.Unmarshal(r, &balances)
	return
}

// GetOrderHistory used to retrieve your order history.
// market string literal for the market (ie. BTC-LTC). If set to "all", will return for all market
func (b *Bittrex) GetOrderHistory(market string) (orders []Order, err error) {
	resource := "orders/closed"

	if market != "" {
		resource += "?marketSymbol=" + strings.ToUpper(market)
	}

	r, err := b.client.do("GET", resource, "", true)
	if err != nil {
		return
	}

	err = json.Unmarshal(r, &orders)
	return
}
