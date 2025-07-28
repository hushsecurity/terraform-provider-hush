package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	ErrNotFound      = errors.New("resource not found")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrBadRequest    = errors.New("bad request")
	ErrInternalError = errors.New("internal server error")
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

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/v1/oauth/token", bytes.NewBufferString(data.Encode()))
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
		if closeErr := resp.Body.Close(); closeErr != nil {
			tflog.Warn(ctx, "Failed to close response body", map[string]interface{}{
				"error": closeErr,
			})
		}
	}()

	bodyBytes, _ := io.ReadAll(resp.Body)
	tflog.Debug(ctx, "Auth response", map[string]interface{}{
		"status_code": resp.StatusCode,
	})

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth failed with status %d: %s", resp.StatusCode, string(bodyBytes))
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

	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, bodyReader)
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
		if closeErr := resp.Body.Close(); closeErr != nil {
			tflog.Warn(ctx, "Failed to close response body", map[string]interface{}{
				"error": closeErr,
			})
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	tflog.Debug(ctx, "HTTP request completed", map[string]interface{}{
		"method":      method,
		"path":        path,
		"status_code": resp.StatusCode,
	})

	if resp.StatusCode >= 400 {
		switch resp.StatusCode {
		case http.StatusNotFound:
			return fmt.Errorf("%w: %s", ErrNotFound, string(respBody))
		case http.StatusUnauthorized:
			return fmt.Errorf("%w: %s", ErrUnauthorized, string(respBody))
		case http.StatusBadRequest:
			return fmt.Errorf("%w: %s", ErrBadRequest, string(respBody))
		case http.StatusInternalServerError:
			return fmt.Errorf("%w: %s", ErrInternalError, string(respBody))
		default:
			return fmt.Errorf("unexpected status: %d: %s", resp.StatusCode, string(respBody))
		}
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}
	}
	return nil
}
