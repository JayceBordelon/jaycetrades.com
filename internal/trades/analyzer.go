package trades

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"jaycetrades.com/internal/sentiment"
)

type Trade struct {
	Symbol         string  `json:"symbol"`
	ContractType   string  `json:"contract_type"` // CALL or PUT
	StrikePrice    float64 `json:"strike_price"`
	Expiration     string  `json:"expiration"`
	DTE            int     `json:"dte"` // Days to expiration
	EstimatedPrice float64 `json:"estimated_price"`
	Thesis         string  `json:"thesis"`
	SentimentScore float64 `json:"sentiment_score"`
}

type Analyzer struct {
	apiKey     string
	httpClient *http.Client
}

func NewAnalyzer(apiKey string) *Analyzer {
	return &Analyzer{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

type claudeRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

func (a *Analyzer) GetTopTrades(ctx context.Context, sentimentData []sentiment.TickerMention) ([]Trade, error) {
	sentimentJSON, err := json.Marshal(sentimentData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sentiment data: %w", err)
	}

	today := time.Now().Format("2006-01-02")
	weekday := time.Now().Weekday().String()

	prompt := fmt.Sprintf(AnalysisPrompt, today, weekday, string(sentimentJSON))

	reqBody := claudeRequest{
		Model:     "claude-sonnet-4-20250514",
		MaxTokens: 4096,
		Messages: []message{
			{Role: "user", Content: prompt},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Claude request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Claude API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Claude API returned status %d", resp.StatusCode)
	}

	var claudeResp claudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
		return nil, fmt.Errorf("failed to decode Claude response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return nil, fmt.Errorf("empty response from Claude")
	}

	var trades []Trade
	if err := json.Unmarshal([]byte(claudeResp.Content[0].Text), &trades); err != nil {
		return nil, fmt.Errorf("failed to parse trades from Claude response: %w", err)
	}

	// Enrich with sentiment scores
	sentimentMap := make(map[string]float64)
	for _, s := range sentimentData {
		sentimentMap[s.Symbol] = s.Sentiment
	}
	for i := range trades {
		if score, ok := sentimentMap[trades[i].Symbol]; ok {
			trades[i].SentimentScore = score
		}
	}

	return trades, nil
}
