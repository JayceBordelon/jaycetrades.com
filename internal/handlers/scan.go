package handlers

import (
	"net/http"

	"jaycetrades.com/internal/agent"
	"jaycetrades.com/internal/schwab"
	"jaycetrades.com/internal/sentiment"

	"github.com/gin-gonic/gin"
)

type ScanHandler struct {
	schwabClient *schwab.Client
	scraper      *sentiment.Scraper
}

func NewScanHandler(schwabClient *schwab.Client, scraper *sentiment.Scraper) *ScanHandler {
	return &ScanHandler{
		schwabClient: schwabClient,
		scraper:      scraper,
	}
}

type ScanRequest struct {
	MaxPremium      float64 `json:"max_premium"`
	MaxDaysToExpiry int     `json:"max_days_to_expiry"`
	TopN            int     `json:"top_n"`
}

type ScanResponse struct {
	Candidates []OptionCandidateResponse `json:"candidates"`
	Count      int                       `json:"count"`
}

type OptionCandidateResponse struct {
	Symbol            string  `json:"symbol"`
	OptionSymbol      string  `json:"option_symbol"`
	ContractType      string  `json:"contract_type"`
	StrikePrice       float64 `json:"strike_price"`
	ExpirationDate    string  `json:"expiration_date"`
	DaysToExpiry      int     `json:"days_to_expiry"`
	Bid               float64 `json:"bid"`
	Ask               float64 `json:"ask"`
	Premium           float64 `json:"premium"`
	Volume            int64   `json:"volume"`
	OpenInterest      int     `json:"open_interest"`
	ImpliedVolatility float64 `json:"implied_volatility"`
	Delta             float64 `json:"delta"`
	UnderlyingPrice   float64 `json:"underlying_price"`
	SentimentScore    float64 `json:"sentiment_score"`
	Mentions          int     `json:"mentions"`
	Score             float64 `json:"score"`
	Reasoning         string  `json:"reasoning"`
}

// Scan runs the options discovery agent
func (h *ScanHandler) Scan(c *gin.Context) {
	if !h.schwabClient.IsAuthenticated() {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "not authenticated with Schwab",
			"message": "Please visit /auth/login to authenticate first",
		})
		return
	}

	var req ScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Use defaults
		req = ScanRequest{
			MaxPremium:      500,
			MaxDaysToExpiry: 7,
			TopN:            10,
		}
	}

	if req.MaxPremium <= 0 {
		req.MaxPremium = 500
	}
	if req.MaxDaysToExpiry <= 0 {
		req.MaxDaysToExpiry = 7
	}
	if req.TopN <= 0 {
		req.TopN = 10
	}

	scanner := agent.NewScanner(h.schwabClient, h.scraper)
	config := agent.ScanConfig{
		MaxPremium:       req.MaxPremium,
		MaxDaysToExpiry:  req.MaxDaysToExpiry,
		MinVolume:        100,
		MinOpenInterest:  50,
		TickerLimit:      20,
	}

	candidates, err := scanner.Scan(c.Request.Context(), config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	topPicks := scanner.GetTopPicks(candidates, req.TopN)

	response := ScanResponse{
		Candidates: make([]OptionCandidateResponse, len(topPicks)),
		Count:      len(topPicks),
	}

	for i, cand := range topPicks {
		response.Candidates[i] = OptionCandidateResponse{
			Symbol:            cand.Symbol,
			OptionSymbol:      cand.OptionSymbol,
			ContractType:      cand.ContractType,
			StrikePrice:       cand.StrikePrice,
			ExpirationDate:    cand.ExpirationDate,
			DaysToExpiry:      cand.DaysToExpiry,
			Bid:               cand.Bid,
			Ask:               cand.Ask,
			Premium:           cand.Premium,
			Volume:            cand.Volume,
			OpenInterest:      cand.OpenInterest,
			ImpliedVolatility: cand.ImpliedVolatility,
			Delta:             cand.Delta,
			UnderlyingPrice:   cand.UnderlyingPrice,
			SentimentScore:    cand.SentimentScore,
			Mentions:          cand.Mentions,
			Score:             cand.Score,
			Reasoning:         cand.Reasoning,
		}
	}

	c.JSON(http.StatusOK, response)
}

// GetTrending returns currently trending tickers from sentiment analysis
func (h *ScanHandler) GetTrending(c *gin.Context) {
	tickers, err := h.scraper.GetTrendingTickers(c.Request.Context(), 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tickers": tickers,
		"count":   len(tickers),
	})
}
