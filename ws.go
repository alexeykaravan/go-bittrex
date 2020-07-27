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

	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"

	"github.com/google/uuid"
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
	b.client.wsClient = signalr.NewWebsocketClient()

	b.client.wsClient.OnClientMethod = func(hub string, method string, messages []json.RawMessage) {

		switch method {
		case ORDERBOOK, TICKER, ORDER:
		case AUTHEXPIRED:
			_, err := b.authentication()
			if err != nil {
				fmt.Printf("authentication error: %s\n", err.Error())
			}
		default:
			//handle unsupported type
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
				fmt.Printf("unsupported message type: %v", p.Method)
			}

			select {
			case dataCh <- p:
			default:
				fmt.Printf("missed message: %v", p)
			}
		}
	}

	b.client.wsClient.OnMessageError = func(err error) {
		fmt.Printf("ERROR OCCURRED: %s\n", err.Error())
	}

	err := doAsyncTimeout(
		func() error {
			return b.client.wsClient.Connect("https", WSBASE, []string{WSHUB})
		}, func(err error) {
			if err == nil {
				b.client.wsClient.Close()
			}
		}, timeout)
	if err != nil {
		return err
	}

	auth, err := b.authentication()
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", auth)

	isauth, err := b.client.wsClient.CallHub(WSHUB, "IsAuthenticated")
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", isauth)

	s, err := b.client.wsClient.CallHub(WSHUB, "Subscribe", []interface{}{"heartbeat", "order"})
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", s)

	return nil
}

//SubscribeOrderbook func
func (b *Bittrex) SubscribeOrderbook(market string) error {

	msg, err := b.client.wsClient.CallHub(WSHUB, "Subscribe", []interface{}{"orderbook_" + market + "_25"})
	if err != nil {
		return err
	}

	r := []Responce{}

	err = json.Unmarshal(msg, &r)
	if err != nil {
		return err
	}

	for _, i := range r {
		if !i.Success {
			return fmt.Errorf(fmt.Sprintf("%v", i.ErrorCode))
		}
	}

	return nil
}

//SubscribeTicker func
func (b *Bittrex) SubscribeTicker(market string) error {

	msg, err := b.client.wsClient.CallHub(WSHUB, "Subscribe", []interface{}{"ticker_" + market})
	if err != nil {
		return err
	}

	r := []Responce{}

	err = json.Unmarshal(msg, &r)
	if err != nil {
		return err
	}

	for _, i := range r {
		if !i.Success {
			return fmt.Errorf(fmt.Sprintf("%v", i.ErrorCode))
		}
	}

	return nil
}

func (b *Bittrex) authentication() ([]byte, error) {
	apiTimestamp := time.Now().UnixNano() / 1000000
	UUID := uuid.New().String()

	preSign := strings.Join([]string{fmt.Sprintf("%d", apiTimestamp), UUID}, "")

	mac := hmac.New(sha512.New, []byte(b.client.apiSecret))
	_, err := mac.Write([]byte(preSign))
	sig := hex.EncodeToString(mac.Sum(nil))

	msg, err := b.client.wsClient.CallHub(WSHUB, "Authenticate", b.client.apiKey, apiTimestamp, UUID, sig)
	if err != nil {
		return nil, err
	}

	return msg, nil
}
