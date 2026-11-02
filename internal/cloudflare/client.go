package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const apiBaseURL = "https://api.cloudflare.com/client/v4"

const (
	// maxAttempts is the total tries for one request when Cloudflare returns 429.
	maxAttempts = 4
	// maxRetryWait caps how long one Retry-After header can stall the TUI.
	maxRetryWait = 10 * time.Second
	// maxResponseBytes caps how much of a response body is read into memory.
	maxResponseBytes = 10 << 20
)

// Client handles Cloudflare API communication
type Client struct {
	apiToken   string
	httpClient *http.Client
	baseURL    string
	wait       func(time.Duration) // replaced in tests to skip retry sleeps
}

// NewClient creates a new Cloudflare API client
func NewClient(apiToken string) *Client {
	return &Client{
		apiToken: apiToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: apiBaseURL,
		wait:    time.Sleep,
	}
}

// retryDelay returns how long to wait before retry number attempt (1-based).
// It honors a numeric Retry-After header, capped at maxRetryWait.
func retryDelay(resp *http.Response, attempt int) time.Duration {
	delay := time.Duration(1<<(attempt-1)) * time.Second
	if secs, err := strconv.Atoi(resp.Header.Get("Retry-After")); err == nil && secs >= 0 {
		delay = time.Duration(secs) * time.Second
	}
	if delay > maxRetryWait {
		delay = maxRetryWait
	}
	return delay
}

// doRequest executes an authenticated API request and returns the response body.
// path is relative to the base URL and must already be escaped.
// A 429 means Cloudflare did not process the request, so retrying is safe for every method.
func (c *Client) doRequest(method, path string, body []byte) ([]byte, error) {
	for attempt := 1; ; attempt++ {
		var reader io.Reader
		if body != nil {
			reader = bytes.NewReader(body)
		}
		req, err := http.NewRequest(method, c.baseURL+path, reader)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+c.apiToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request failed: %w", err)
		}

		// Read one byte past the cap so an oversized body is detected, not silently cut.
		respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}
		if len(respBody) > maxResponseBytes {
			return nil, fmt.Errorf("response too large (over %d bytes)", maxResponseBytes)
		}

		if resp.StatusCode == http.StatusTooManyRequests && attempt < maxAttempts {
			c.wait(retryDelay(resp, attempt))
			continue
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
		}

		return respBody, nil
	}
}

// apiErrorResponse is the common error structure in Cloudflare API responses.
type apiErrorResponse struct {
	Success bool `json:"success"`
	Errors  []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
}

// checkAPIError unmarshals the common success/error fields and returns an error if the request failed.
func checkAPIError(data []byte) error {
	var resp apiErrorResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	if !resp.Success {
		if len(resp.Errors) > 0 {
			return fmt.Errorf("API error: %s", resp.Errors[0].Message)
		}
		return fmt.Errorf("API request failed")
	}
	return nil
}

// ListDNSRecords fetches all DNS records of a specific type from a zone.
// Handles pagination automatically to retrieve all records.
func (c *Client) ListDNSRecords(zoneID string, recordType string) ([]DNSRecord, error) {
	var allRecords []DNSRecord
	page := 1
	perPage := 100  // Maximum allowed by Cloudflare API
	maxPages := 100 // Safety limit: 10,000 records max

	for {
		path := fmt.Sprintf("/zones/%s/dns_records?page=%d&per_page=%d", url.PathEscape(zoneID), page, perPage)
		if recordType != "" {
			path += "&type=" + url.QueryEscape(recordType)
		}

		data, err := c.doRequest("GET", path, nil)
		if err != nil {
			return nil, err
		}

		var apiResp APIResponse
		if err := json.Unmarshal(data, &apiResp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		if !apiResp.Success {
			if len(apiResp.Errors) > 0 {
				return nil, fmt.Errorf("API error: %s", apiResp.Errors[0].Message)
			}
			return nil, fmt.Errorf("API request failed")
		}

		allRecords = append(allRecords, apiResp.Result...)

		if apiResp.ResultInfo == nil || page >= apiResp.ResultInfo.TotalPages {
			break
		}

		page++
		if page > maxPages {
			break
		}
	}

	return allRecords, nil
}

// CreateDNSRecord creates a new DNS record in Cloudflare
func (c *Client) CreateDNSRecord(zoneID string, record DNSRecord) (*DNSRecord, error) {
	path := fmt.Sprintf("/zones/%s/dns_records", url.PathEscape(zoneID))

	body, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	data, err := c.doRequest("POST", path, body)
	if err != nil {
		return nil, err
	}

	if err := checkAPIError(data); err != nil {
		return nil, err
	}

	var result struct {
		Result DNSRecord `json:"result"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to decode result: %w", err)
	}

	return &result.Result, nil
}

// DeleteDNSRecord deletes a DNS record from Cloudflare
func (c *Client) DeleteDNSRecord(zoneID, recordID string) error {
	path := fmt.Sprintf("/zones/%s/dns_records/%s", url.PathEscape(zoneID), url.PathEscape(recordID))

	data, err := c.doRequest("DELETE", path, nil)
	if err != nil {
		return err
	}

	return checkAPIError(data)
}

// UpdateDNSRecord updates an existing DNS record in Cloudflare
func (c *Client) UpdateDNSRecord(zoneID, recordID string, record DNSRecord) (*DNSRecord, error) {
	path := fmt.Sprintf("/zones/%s/dns_records/%s", url.PathEscape(zoneID), url.PathEscape(recordID))

	body, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	data, err := c.doRequest("PATCH", path, body)
	if err != nil {
		return nil, err
	}

	if err := checkAPIError(data); err != nil {
		return nil, err
	}

	var result struct {
		Result DNSRecord `json:"result"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to decode result: %w", err)
	}

	return &result.Result, nil
}
