package schwab

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

// GetOptionChain fetches option chain for a symbol
func (c *Client) GetOptionChain(symbol string, daysToExpiration int) (*OptionChainResponse, error) {
	params := url.Values{
		"symbol":       {symbol},
		"contractType": {"ALL"},
		"strategy":     {"SINGLE"},
		"range":        {"ALL"},
		"toDate":       {time.Now().AddDate(0, 0, daysToExpiration).Format("2006-01-02")},
	}

	body, err := c.doRequest("GET", "/chains", params)
	if err != nil {
		return nil, err
	}

	var chain OptionChainResponse
	if err := json.Unmarshal(body, &chain); err != nil {
		return nil, fmt.Errorf("failed to parse option chain: %w", err)
	}

	return &chain, nil
}

// GetQuotes fetches quotes for multiple symbols
func (c *Client) GetQuotes(symbols []string) (map[string]Quote, error) {
	params := url.Values{
		"symbols": {joinSymbols(symbols)},
	}

	body, err := c.doRequest("GET", "/quotes", params)
	if err != nil {
		return nil, err
	}

	var quotes map[string]Quote
	if err := json.Unmarshal(body, &quotes); err != nil {
		return nil, fmt.Errorf("failed to parse quotes: %w", err)
	}

	return quotes, nil
}

// GetMovers gets top movers for an index
func (c *Client) GetMovers(index string) ([]Mover, error) {
	body, err := c.doRequest("GET", fmt.Sprintf("/movers/%s", index), nil)
	if err != nil {
		return nil, err
	}

	var movers MoversResponse
	if err := json.Unmarshal(body, &movers); err != nil {
		return nil, fmt.Errorf("failed to parse movers: %w", err)
	}

	return movers.Screeners, nil
}

type MoversResponse struct {
	Screeners []Mover `json:"screeners"`
}

type Mover struct {
	Symbol        string  `json:"symbol"`
	Description   string  `json:"description"`
	Volume        int64   `json:"volume"`
	LastPrice     float64 `json:"lastPrice"`
	NetChange     float64 `json:"netChange"`
	PercentChange float64 `json:"netPercentChange"`
}

func joinSymbols(symbols []string) string {
	result := ""
	for i, s := range symbols {
		if i > 0 {
			result += ","
		}
		result += s
	}
	return result
}
