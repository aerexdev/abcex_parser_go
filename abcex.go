// Package abcex provides a client for fetching USDT/RUB bid and ask prices
// from the ABCEX exchange order book.
//
// API endpoint: https://hub.abcex.io/api/v2/exchange/public/orderbook/depth?instrumentCode=USDTRUB
// No authentication required.
package abcex

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"
)

const (
	OrderbookURL = "https://hub.abcex.io/api/v2/exchange/public/orderbook/depth?instrumentCode=USDTRUB"
	DefaultTimeout = 15 * time.Second
)

// Tick holds bid and ask prices as integer micro-units (value × 10⁶).
// For example, a price of 87.66 RUB/USDT is stored as 87_660_000.
//
//   Green = best ask (lowest sell offer) — price to buy USDT
//   Red   = best bid (highest buy offer) — price to sell USDT
type Tick struct {
	Green int64 // ask in micro-RUB
	Red   int64 // bid in micro-RUB
}

// Mid returns the midpoint between ask and bid.
// Falls back to whichever side is non-zero if one is missing.
func (t Tick) Mid() int64 {
	switch {
	case t.Green == 0 && t.Red == 0:
		return 0
	case t.Green == 0:
		return t.Red
	case t.Red == 0:
		return t.Green
	default:
		return (t.Green + t.Red) / 2
	}
}

// Client fetches USDT/RUB order book data from ABCEX.
type Client struct {
	http *http.Client
}

// New creates a Client with the default 15-second HTTP timeout.
func New() *Client {
	return &Client{http: &http.Client{Timeout: DefaultTimeout}}
}

// NewWithHTTP creates a Client using the provided *http.Client.
func NewWithHTTP(hc *http.Client) *Client {
	return &Client{http: hc}
}

type orderEntry struct {
	Qty   string `json:"qty"`
	Price string `json:"price"`
}

type orderbookResponse struct {
	Ask []orderEntry `json:"ask"` // sorted ascending  — ask[0] = best (lowest) ask
	Bid []orderEntry `json:"bid"` // sorted descending — bid[0] = best (highest) bid
}

// FetchTick returns the best ask (Green) and best bid (Red) from the ABCEX
// USDT/RUB order book as integer micro-units (price × 10⁶).
func (c *Client) FetchTick(ctx context.Context) (Tick, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, OrderbookURL, nil)
	if err != nil {
		return Tick{}, fmt.Errorf("abcex: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Origin", "https://abcex.io")
	req.Header.Set("Referer", "https://abcex.io/")

	resp, err := c.http.Do(req)
	if err != nil {
		return Tick{}, fmt.Errorf("abcex: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Tick{}, fmt.Errorf("abcex: unexpected HTTP status %d", resp.StatusCode)
	}

	var ob orderbookResponse
	if err := json.NewDecoder(resp.Body).Decode(&ob); err != nil {
		return Tick{}, fmt.Errorf("abcex: decode response: %w", err)
	}

	var tick Tick

	// Best ask = lowest sell offer = ask[0] (list is sorted ascending)
	if len(ob.Ask) > 0 && ob.Ask[0].Price != "" {
		tick.Green, _ = decimalToMicros(ob.Ask[0].Price)
	}

	// Best bid = highest buy offer = bid[0] (list is sorted descending)
	if len(ob.Bid) > 0 && ob.Bid[0].Price != "" {
		tick.Red, _ = decimalToMicros(ob.Bid[0].Price)
	}

	if tick.Green == 0 && tick.Red == 0 {
		return Tick{}, fmt.Errorf("abcex: no usable prices in order book response")
	}

	return tick, nil
}

// decimalToMicros converts a decimal string (e.g. "87.66") to int64 micro-units (×10⁶).
// Uses big.Rat for exact arithmetic — no float precision loss.
func decimalToMicros(s string) (int64, error) {
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		return 0, fmt.Errorf("abcex: invalid decimal %q", s)
	}
	r.Mul(r, big.NewRat(1_000_000, 1))
	r.Add(r, big.NewRat(1, 2)) // round half-up
	num := new(big.Int).Quo(r.Num(), r.Denom())
	if !num.IsInt64() {
		return 0, fmt.Errorf("abcex: value %q out of int64 range", s)
	}
	return num.Int64(), nil
}
