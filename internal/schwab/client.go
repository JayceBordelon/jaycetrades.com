package schwab

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	authURL    = "https://api.schwabapi.com/v1/oauth/authorize"
	tokenURL   = "https://api.schwabapi.com/v1/oauth/token"
	apiBaseURL = "https://api.schwabapi.com/marketdata/v1"
)

type Client struct {
	creds      Credentials
	session    *Session
	httpClient *http.Client
}

func NewClient(creds Credentials) *Client {
	return &Client{
		creds: creds,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetAuthURL returns the URL to redirect users for OAuth consent
func (c *Client) GetAuthURL() string {
	params := url.Values{
		"client_id":     {c.creds.ClientID},
		"redirect_uri":  {c.creds.RedirectURI},
		"response_type": {"code"},
	}
	return fmt.Sprintf("%s?%s", authURL, params.Encode())
}

// ExchangeCode exchanges the authorization code for tokens
func (c *Client) ExchangeCode(code string) error {
	data := url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {c.creds.RedirectURI},
	}

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	// Basic auth with client credentials
	auth := base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("%s:%s", c.creds.ClientID, c.creds.ClientSecret)),
	)
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token exchange failed: %s", string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}

	c.session = &Session{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}

	return nil
}

// RefreshAccessToken refreshes the access token
func (c *Client) RefreshAccessToken() error {
	if c.session == nil {
		return fmt.Errorf("no session to refresh")
	}

	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {c.session.RefreshToken},
	}

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	auth := base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("%s:%s", c.creds.ClientID, c.creds.ClientSecret)),
	)
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}

	c.session.AccessToken = tokenResp.AccessToken
	c.session.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return nil
}

// IsAuthenticated checks if we have a valid session
func (c *Client) IsAuthenticated() bool {
	return c.session != nil && time.Now().Before(c.session.ExpiresAt)
}

// SetSession allows restoring a session from storage
func (c *Client) SetSession(session *Session) {
	c.session = session
}

// GetSession returns current session for persistence
func (c *Client) GetSession() *Session {
	return c.session
}

// doRequest makes an authenticated API request
func (c *Client) doRequest(method, endpoint string, params url.Values) ([]byte, error) {
	if !c.IsAuthenticated() {
		if c.session != nil {
			if err := c.RefreshAccessToken(); err != nil {
				return nil, fmt.Errorf("not authenticated: %w", err)
			}
		} else {
			return nil, fmt.Errorf("not authenticated")
		}
	}

	fullURL := apiBaseURL + endpoint
	if len(params) > 0 {
		fullURL += "?" + params.Encode()
	}

	req, err := http.NewRequest(method, fullURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.session.AccessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
