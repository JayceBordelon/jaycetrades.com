package schwab

import "time"

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

type OptionChainResponse struct {
	Symbol           string                     `json:"symbol"`
	Status           string                     `json:"status"`
	Underlying       Underlying                 `json:"underlying"`
	Strategy         string                     `json:"strategy"`
	Interval         float64                    `json:"interval"`
	IsDelayed        bool                       `json:"isDelayed"`
	IsIndex          bool                       `json:"isIndex"`
	DaysToExpiration float64                    `json:"daysToExpiration"`
	NumberOfContracts int                       `json:"numberOfContracts"`
	CallExpDateMap   map[string]map[string][]OptionContract `json:"callExpDateMap"`
	PutExpDateMap    map[string]map[string][]OptionContract `json:"putExpDateMap"`
}

type Underlying struct {
	Symbol            string  `json:"symbol"`
	Description       string  `json:"description"`
	Change            float64 `json:"change"`
	PercentChange     float64 `json:"percentChange"`
	Close             float64 `json:"close"`
	QuoteTime         int64   `json:"quoteTime"`
	TradeTime         int64   `json:"tradeTime"`
	Bid               float64 `json:"bid"`
	Ask               float64 `json:"ask"`
	Last              float64 `json:"last"`
	Mark              float64 `json:"mark"`
	MarkChange        float64 `json:"markChange"`
	MarkPercentChange float64 `json:"markPercentChange"`
	BidSize           int     `json:"bidSize"`
	AskSize           int     `json:"askSize"`
	HighPrice         float64 `json:"highPrice"`
	LowPrice          float64 `json:"lowPrice"`
	OpenPrice         float64 `json:"openPrice"`
	TotalVolume       int64   `json:"totalVolume"`
	FiftyTwoWeekHigh  float64 `json:"fiftyTwoWeekHigh"`
	FiftyTwoWeekLow   float64 `json:"fiftyTwoWeekLow"`
}

type OptionContract struct {
	PutCall                string  `json:"putCall"`
	Symbol                 string  `json:"symbol"`
	Description            string  `json:"description"`
	ExchangeName           string  `json:"exchangeName"`
	Bid                    float64 `json:"bid"`
	Ask                    float64 `json:"ask"`
	Last                   float64 `json:"last"`
	Mark                   float64 `json:"mark"`
	BidSize                int     `json:"bidSize"`
	AskSize                int     `json:"askSize"`
	LastSize               int     `json:"lastSize"`
	HighPrice              float64 `json:"highPrice"`
	LowPrice               float64 `json:"lowPrice"`
	OpenPrice              float64 `json:"openPrice"`
	ClosePrice             float64 `json:"closePrice"`
	TotalVolume            int64   `json:"totalVolume"`
	OpenInterest           int     `json:"openInterest"`
	VolatilityMeasurement  string  `json:"volatilityMeasurement"`
	ImpliedVolatility      float64 `json:"volatility"`
	Delta                  float64 `json:"delta"`
	Gamma                  float64 `json:"gamma"`
	Theta                  float64 `json:"theta"`
	Vega                   float64 `json:"vega"`
	Rho                    float64 `json:"rho"`
	StrikePrice            float64 `json:"strikePrice"`
	ExpirationDate         string  `json:"expirationDate"`
	DaysToExpiration       int     `json:"daysToExpiration"`
	ExpirationType         string  `json:"expirationType"`
	InTheMoney             bool    `json:"inTheMoney"`
	Multiplier             float64 `json:"multiplier"`
	PercentChange          float64 `json:"percentChange"`
	MarkChange             float64 `json:"markChange"`
	MarkPercentChange      float64 `json:"markPercentChange"`
}

type Quote struct {
	Symbol        string  `json:"symbol"`
	Bid           float64 `json:"bidPrice"`
	Ask           float64 `json:"askPrice"`
	Last          float64 `json:"lastPrice"`
	Volume        int64   `json:"totalVolume"`
	PercentChange float64 `json:"netPercentChangeInDouble"`
}

// Credentials holds Schwab API credentials
type Credentials struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// Session holds active auth session
type Session struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}
