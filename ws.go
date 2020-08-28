package bittrex

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/thebotguys/signalr"
)

// Packet contains
type Packet struct {
	Method    string
	OrderBook OrderBook
	Ticker    Ticker
	Order     Order
}

//Responce struct
type Responce struct {
	Success   bool        `json:"Success"`
	ErrorCode interface{} `json:"ErrorCode"`
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

// StartListener func
func (b *Bittrex) StartListener(dataCh chan<- Packet) error {
	const timeout = 15 * time.Second
	client := signalr.NewWebsocketClient()

	client.OnClientMethod = func(hub string, method string, messages []json.RawMessage) {

		switch method {
		case ORDERBOOK, TICKER, ORDER:
		case HEARTBEAT:
		case AUTHEXPIRED:
			fmt.Printf("AUTHEXPIRED\n")
		default:
			//handle unsupported type
			fmt.Printf("unsupported message type: %s\n", method)
			return
		}

		for _, msg := range messages {

			dbuf, err := base64.StdEncoding.DecodeString(strings.Trim(string(msg), `"`))
			if err != nil {
				fmt.Printf("DecodeString error: %s %s\n", err.Error(), string(msg))
				continue
			}

			r, err := zlib.NewReader(bytes.NewReader(append([]byte{120, 156}, dbuf...)))
			if err != nil {
				fmt.Printf("unzip error %s %s \n", err.Error(), string(msg))
				continue
			}
			defer r.Close()

			var out bytes.Buffer
			io.Copy(&out, r)

			p := Packet{Method: method}

			switch method {
			case ORDERBOOK:
				json.Unmarshal([]byte(out.String()), &p.OrderBook)
			case TICKER:
				json.Unmarshal([]byte(out.String()), &p.Ticker)
			case ORDER:
				json.Unmarshal([]byte(out.String()), &p.Order)
			default:
				//handle unsupported type
				//fmt.Printf("unsupported message type: %v", p.Method)
			}

			select {
			case dataCh <- p:
			default:
				fmt.Printf("missed message: %v", p)
			}
		}
	}

	client.OnMessageError = func(err error) {
		fmt.Printf("ERROR OCCURRED: %s\n", err.Error())
	}

	err := doAsyncTimeout(
		func() error {
			return client.Connect("https", WSBASE, []string{WSHUB})
		}, func(err error) {
			if err == nil {
				client.Close()
			}
		}, timeout)
	if err != nil {
		return err
	}

	_, err = b.Authentication()
	if err != nil {
		return err
	}

	_, err = client.CallHub(WSHUB, "IsAuthenticated")
	if err != nil {
		return err
	}

	_, err = client.CallHub(WSHUB, "Subscribe", []interface{}{"heartbeat", "order"})
	if err != nil {
		return err
	}

	go func() {
		ticker := time.NewTicker(8 * time.Minute)

		for {
			auth, err := b.Authentication()
			if err != nil {
				fmt.Printf("authentication error: %s - %s\n", auth, err)
			}

			<-ticker.C
		}
	}()

	return nil
}

//Authentication func
func (b *Bittrex) Authentication() ([]byte, error) {
	//apiTimestamp := time.Now().UnixNano() / 1000000
	//UUID := uuid.New().String()

	//preSign := strings.Join([]string{fmt.Sprintf("%d", apiTimestamp), UUID}, "")

	//mac := hmac.New(sha512.New, []byte(b.client.apiSecret))
	//_, err := mac.Write([]byte(preSign))
	//sig := hex.EncodeToString(mac.Sum(nil))

	/*msg, err := client.CallHub(WSHUB, "Authenticate", b.client.apiKey, apiTimestamp, UUID, sig)
	if err != nil {
		return nil, err
	}
	*/

	return nil, nil
}

// SubscribeExchangeUpdates subscribes for updates of the market.
// Updates will be sent to dataCh.
// To stop subscription, send to, or close 'stop'.
func (b *Bittrex) SubscribeExchangeUpdates(market string, dataCh chan<- Packet, stop <-chan bool, orderbook bool) error {
	const timeout = 5 * time.Second
	client := signalr.NewWebsocketClient()
	client.OnClientMethod = func(hub string, method string, messages []json.RawMessage) {

		if hub != WSHUB {
			return
		}

		p := Packet{Method: method}

		for _, msg := range messages {

			dbuf, err := base64.StdEncoding.DecodeString(strings.Trim(string(msg), `"`))
			if err != nil {
				fmt.Printf("DecodeString error: %s %s\n", err.Error(), string(msg))
				continue
			}

			r, err := zlib.NewReader(bytes.NewReader(append([]byte{120, 156}, dbuf...)))
			if err != nil {
				fmt.Printf("unzip error %s %s \n", err.Error(), string(msg))
				continue
			}
			defer r.Close()

			var out bytes.Buffer
			io.Copy(&out, r)

			switch method {
			case ORDERBOOK:
				json.Unmarshal([]byte(out.String()), &p.OrderBook)
			case TICKER:
				json.Unmarshal([]byte(out.String()), &p.Ticker)
			}

			select {
			case dataCh <- p:
			default:
				fmt.Printf("missed message: %v", p)
			}
		}
	}

	client.OnMessageError = func(err error) {
		fmt.Printf("ERROR OCCURRED: %s\n", err.Error())
	}

	err := doAsyncTimeout(
		func() error {
			return client.Connect("https", WSBASE, []string{WSHUB})
		}, func(err error) {
			if err == nil {
				client.Close()
			}
		}, timeout)
	if err != nil {
		return err
	}

	defer client.Close()

	_, err = client.CallHub(WSHUB, "Subscribe", []interface{}{"ticker_" + market})
	if err != nil {
		return err
	}

	if orderbook {
		_, err := client.CallHub(WSHUB, "Subscribe", []interface{}{"orderbook_" + market + "_25"})
		if err != nil {
			return err
		}
	}

	select {
	case <-stop:
	case <-client.DisconnectedChannel:
	}
	return nil
}
