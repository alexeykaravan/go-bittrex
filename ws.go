package bittrex

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"

	"github.com/shopspring/decimal"

	"github.com/thebotguys/signalr"
)

type OrderUpdate struct {
	Orderb
	Type int
}

type Fill struct {
	Orderb
	OrderType string
	Timestamp jTime
}

// ExchangeState contains fills and order book updates for a market.
type ExchangeState struct {
	MarketName string
	Nounce     int
	Buys       []OrderUpdate
	Sells      []OrderUpdate
	Fills      []Fill
	Initial    bool
}

// Order struct
type Order3 struct {
	Uuid              string
	OrderUuid         string
	Id                int64
	Exchange          string
	OrderType         string
	Quantity          decimal.Decimal
	QuantityRemaining decimal.Decimal
	Limit             decimal.Decimal
	CommissionPaid    decimal.Decimal
	Price             decimal.Decimal
	PricePerUnit      decimal.Decimal
	Opened            jTime
	Closed            jTime
	IsOpen            bool
	CancelInitiated   bool
	ImmediateOrCancel bool
	IsConditional     bool
	Condition         string
	ConditionTarget   decimal.Decimal
	Updated           jTime
}

// OrderState contains
type OrderState struct {
	AccountUuid string
	Nonce       int
	Order       Order3
}

// doAsyncTimeout runs f in a different goroutine
//	if f returns before timeout elapses, doAsyncTimeout returns the result of f().
//	otherwise it returns "operation timeout" error, and calls tmFunc after f returns.
func doAsyncTimeout(f func() error, tmFunc func(error), timeout time.Duration) error {
	errs := make(chan error)
	go func() {
		err := f()
		select {
		case errs <- err:
		default:
			if tmFunc != nil {
				tmFunc(err)
			}
		}
	}()
	select {
	case err := <-errs:
		return err
	case <-time.After(timeout):
		return errors.New("operation timeout")
	}
}

func sendStateAsync(dataCh chan<- ExchangeState, st ExchangeState) {
	select {
	case dataCh <- st:
	default:
	}
}

func sendOrderAsync(dataCh chan<- OrderState, st OrderState) {
	select {
	case dataCh <- st:
	default:
	}

}

func subForMarket(client *signalr.Client, market string) (json.RawMessage, error) {
	_, err := client.CallHub(WS_HUB, "SubscribeToExchangeDeltas", market)
	if err != nil {
		return json.RawMessage{}, err
	}

	return client.CallHub(WS_HUB, "QueryExchangeState", market)
}

func parseStates(messages []json.RawMessage, dataCh chan<- ExchangeState, market string) {
	for _, msg := range messages {
		var st ExchangeState
		if err := json.Unmarshal(msg, &st); err != nil {
			continue
		}

		if st.MarketName != market {
			continue
		}
		sendStateAsync(dataCh, st)
	}
}

func parseOrders(messages []json.RawMessage, dataCh chan<- OrderState) {
	for _, msg := range messages {
		fmt.Printf("msg: %s\n", string(msg))

		var st OrderState
		if err := json.Unmarshal(msg, &st); err != nil {
			fmt.Printf("error : %s\n", err.Error())
		}

		sendOrderAsync(dataCh, st)
	}
}

// SubscribeExchangeUpdate subscribes for updates of the market.
// Updates will be sent to dataCh.
// To stop subscription, send to, or close 'stop'.
func (b *Bittrex) SubscribeExchangeUpdate(market string, dataCh chan<- ExchangeState, stop <-chan bool) error {
	const timeout = 15 * time.Second
	client := signalr.NewWebsocketClient()

	client.OnClientMethod = func(hub string, method string, messages []json.RawMessage) {
		if hub != WS_HUB || method != "updateExchangeState" {
			return
		}

		parseStates(messages, dataCh, market)
	}

	err := doAsyncTimeout(
		func() error {
			return client.Connect("https", WS_BASE, []string{WS_HUB})
		}, func(err error) {
			if err == nil {
				client.Close()
			}
		}, timeout)

	if err != nil {
		return err
	}

	defer client.Close()

	var msg json.RawMessage
	err = doAsyncTimeout(
		func() error {
			var err error
			msg, err = subForMarket(client, market)
			return err
		}, nil, timeout)
	if err != nil {
		return err
	}

	var st ExchangeState
	if err = json.Unmarshal(msg, &st); err != nil {
		return err
	}
	st.Initial = true
	st.MarketName = market
	sendStateAsync(dataCh, st)
	select {
	case <-stop:
	case <-client.DisconnectedChannel:
	}
	return nil
}

// SubscribeOrderUpdate func
// To stop subscription, send to, or close 'stop'.
func (b *Bittrex) SubscribeOrderUpdate(dataCh chan<- OrderState, stop <-chan bool, APIkey, APIsecret string) error {
	const timeout = 15 * time.Second
	client := signalr.NewWebsocketClient()

	client.OnClientMethod = func(hub string, method string, messages []json.RawMessage) {

		if hub != WS_HUB || method != "updateOrderState" {
			return
		}
		parseOrders(messages, dataCh)
	}
	err := doAsyncTimeout(
		func() error {
			return client.Connect("https", WS_BASE, []string{WS_HUB})
		}, func(err error) {
			if err == nil {
				client.Close()
			}
		}, timeout)
	if err != nil {
		return err
	}
	defer client.Close()

	var r json.RawMessage
	r, err = client.CallHub(WS_HUB, "GetAuthContext", APIkey)
	if err != nil {
		return err
	}

	var st string
	if err = json.Unmarshal(r, &st); err != nil {
		return err
	}

	h := hmac.New(sha512.New, []byte(APIsecret))
	h.Write([]byte(st))
	sha := hex.EncodeToString(h.Sum(nil))

	_, err = client.CallHub(WS_HUB, "Authenticate", APIkey, sha)
	if err != nil {
		return err
	}

	select {
	case <-stop:
	case <-client.DisconnectedChannel:
	}
	return nil
}
