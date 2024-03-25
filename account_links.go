// sky-accounts/pkg/clientlib/accountslib/accounts_link.go
package accountsclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// AccountLink represents the structure of an account link.
type AccountLink struct {
	UserID      uuid.UUID `json:"user_id"`
	AccountType string    `json:"account_type"`
	AccountID   uuid.UUID `json:"account_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// AccountLinkRequest represents the structure of an account link request.
type AccountLinkRequest struct {
	UserID      uuid.UUID `json:"user_id"`
	AccountType string    `json:"account_type"`
	AccountID   uuid.UUID `json:"account_id"`
}

func (c *Client) CreateAccountLink(alr AccountLinkRequest) (*AccountLink, error) {
	// Form the correct URL by combining the base URL with the path to the API endpoint
	url := fmt.Sprintf("%s/account_links", c.BaseURL)

	// Marshal the AccountLinkRequest to JSON
	requestBody, err := json.Marshal(alr)
	if err != nil {
		return nil, err
	}

	// Create a new HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	// Add necessary headers
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Add("X-Api-Key", c.ApiKey)

	// Send the request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create account link, status: %d, response: %s", resp.StatusCode, string(bodyBytes))
	}

	// Decode the response body into an AccountLink struct
	var accountLink AccountLink
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&accountLink)
	if err != nil {
		return nil, err
	}

	// Return the created AccountLink
	return &accountLink, nil
}

// GetAccountLink retrieves an account link by user ID, account type, and account ID
func (c *Client) GetAccountLink(userID uuid.UUID, accountID uuid.UUID) (*AccountLink, error) {
	// Construct the URL
	url := fmt.Sprintf("%s/account_link/%s/%s", c.BaseURL, userID, accountID)

	// Create a new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add headers
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Add("X-Api-Key", c.ApiKey)

	// Send the request and get the response
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Check the status code
	if res.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", res.StatusCode, string(bodyBytes))
	}

	// Decode the response
	var accountLink AccountLink
	err = json.NewDecoder(res.Body).Decode(&accountLink)
	if err != nil {
		return nil, err
	}

	// Return the account link
	return &accountLink, nil
}

func (c *Client) GetAccountLinksByUserID(userID uuid.UUID) ([]AccountLink, error) {
	// Prepare the API endpoint with the provided user ID
	url := fmt.Sprintf("%s/api/v1/accountLinks/%s", c.BaseURL, userID)

	// Prepare a new HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new HTTP request: %w", err)
	}

	// Add necessary headers to the request
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Add("X-API-Key", c.ApiKey)

	// Send the HTTP request
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer res.Body.Close()

	// Check the HTTP response status code
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received unexpected status code: %d", res.StatusCode)
	}

	// Parse the HTTP response body
	var accountLinks []AccountLink
	if err := json.NewDecoder(res.Body).Decode(&accountLinks); err != nil {
		return nil, fmt.Errorf("failed to parse HTTP response: %w", err)
	}

	// Return the list of account links
	return accountLinks, nil
}

// GetAccountLinksByAccountID fetches account links by account ID from the remote server.
func (c *Client) GetAccountLinksByAccountID(accountID uuid.UUID) ([]AccountLink, error) {
	// Construct the URL for the API endpoint
	apiEndpoint := fmt.Sprintf("%s/api/account_links/%s", c.BaseURL, accountID.String())

	// Create the HTTP request
	req, err := http.NewRequest(http.MethodGet, apiEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add necessary headers, such as Authorization
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("X-Api-Key", c.ApiKey)

	// Send the HTTP request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check the HTTP status code
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned non-200 status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse the response body
	var accountLinks []AccountLink
	err = json.NewDecoder(resp.Body).Decode(&accountLinks)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return accountLinks, nil
}

func (c *Client) GetAccountLinksByAccountType(accountType string) ([]AccountLink, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/accountlinks/accounttype/%s", c.BaseURL, accountType), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("X-API-KEY", c.ApiKey)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	var accountLinks []AccountLink

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &accountLinks); err != nil {
		return nil, err
	}

	return accountLinks, nil
}

func (c *Client) UpdateAccountLink(userID uuid.UUID, accountType string, accountID uuid.UUID) error {
	// Create a new request struct
	req := AccountLinkRequest{
		UserID:      userID,
		AccountType: accountType,
		AccountID:   accountID,
	}

	// Marshal the request struct to JSON
	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Construct the URL for the request
	url := fmt.Sprintf("%s/accountlink/%s", c.BaseURL, userID)

	// Create a new HTTP request
	httpRequest, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add necessary headers
	httpRequest.Header.Add("Content-Type", "application/json")
	httpRequest.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Token))

	// Send the HTTP request
	resp, err := c.HttpClient.Do(httpRequest)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check for a successful status code
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("received non-OK HTTP status: %s, Body: %s", resp.Status, string(bodyBytes))
	}

	return nil
}

func (c *Client) DeleteAccountLink(accountLinkRequest *AccountLinkRequest) error {
	// Create a new request using http
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/users/%s/accounts/%s?type=%s", c.BaseURL, accountLinkRequest.UserID, accountLinkRequest.AccountID, accountLinkRequest.AccountType), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// add authorization header to the req
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Token))

	// Send req using http Client
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("DeleteAccountLink failed: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		// Read the body
		body, readErr := io.ReadAll(res.Body)
		if readErr != nil {
			return fmt.Errorf("failed reading server response: %w", readErr)
		}

		return fmt.Errorf("DeleteAccountLink returned a non-200 status code, response: %s", body)
	}

	return nil
}

func (c *Client) ListAccountLinks(userID uuid.UUID) ([]AccountLink, error) {
	url := fmt.Sprintf("%s/api/v1/account_links/%s", c.BaseURL, userID.String())

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-API-Key", c.ApiKey)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var accountLinks []AccountLink
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&accountLinks); err != nil {
		return nil, err
	}

	return accountLinks, nil
}

func (c *Client) IsUserLinkedToAccount(userID uuid.UUID, accountID uuid.UUID) (bool, error) {
	// Create a new request to get AccountLink information
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/accountLink/%s", c.BaseURL, accountID), nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	// Add necessary headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("X-Api-Key", c.ApiKey)

	// Send the request and get the response
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("server responded with status code %d", resp.StatusCode)
	}

	// Decode the response body
	var accountLink AccountLink
	err = json.NewDecoder(resp.Body).Decode(&accountLink)
	if err != nil {
		return false, fmt.Errorf("failed to decode response body: %w", err)
	}

	// Check if the user is linked to the account
	if accountLink.UserID == userID {
		return true, nil
	}

	return false, nil
}

// GetLinkedAccountsForUser fetches all the linked accounts for a specific user.
func (c *Client) GetLinkedAccountsForUser(userID uuid.UUID) ([]AccountLink, error) {
	var accounts []AccountLink

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/accounts/%s/linked", c.BaseURL, userID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-Api-Key", c.ApiKey)

	res, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: got %v", res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	err = json.Unmarshal(body, &accounts)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	return accounts, nil
}
