package bittrex

import (
	"bytes"
	"compress/zlib"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/thebotguys/signalr"
)

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

//Authentication func
func (b *Bittrex) Authentication(c *signalr.Client) error {
	r := &Responce{}

	apiTimestamp := time.Now().UnixNano() / 1000000
	UUID := uuid.New().String()

	preSign := strings.Join([]string{fmt.Sprintf("%d", apiTimestamp), UUID}, "")

	mac := hmac.New(sha512.New, []byte(b.client.apiSecret))
	_, err := mac.Write([]byte(preSign))
	sig := hex.EncodeToString(mac.Sum(nil))

	auth, err := c.CallHub(WSHUB, "Authenticate", b.client.apiKey, apiTimestamp, UUID, sig)
	if err != nil {
		return err
	}

	_ = json.Unmarshal(auth, r)

	if !r.Success {
		return fmt.Errorf("%s", r.ErrorCode)
	}

	return nil
}

// SubscribeTickerUpdates subscribes for updates of the market.
func (b *Bittrex) SubscribeTickerUpdates(market string, ticker chan<- Ticker) error {
	const timeout = 5 * time.Second
	client := signalr.NewWebsocketClient()

	var updTime int64

	client.OnClientMethod = func(hub string, method string, messages []json.RawMessage) {
		if hub != WSHUB {
			return
		}

		switch method {
		case HEARTBEAT, TICKER, TRADE:
			atomic.StoreInt64(&updTime, time.Now().Unix())

		default:
			fmt.Printf("unsupported message type: %s\n", method)
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

			fmt.Println(out.String())

			p := Ticker{}
			json.Unmarshal([]byte(out.String()), &p)

			select {
			case ticker <- p:
			default:
				fmt.Printf("ticker send err: %s %d \n", market, len(ticker))
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

	_, err = client.CallHub(WSHUB, "Subscribe", []interface{}{"heartbeat", "ticker_" + market, "trade_" + market})
	if err != nil {
		return err
	}

	tick := time.NewTicker(1 * time.Minute)

	for {
		select {
		case <-client.DisconnectedChannel:
			return errors.New("client.DisconnectedChannel")
		case <-tick.C:
			if time.Now().Unix()-atomic.LoadInt64(&updTime) > 60 {
				return errors.New("ticker messages timeout")
			}
		}
	}
}

// SubscribeOrderUpdates func
func (b *Bittrex) SubscribeOrderUpdates(dataCh chan<- OrderUpdate) error {
	const timeout = 15 * time.Second
	client := signalr.NewWebsocketClient()

	client.OnClientMethod = func(hub string, method string, messages []json.RawMessage) {

		switch method {
		case ORDER:
		case HEARTBEAT:
			//fmt.Printf("HEARTBEAT\n")
		case AUTHEXPIRED:
			//fmt.Printf("AUTHEXPIRED\n")
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

			p := OrderUpdate{}

			switch method {
			case ORDER:
				json.Unmarshal([]byte(out.String()), &p)
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

	defer client.Close()

	err = b.Authentication(client)
	if err != nil {
		return err
	}

	_, err = client.CallHub(WSHUB, "Subscribe", []interface{}{"heartbeat", "order"})
	if err != nil {
		return err
	}

	ticker := time.NewTicker(5 * time.Minute)

	for {
		<-ticker.C

		err := b.Authentication(client)
		if err != nil {
			fmt.Printf("authentication error: %s\n", err)
			return err
		}
	}
}

// SubscribeOrderbookUpdates subscribes for updates of the market.
// Updates will be sent to dataCh.
// To stop subscription, send to, or close 'stop'.
func (b *Bittrex) SubscribeOrderbookUpdates(market string, orderbook chan<- OrderBook, stop chan bool) error {
	const timeout = 5 * time.Second
	client := signalr.NewWebsocketClient()

	var updTime time.Time

	client.OnClientMethod = func(hub string, method string, messages []json.RawMessage) {
		if hub != WSHUB {
			return
		}

		switch method {
		case HEARTBEAT, ORDERBOOK:
			updTime = time.Now()
		default:
			fmt.Printf("unsupported message type: %s %v\n", method, messages)
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

			p := OrderBook{}

			switch method {
			case ORDER:
				json.Unmarshal([]byte(out.String()), &p)
			default:

				//handle unsupported type
				//fmt.Printf("%s\n", out.String())
			}

			err = json.Unmarshal([]byte(out.String()), &p)
			if err != nil {
				fmt.Printf("orderbook Unmarshal err: %s %s\n", err.Error(), market)
			}

			select {
			case orderbook <- p:
			default:
				fmt.Printf("orderbook send err: %s %d  \n", market, len(orderbook))
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

	_, err = client.CallHub(WSHUB, "Subscribe", []interface{}{"heartbeat", "orderbook_" + market + "_25"})
	if err != nil {
		return err
	}

	tick := time.NewTicker(1 * time.Minute)

	for {
		select {
		case <-client.DisconnectedChannel:
			return errors.New("client.DisconnectedChannel")
		case <-stop:
			return errors.New("StopChannel")
		case <-tick.C:
			if time.Now().Sub(updTime) > time.Minute {
				return errors.New("orderook messages timeout")
			}
		}

	}
}
