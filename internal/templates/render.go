package templates

import (
	"bytes"
	"embed"
	"html/template"
	"time"
)

//go:embed email.html
var templateFS embed.FS

type Trade struct {
	Symbol         string
	ContractType   string
	StrikePrice    float64
	Expiration     string
	DTE            int
	EstimatedPrice float64
	Thesis         string
	SentimentScore float64
}

type EmailData struct {
	Subject string
	Date    string
	Trades  []Trade
}

func RenderEmail(trades []Trade) (string, error) {
	tmpl, err := template.ParseFS(templateFS, "email.html")
	if err != nil {
		return "", err
	}

	data := EmailData{
		Subject: "Today's Top Options Plays",
		Date:    time.Now().Format("Monday, Jan 2, 2006"),
		Trades:  trades,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
