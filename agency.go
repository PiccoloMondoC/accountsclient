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

// Agency represents the structure of an agency.
type Agency struct {
	ID            uuid.UUID `json:"id"`
	UserAccountID uuid.UUID `json:"user_account_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateAgencyAccountEvent struct {
	UserID     uuid.UUID `json:"user_id"`
	AgencyName string    `json:"agency_name"`
	AgencyID   uuid.UUID `json:"agency_id"`
}

type AddMemberToAgencyAccountEvent struct {
	UserID      uuid.UUID `json:"user_id"`
	AccountType string    `json:"account_type"`
	AccountID   uuid.UUID `json:"account_id"`
	Role        string    `json:"role"`
}

func (c *Client) CreateAgencyAccount(userID uuid.UUID, agencyName string) (*Agency, error) {
	requestURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	requestURL.Path = path.Join(requestURL.Path, "/api/v1/agency/")

	agencyAccountEvent := CreateAgencyAccountEvent{
		UserID:     userID,
		AgencyName: agencyName,
		AgencyID:   uuid.New(), // Generate a new UUID for agency account
	}

	payload, err := json.Marshal(agencyAccountEvent)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", requestURL.String(), bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token)) // This assumes you're using Bearer token authentication
	req.Header.Set("X-Api-Key", c.ApiKey)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create agency account: %s", string(bodyBytes))
	}

	var newAgencyAccount Agency
	err = json.NewDecoder(resp.Body).Decode(&newAgencyAccount)
	if err != nil {
		return nil, err
	}

	return &newAgencyAccount, nil
}

func (c *Client) GetAgencyAccountByID(agencyID uuid.UUID) (*Agency, error) {
	// Build the URL for the request.
	url := fmt.Sprintf("%s/agency/%s", c.BaseURL, agencyID)

	// Create the request.
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add necessary headers.
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-API-Key", c.ApiKey)

	// Send the request and get a response.
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check the status code of the response.
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK status code: %d", resp.StatusCode)
	}

	// Decode the response body into an Agency struct.
	var agency Agency
	if err := json.NewDecoder(resp.Body).Decode(&agency); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return &agency, nil
}

func (c *Client) GetAgencyAccountsByUserID(userID uuid.UUID) ([]Agency, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/users/%s/agencyaccounts", c.BaseURL, userID), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Add("X-Api-Key", c.ApiKey)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var agencies []Agency
	err = json.NewDecoder(resp.Body).Decode(&agencies)
	if err != nil {
		return nil, err
	}

	return agencies, nil
}

func (c *Client) UpdateAgencyAccount(agencyID uuid.UUID, updatedUserAccountID uuid.UUID) error {
	// Create the request URL
	reqURL := fmt.Sprintf("%s/api/v1/agencies/%s", c.BaseURL, agencyID)

	// Create the request body
	requestBody := map[string]uuid.UUID{
		"updatedUserAccountID": updatedUserAccountID,
	}
	jsonRequestBody, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	// Create the request
	req, err := http.NewRequest(http.MethodPut, reqURL, bytes.NewBuffer(jsonRequestBody))
	if err != nil {
		return err
	}

	// Set request headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))

	// Send the request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check for a successful response
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("received non-OK HTTP status: %s", string(bodyBytes))
	}

	return nil
}

func (c *Client) DeleteAgencyAccount(agencyID uuid.UUID) error {
	// Build request URL
	requestURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return fmt.Errorf("error parsing base URL: %w", err)
	}

	requestURL.Path = path.Join(requestURL.Path, fmt.Sprintf("/api/agencies/%s", agencyID))

	// Create new request
	req, err := http.NewRequest(http.MethodDelete, requestURL.String(), nil)
	if err != nil {
		return fmt.Errorf("error creating new request: %w", err)
	}

	// Add authorization header
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-Api-Key", c.ApiKey)

	// Send the request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Handle non-2XX status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("non-2XX status code: %d. Response body: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (c *Client) ListAgencyAccounts(userID uuid.UUID) ([]Agency, error) {
	// Prepare request
	req, err := http.NewRequest("GET", c.BaseURL+"/agency/accounts/"+userID.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("x-api-key", c.ApiKey)

	// Send request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected server response: %v", resp.Status)
	}

	// Parse response data
	var agencyAccounts []Agency
	if err := json.NewDecoder(resp.Body).Decode(&agencyAccounts); err != nil {
		return nil, err
	}

	return agencyAccounts, nil
}

func (c *Client) AddMemberToAgencyAccount(e AddMemberToAgencyAccountEvent) error {
	// First, marshal the input data to JSON
	requestBody, err := json.Marshal(e)
	if err != nil {
		return err
	}

	// Create the request
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v1/accounts/agencies/members", c.BaseURL), bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}

	// Add any necessary headers to the request
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("X-Api-Key", c.ApiKey)

	// Send the request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the HTTP status of the response
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bad status: %s body: %s", resp.Status, string(body))
	}

	return nil
}

func (c *Client) RemoveMemberFromAgencyAccount(userID uuid.UUID, agencyID uuid.UUID) error {
	// Create the endpoint url.
	// Assuming the endpoint is '/agency/{agencyID}/member/{userID}', replace with correct one if different.
	endpoint := fmt.Sprintf("%s/agency/%s/member/%s", c.BaseURL, agencyID, userID)

	// Create a new request.
	req, err := http.NewRequest(http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	// Add necessary headers.
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Token))

	// Send the request.
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the status code and handle errors.
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bad response from server: %s", body)
	}

	return nil
}

func (c *Client) GetMembersOfAgencyAccount(agencyID uuid.UUID) ([]AccountMembership, error) {
	// The endpoint URI should be in a format similar to "/agency/{agencyID}/members"
	requestURL := fmt.Sprintf("%s/agency/%s/members", c.BaseURL, agencyID)

	// Create a new request
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create new request: %w", err)
	}

	// Set the Authorization header
	req.Header.Set("Authorization", "Bearer "+c.Token)

	// Send the request
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer res.Body.Close()

	// Check if the status code indicates success
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("get members of agency account: %v - %s", res.StatusCode, string(body))
	}

	// Decode the response
	var memberships []AccountMembership
	err = json.NewDecoder(res.Body).Decode(&memberships)
	if err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	// Return the list of memberships
	return memberships, nil
}

func (c *Client) UpdateMemberRoleInAgencyAccount(agencyID uuid.UUID, memberID uuid.UUID, newRoleID uuid.UUID) error {
	endpoint := fmt.Sprintf("%s/agencies/%s/members/%s", c.BaseURL, agencyID, memberID)

	updateRoleRequest := map[string]interface{}{
		"role_id": newRoleID,
	}
	jsonValue, _ := json.Marshal(updateRoleRequest)

	req, err := http.NewRequest("PATCH", endpoint, bytes.NewBuffer(jsonValue))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %s: %s", resp.Status, string(bodyBytes))
	}

	return nil
}
