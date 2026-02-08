package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const apiBaseURL = "https://api.cloudflare.com/client/v4"

// Client handles Cloudflare API communication
type Client struct {
	apiToken   string
	httpClient *http.Client
}

// NewClient creates a new Cloudflare API client
func NewClient(apiToken string) *Client {
	return &Client{
		apiToken: apiToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ListDNSRecords fetches all DNS records of a specific type from a zone
// Handles pagination automatically to retrieve all records
func (c *Client) ListDNSRecords(zoneID string, recordType string) ([]DNSRecord, error) {
	var allRecords []DNSRecord
	page := 1
	perPage := 100   // Maximum allowed by Cloudflare API
	maxPages := 100  // Safety limit: 10,000 records max

	for {
		// Build URL with query parameters including pagination
		url := fmt.Sprintf("%s/zones/%s/dns_records?page=%d&per_page=%d", apiBaseURL, zoneID, page, perPage)
		if recordType != "" {
			url += fmt.Sprintf("&type=%s", recordType)
		}

		// Create request
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Add authentication header
		req.Header.Set("Authorization", "Bearer "+c.apiToken)
		req.Header.Set("Content-Type", "application/json")

		// Execute request
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request failed: %w", err)
		}

		// Read response body
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		// Check HTTP status
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
		}

		// Parse JSON response
		var apiResp APIResponse
		if err := json.Unmarshal(body, &apiResp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		// Check API success flag
		if !apiResp.Success {
			if len(apiResp.Errors) > 0 {
				return nil, fmt.Errorf("API error: %s", apiResp.Errors[0].Message)
			}
			return nil, fmt.Errorf("API request failed")
		}

		// Append results from this page
		allRecords = append(allRecords, apiResp.Result...)

		// Check if there are more pages
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
	url := fmt.Sprintf("%s/zones/%s/dns_records", apiBaseURL, zoneID)

	// Prepare request body
	body, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var apiResp struct {
		Success bool      `json:"success"`
		Result  DNSRecord `json:"result"`
		Errors  []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(bodyBytes, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.Success {
		if len(apiResp.Errors) > 0 {
			return nil, fmt.Errorf("API error: %s", apiResp.Errors[0].Message)
		}
		return nil, fmt.Errorf("API request failed")
	}

	return &apiResp.Result, nil
}

// DeleteDNSRecord deletes a DNS record from Cloudflare
func (c *Client) DeleteDNSRecord(zoneID, recordID string) error {
	url := fmt.Sprintf("%s/zones/%s/dns_records/%s", apiBaseURL, zoneID, recordID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var apiResp struct {
		Success bool `json:"success"`
		Errors  []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(bodyBytes, &apiResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.Success {
		if len(apiResp.Errors) > 0 {
			return fmt.Errorf("API error: %s", apiResp.Errors[0].Message)
		}
		return fmt.Errorf("API request failed")
	}

	return nil
}

// UpdateDNSRecord updates an existing DNS record in Cloudflare
func (c *Client) UpdateDNSRecord(zoneID, recordID string, record DNSRecord) (*DNSRecord, error) {
	url := fmt.Sprintf("%s/zones/%s/dns_records/%s", apiBaseURL, zoneID, recordID)

	// Prepare request body
	body, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var apiResp struct {
		Success bool      `json:"success"`
		Result  DNSRecord `json:"result"`
		Errors  []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(bodyBytes, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.Success {
		if len(apiResp.Errors) > 0 {
			return nil, fmt.Errorf("API error: %s", apiResp.Errors[0].Message)
		}
		return nil, fmt.Errorf("API request failed")
	}

	return &apiResp.Result, nil
}
