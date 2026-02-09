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

// doRequest executes an authenticated API request and returns the response body.
func (c *Client) doRequest(method, url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
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
		url := fmt.Sprintf("%s/zones/%s/dns_records?page=%d&per_page=%d", apiBaseURL, zoneID, page, perPage)
		if recordType != "" {
			url += fmt.Sprintf("&type=%s", recordType)
		}

		data, err := c.doRequest("GET", url, nil)
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
	url := fmt.Sprintf("%s/zones/%s/dns_records", apiBaseURL, zoneID)

	body, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	data, err := c.doRequest("POST", url, bytes.NewBuffer(body))
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
	url := fmt.Sprintf("%s/zones/%s/dns_records/%s", apiBaseURL, zoneID, recordID)

	data, err := c.doRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	return checkAPIError(data)
}

// UpdateDNSRecord updates an existing DNS record in Cloudflare
func (c *Client) UpdateDNSRecord(zoneID, recordID string, record DNSRecord) (*DNSRecord, error) {
	url := fmt.Sprintf("%s/zones/%s/dns_records/%s", apiBaseURL, zoneID, recordID)

	body, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	data, err := c.doRequest("PATCH", url, bytes.NewBuffer(body))
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
