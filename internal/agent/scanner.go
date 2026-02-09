package agent

import (
	"context"
	"fmt"
	"log"
	"sort"
	"sync"

	"jaycetrades.com/internal/schwab"
	"jaycetrades.com/internal/sentiment"
)

type Scanner struct {
	schwabClient *schwab.Client
	scraper      *sentiment.Scraper
}

type ScanConfig struct {
	MaxPremium       float64 // Maximum option premium in dollars
	MaxDaysToExpiry  int     // Maximum days until expiration
	MinVolume        int64   // Minimum option volume
	MinOpenInterest  int     // Minimum open interest
	TickerLimit      int     // Max tickers to scan from sentiment
}

type OptionCandidate struct {
	Symbol            string
	OptionSymbol      string
	ContractType      string  // CALL or PUT
	StrikePrice       float64
	ExpirationDate    string
	DaysToExpiry      int
	Bid               float64
	Ask               float64
	Premium           float64 // Cost to buy (ask * 100)
	Volume            int64
	OpenInterest      int
	ImpliedVolatility float64
	Delta             float64
	UnderlyingPrice   float64
	SentimentScore    float64
	Mentions          int
	Score             float64 // Overall ranking score
	Reasoning         string
}

func NewScanner(schwabClient *schwab.Client, scraper *sentiment.Scraper) *Scanner {
	return &Scanner{
		schwabClient: schwabClient,
		scraper:      scraper,
	}
}

func DefaultScanConfig() ScanConfig {
	return ScanConfig{
		MaxPremium:       500.0,  // $500 max premium
		MaxDaysToExpiry:  7,      // Weekly or 0DTE
		MinVolume:        100,    // Minimum volume
		MinOpenInterest:  50,     // Minimum open interest
		TickerLimit:      20,     // Top 20 trending tickers
	}
}

// Scan performs the full scanning workflow
func (s *Scanner) Scan(ctx context.Context, config ScanConfig) ([]OptionCandidate, error) {
	log.Println("Starting options scan...")

	// Step 1: Get trending tickers from sentiment analysis
	log.Println("Fetching trending tickers from social media...")
	tickers, err := s.scraper.GetTrendingTickers(ctx, config.TickerLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to get trending tickers: %w", err)
	}

	if len(tickers) == 0 {
		return nil, fmt.Errorf("no trending tickers found")
	}

	log.Printf("Found %d trending tickers", len(tickers))

	// Step 2: Fetch option chains concurrently
	candidates := s.fetchOptionChains(ctx, tickers, config)

	// Step 3: Score and rank candidates
	s.scoreAndRank(candidates)

	// Sort by score
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	log.Printf("Found %d option candidates", len(candidates))

	return candidates, nil
}

func (s *Scanner) fetchOptionChains(ctx context.Context, tickers []sentiment.TickerMention, config ScanConfig) []OptionCandidate {
	var candidates []OptionCandidate
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Semaphore for rate limiting
	sem := make(chan struct{}, 5) // Max 5 concurrent requests

	for _, ticker := range tickers {
		wg.Add(1)
		go func(t sentiment.TickerMention) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			chain, err := s.schwabClient.GetOptionChain(t.Symbol, config.MaxDaysToExpiry)
			if err != nil {
				log.Printf("Failed to get option chain for %s: %v", t.Symbol, err)
				return
			}

			// Process calls
			for _, strikes := range chain.CallExpDateMap {
				for _, contracts := range strikes {
					for _, contract := range contracts {
						if candidate := s.filterContract(contract, chain.Underlying, t, config, "CALL"); candidate != nil {
							mu.Lock()
							candidates = append(candidates, *candidate)
							mu.Unlock()
						}
					}
				}
			}

			// Process puts
			for _, strikes := range chain.PutExpDateMap {
				for _, contracts := range strikes {
					for _, contract := range contracts {
						if candidate := s.filterContract(contract, chain.Underlying, t, config, "PUT"); candidate != nil {
							mu.Lock()
							candidates = append(candidates, *candidate)
							mu.Unlock()
						}
					}
				}
			}
		}(ticker)
	}

	wg.Wait()
	return candidates
}

func (s *Scanner) filterContract(
	contract schwab.OptionContract,
	underlying schwab.Underlying,
	ticker sentiment.TickerMention,
	config ScanConfig,
	contractType string,
) *OptionCandidate {
	// Calculate premium (ask price * 100 shares per contract)
	premium := contract.Ask * 100

	// Apply filters
	if premium > config.MaxPremium || premium <= 0 {
		return nil
	}
	if contract.DaysToExpiration > config.MaxDaysToExpiry {
		return nil
	}
	if contract.TotalVolume < config.MinVolume {
		return nil
	}
	if contract.OpenInterest < config.MinOpenInterest {
		return nil
	}
	// Skip if bid/ask spread is too wide (> 20%)
	if contract.Bid > 0 {
		spread := (contract.Ask - contract.Bid) / contract.Bid
		if spread > 0.20 {
			return nil
		}
	}

	return &OptionCandidate{
		Symbol:            underlying.Symbol,
		OptionSymbol:      contract.Symbol,
		ContractType:      contractType,
		StrikePrice:       contract.StrikePrice,
		ExpirationDate:    contract.ExpirationDate,
		DaysToExpiry:      contract.DaysToExpiration,
		Bid:               contract.Bid,
		Ask:               contract.Ask,
		Premium:           premium,
		Volume:            contract.TotalVolume,
		OpenInterest:      contract.OpenInterest,
		ImpliedVolatility: contract.ImpliedVolatility,
		Delta:             contract.Delta,
		UnderlyingPrice:   underlying.Last,
		SentimentScore:    ticker.Sentiment,
		Mentions:          ticker.Mentions,
	}
}

func (s *Scanner) scoreAndRank(candidates []OptionCandidate) {
	for i := range candidates {
		c := &candidates[i]
		score := 0.0
		reasons := []string{}

		// Sentiment score (0-30 points)
		sentimentPoints := (c.SentimentScore + 1) * 15 // Convert -1..1 to 0..30
		score += sentimentPoints
		if c.SentimentScore > 0.3 {
			reasons = append(reasons, fmt.Sprintf("Strong bullish sentiment (%.2f)", c.SentimentScore))
		}

		// Mention count (0-20 points)
		mentionPoints := float64(c.Mentions)
		if mentionPoints > 20 {
			mentionPoints = 20
		}
		score += mentionPoints
		if c.Mentions >= 5 {
			reasons = append(reasons, fmt.Sprintf("High social mention count (%d)", c.Mentions))
		}

		// Volume relative to open interest (0-20 points)
		if c.OpenInterest > 0 {
			volRatio := float64(c.Volume) / float64(c.OpenInterest)
			if volRatio > 1 {
				score += 20
				reasons = append(reasons, "Unusual volume vs open interest")
			} else {
				score += volRatio * 20
			}
		}

		// Delta preference - favor slightly OTM (0.3-0.4 delta) (0-15 points)
		absDelta := c.Delta
		if absDelta < 0 {
			absDelta = -absDelta
		}
		if absDelta >= 0.25 && absDelta <= 0.45 {
			score += 15
			reasons = append(reasons, "Optimal delta range")
		} else if absDelta >= 0.15 && absDelta <= 0.55 {
			score += 8
		}

		// Prefer 0DTE or 1DTE for day trades (0-15 points)
		if c.DaysToExpiry <= 1 {
			score += 15
			reasons = append(reasons, "0DTE/1DTE opportunity")
		} else if c.DaysToExpiry <= 3 {
			score += 10
			reasons = append(reasons, "Short-dated expiry")
		} else {
			score += 5
		}

		// Prefer lower premiums for better risk/reward (0-10 points, inversely scaled)
		premiumScore := (1 - (c.Premium / 500)) * 10
		if premiumScore < 0 {
			premiumScore = 0
		}
		score += premiumScore

		c.Score = score
		if len(reasons) > 0 {
			c.Reasoning = fmt.Sprintf("%v", reasons)
		} else {
			c.Reasoning = "Meets basic criteria"
		}
	}
}

// GetTopPicks returns the top N option candidates
func (s *Scanner) GetTopPicks(candidates []OptionCandidate, n int) []OptionCandidate {
	if len(candidates) <= n {
		return candidates
	}
	return candidates[:n]
}
