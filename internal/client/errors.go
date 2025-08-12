package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type APIError struct {
	URL        string `json:"url"`
	Method     string `json:"method"`
	Detail     string `json:"detail"`
	Status     int    `json:"status"`
	Title      string `json:"title"`
	Type       string `json:"type"`
	StatusCode int    `json:"status_code"`
}

func (e *APIError) Error() string {
	message := e.Detail
	if message == "" {
		message = e.Title
	}
	if message == "" {
		message = http.StatusText(e.Status)
	}

	return fmt.Sprintf("%s request to %s failed with status code %d: %s",
		e.Method, e.URL, e.Status, message)
}

// IsNotFound returns true if the error is a 404 Not Found
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsValidationError returns true if the error is a 422 Unprocessable Entity
func (e *APIError) IsValidationError() bool {
	return e.StatusCode == 422
}

// IsUnauthorized returns true if the error is a 401 Unauthorized
func (e *APIError) IsUnauthorized() bool {
	return e.StatusCode == http.StatusUnauthorized
}

// IsConflict returns true if the error is a 409 Conflict
func (e *APIError) IsConflict() bool {
	return e.StatusCode == http.StatusConflict
}

// ParseErrorResponse parses the HTTP response into a simple error
func ParseErrorResponse(resp *http.Response, method, url string) error {
	defer func() {
		_ = resp.Body.Close() // Ignore close error
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("could not parse error response: %v", err)
	}

	errorResponse := &APIError{}
	err = json.Unmarshal(body, errorResponse)
	if err != nil {
		// If JSON parsing fails, create a simple error
		return &APIError{
			URL:        url,
			Method:     method,
			Detail:     string(body),
			Status:     resp.StatusCode,
			Title:      http.StatusText(resp.StatusCode),
			StatusCode: resp.StatusCode,
		}
	}

	errorResponse.URL = url
	errorResponse.Method = method
	errorResponse.StatusCode = resp.StatusCode
	return errorResponse
}

// Helper functions for common error checks
func IsNotFoundError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.IsNotFound()
	}
	return false
}

func IsValidationError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.IsValidationError()
	}
	return false
}

func IsUnauthorizedError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.IsUnauthorized()
	}
	return false
}

func IsConflictError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.IsConflict()
	}
	return false
}
