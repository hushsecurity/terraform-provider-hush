package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const defaultBaseURL = "https://api.us.hush-security.com/"

type Token struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

type Client struct {
	BaseURL           string
	ClientID          string
	ClientSecret      string
	Token             *Token
	TokenCreationTime int64
	httpClient        *http.Client
	mu                sync.RWMutex // Protect token refresh
}

func NewClient(ctx context.Context, clientID, clientSecret, baseURL string) (*Client, error) {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	client := &Client{
		BaseURL:      baseURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		httpClient:   http.DefaultClient,
	}

	// Get initial token
	if err := client.refreshToken(ctx); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) refreshToken(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	authURL := c.BaseURL + "/v1/oauth/token"
	req, err := http.NewRequestWithContext(ctx, "POST", authURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to build auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.ClientID, c.ClientSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("auth request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close() // Ignore close error
	}()

	if resp.StatusCode != http.StatusOK {
		return ParseErrorResponse(resp, "POST", authURL)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read auth response: %w", err)
	}

	var token Token
	if err := json.Unmarshal(bodyBytes, &token); err != nil {
		return fmt.Errorf("auth decode failed: %w", err)
	}
	if token.AccessToken == "" {
		return fmt.Errorf("auth succeeded but access_token is empty")
	}

	c.Token = &token
	c.TokenCreationTime = time.Now().Unix()
	return nil
}

func (c *Client) doRequest(ctx context.Context, method, path string, body any, result any) error {
	// Check if token needs refresh (refresh if expiring within 30 seconds)
	now := time.Now().Unix()
	c.mu.RLock()
	needsRefresh := c.Token == nil || c.TokenCreationTime+c.Token.ExpiresIn-now < 30
	c.mu.RUnlock()

	if needsRefresh {
		if err := c.refreshToken(ctx); err != nil {
			return fmt.Errorf("token refresh failed: %w", err)
		}
	}

	var bodyReader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(buf)
	}

	fullURL := c.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	c.mu.RLock()
	accessToken := c.Token.AccessToken
	c.mu.RUnlock()

	req.Header.Set("Authorization", "Bearer "+accessToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close() // Ignore close error
	}()

	if resp.StatusCode >= 400 {
		return ParseErrorResponse(resp, method, fullURL)
	}

	if result != nil {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read response: %w", err)
		}

		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}
	}
	return nil
}
