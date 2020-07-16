package bittrex

// NewOrder struct
type NewOrder struct {
	MarketSymbol  string `json:"marketSymbol"`
	Direction     string `json:"direction"`
	Type          string `json:"type"`
	Quantity      string `json:"quantity"`
	Ceiling       string `json:"ceiling,omitempty"`
	Limit         string `json:"limit"`
	TimeInForce   string `json:"timeInForce"`
	ClientOrderID string `json:"clientOrderId,omitempty"`
	UseAwards     bool   `json:"useAwards"`
}
