package accountslib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/google/uuid"
)

// AccountMembership represents the structure of an account membership.
type AccountMembership struct {
	ID          uuid.UUID `json:"id"`
	AccountType string    `json:"account_type"`
	AccountID   uuid.UUID `json:"account_id"`
	UserID      uuid.UUID `json:"user_id"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
}

// UpdateAccountMembershipEvent represents the structure of an update account membership event.
type UpdateAccountMembershipEvent struct {
	AccountType string    `json:"account_type,omitempty"`
	AccountID   uuid.UUID `json:"account_id,omitempty"`
	UserID      uuid.UUID `json:"user_id,omitempty"`
	Role        string    `json:"role,omitempty"`
}

// CreateAccountMembership sends a POST request to create a new account membership.
func (c *Client) CreateAccountMembership(accountMembership *AccountMembership) (*AccountMembership, error) {
	// Convert the AccountMembership struct to JSON
	accountMembershipJSON, err := json.Marshal(accountMembership)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal account membership: %v", err)
	}

	// Create a new request
	req, err := http.NewRequest("POST", c.BaseURL+"/api/account-membership", bytes.NewBuffer(accountMembershipJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Add necessary headers
	req.Header.Add("Authorization", "Bearer "+c.Token)
	req.Header.Add("Content-Type", "application/json")

	// Send the request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusCreated {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %v", err)
		}
		return nil, fmt.Errorf("failed to create account membership: %v", string(bodyBytes))
	}

	// Decode the response body
	var createdAccountMembership AccountMembership
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&createdAccountMembership); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %v", err)
	}

	// Return the created account membership
	return &createdAccountMembership, nil
}

// GetAccountMembershipByID retrieves an AccountMembership by ID.
func (c *Client) GetAccountMembershipByID(id uuid.UUID) (*AccountMembership, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/accountmembership/%s", c.BaseURL, id), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Add("X-Api-Key", c.ApiKey)
	req.Header.Add("Content-Type", "application/json")

	res, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK response code %v", res.StatusCode)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var accountMembership AccountMembership
	if err := json.Unmarshal(data, &accountMembership); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	return &accountMembership, nil
}

func (c *Client) GetAccountMembershipsByUserID(userID uuid.UUID) ([]AccountMembership, error) {
	// Create a new URL from the BaseURL of the Client
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	// Specify the path for the endpoint
	u.Path = fmt.Sprintf("/account-memberships/%s", userID)

	// Build a new GET request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Set the headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("X-Api-Key", c.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Decode the response body
	var accountMemberships []AccountMembership
	if err := json.NewDecoder(resp.Body).Decode(&accountMemberships); err != nil {
		return nil, err
	}

	return accountMemberships, nil
}

// AccountMembershipsResponse represents a list of account memberships returned from the server.
type AccountMembershipsResponse struct {
	AccountMemberships []AccountMembership `json:"account_memberships"`
}

// GetAccountMembershipsByAccountID sends a request to the server to retrieve account memberships by account ID.
func (c *Client) GetAccountMembershipsByAccountID(accountID uuid.UUID) (*AccountMembershipsResponse, error) {
	// Prepare the request URL with the account ID
	url := fmt.Sprintf("%s/api/account-memberships/%s", c.BaseURL, accountID.String())

	// Create a new request
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set the necessary headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-API-Key", c.ApiKey)

	// Send the request and handle the response
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server responded with status code %d: %s", resp.StatusCode, body)
	}

	// Decode the response body
	responseData := &AccountMembershipsResponse{}
	if err := json.NewDecoder(resp.Body).Decode(responseData); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	// Return the decoded response data
	return responseData, nil
}

func (c *Client) GetAccountMembershipsByAccountType(accountType string) ([]AccountMembership, error) {
	// Construct the URL
	url := fmt.Sprintf("%s/account_memberships/%s", c.BaseURL, accountType)

	// Create a new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add the authorization header
	req.Header.Add("Authorization", "Bearer "+c.Token)

	// Send the request and get the response
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	// Parse the response body
	var memberships []AccountMembership
	if err := json.NewDecoder(resp.Body).Decode(&memberships); err != nil {
		return nil, err
	}

	return memberships, nil
}

func (c *Client) UpdateAccountMembership(accountMembershipID uuid.UUID, event UpdateAccountMembershipEvent) (AccountMembership, error) {
	// Convert the event to JSON
	jsonEvent, err := json.Marshal(event)
	if err != nil {
		return AccountMembership{}, fmt.Errorf("failed to convert event to JSON: %w", err)
	}

	// Make the request
	req, err := http.NewRequest("PATCH", fmt.Sprintf("%s/account-memberships/%s", c.BaseURL, accountMembershipID), bytes.NewBuffer(jsonEvent))
	if err != nil {
		return AccountMembership{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Set the headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))

	// Send the request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return AccountMembership{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return AccountMembership{}, fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	// Parse the response
	var updatedAccountMembership AccountMembership
	err = json.NewDecoder(resp.Body).Decode(&updatedAccountMembership)
	if err != nil {
		return AccountMembership{}, fmt.Errorf("failed to parse response: %w", err)
	}

	return updatedAccountMembership, nil
}

func (c *Client) DeleteAccountMembership(accountID uuid.UUID, userID uuid.UUID) error {
	// Create a new request
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/account-membership/%s/user/%s", c.BaseURL, accountID, userID), nil)
	if err != nil {
		return err
	}

	// Add headers
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Add("X-Api-Key", c.ApiKey)

	// Execute the request
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// Check for a successful status code
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("server returned status %d: %s", res.StatusCode, body)
	}

	// If everything went fine, return nil
	return nil
}

func (c *Client) ListAccountMemberships(userID uuid.UUID) ([]AccountMembership, error) {
	// Creating URL
	url, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}
	url.Path = path.Join(url.Path, "account-memberships")

	// Creating Request
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+c.Token)
	req.Header.Add("X-Api-Key", c.ApiKey)

	// Sending Request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Reading Response Body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Checking HTTP Response Status Code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d. body: %s", resp.StatusCode, body)
	}

	// Unmarshalling Response Body
	var accountMemberships []AccountMembership
	if err := json.Unmarshal(body, &accountMemberships); err != nil {
		return nil, err
	}

	return accountMemberships, nil
}

func (c *Client) IsUserAMemberOfAccount(userID uuid.UUID, accountID uuid.UUID) (bool, error) {
	endpoint := fmt.Sprintf("%s/api/v1/accounts/%s/members/%s", c.BaseURL, accountID, userID)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return false, fmt.Errorf("creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("X-Api-Key", c.ApiKey)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("unexpected status code: %d. Body: %s", resp.StatusCode, string(body))
	}

	var data struct {
		IsMember bool `json:"is_member"`
	}

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return false, fmt.Errorf("decoding response: %v", err)
	}

	return data.IsMember, nil
}

func (c *Client) GetMembersOfAccount(accountID uuid.UUID) ([]uuid.UUID, error) {
	// Construct the URL for the request
	url, err := url.Parse(fmt.Sprintf("%s/api/v1/accounts/%s/members", c.BaseURL, accountID.String()))
	if err != nil {
		return nil, err
	}

	// Construct the request
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}

	// Add headers
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Add("X-API-KEY", c.ApiKey)

	// Perform the request
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Check the status code
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status code %d", res.StatusCode)
	}

	// Parse the response body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var members []uuid.UUID
	err = json.Unmarshal(body, &members)
	if err != nil {
		return nil, err
	}

	return members, nil
}

// GetRolesForUserInAccount retrieves roles for the given user in the provided account.
func (c *Client) GetRolesForUserInAccount(userID uuid.UUID, accountID uuid.UUID) ([]Role, error) {
	// Create the endpoint url
	endPoint := fmt.Sprintf("%s/api/v1/accounts/%s/users/%s/roles", c.BaseURL, accountID, userID)

	// Prepare a new HTTP request
	req, err := http.NewRequest("GET", endPoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare HTTP request: %v", err)
	}

	// Set headers (including the token and API key)
	req.Header.Add("Authorization", "Bearer "+c.Token)
	req.Header.Add("X-API-Key", c.ApiKey)

	// Send the HTTP request
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer res.Body.Close()

	// Check for HTTP error codes
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned a non-200 status code: %d", res.StatusCode)
	}

	// Decode the HTTP response
	var roles []Role
	if err := json.NewDecoder(res.Body).Decode(&roles); err != nil {
		return nil, fmt.Errorf("failed to decode HTTP response: %v", err)
	}

	return roles, nil
}
