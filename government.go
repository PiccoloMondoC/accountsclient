package accountslib

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/google/uuid"
)

// Government represents the structure of a government agency.
type Government struct {
	ID            uuid.UUID `json:"id"`
	UserAccountID uuid.UUID `json:"user_account_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateGovernmentAccountInput struct {
	UserID         uuid.UUID `json:"user_id"`
	GovernmentName string    `json:"government_name"`
	GovernmentID   uuid.UUID `json:"government_id"`
}

type UpdateGovernmentAccountEvent struct {
	UserID         uuid.UUID `json:"user_id"`
	GovernmentName string    `json:"government_name"`
	GovernmentID   uuid.UUID `json:"government_id"`
}

type AddMemberToGovernmentAccountEvent struct {
	UserID       uuid.UUID `json:"userId"`
	RoleID       uuid.UUID `json:"roleId"`
	GovernmentID uuid.UUID `json:"governmentId"`
}

type UpdateMemberRoleInGovernmentAccountEvent struct {
	UserID       uuid.UUID `json:"user_id"`
	GovernmentID uuid.UUID `json:"government_id"`
	NewRoleID    uuid.UUID `json:"new_role_id"`
}

// CreateGovernmentAccount makes a POST request to create a government account
func (c *Client) CreateGovernmentAccount(input CreateGovernmentAccountInput) (*Government, error) {
	url, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	url.Path = path.Join(url.Path, "government") // Replace "government" with the actual path

	requestBody, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url.String(), bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response from server: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var govAccount Government
	err = json.Unmarshal(body, &govAccount)
	if err != nil {
		return nil, err
	}

	return &govAccount, nil
}

// GetGovernmentAccountByID fetches a government account by its ID.
func (c *Client) GetGovernmentAccountByID(governmentID uuid.UUID) (*Government, error) {
	// Generate the URL for the HTTP request
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %v", err)
	}
	u.Path = path.Join(u.Path, "government", governmentID.String())

	// Create the HTTP request
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %v", err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Add("X-Api-Key", c.ApiKey)

	// Send the HTTP request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check the HTTP response status
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected response status %v: %v", resp.StatusCode, string(bodyBytes))
	}

	// Parse the HTTP response body
	var government Government
	if err := json.NewDecoder(resp.Body).Decode(&government); err != nil {
		return nil, fmt.Errorf("could not parse response: %v", err)
	}

	return &government, nil
}

func (c *Client) GetGovernmentAccountsByUserID(userID uuid.UUID) ([]Government, error) {
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, "government", userID.String())

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("X-Api-Key", c.ApiKey)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var governmentAccounts []Government
	if err := json.Unmarshal(body, &governmentAccounts); err != nil {
		return nil, err
	}

	return governmentAccounts, nil
}

// Validate checks if the UpdateGovernmentAccountEvent is valid
func (e *UpdateGovernmentAccountEvent) Validate() error {
	if e.UserID == uuid.Nil {
		return fmt.Errorf("user id is required")
	}
	if e.GovernmentName == "" {
		return fmt.Errorf("government name is required")
	}
	if e.GovernmentID == uuid.Nil {
		return fmt.Errorf("government id is required")
	}

	return nil
}

func (c *Client) UpdateGovernmentAccount(userID uuid.UUID, governmentID uuid.UUID, newName string) error {
	event := &UpdateGovernmentAccountEvent{
		UserID:         userID,
		GovernmentName: newName,
		GovernmentID:   governmentID,
	}

	// Validate the event
	if err := event.Validate(); err != nil {
		return err
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Create URL
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return err
	}

	u.Path = path.Join(u.Path, "government", event.GovernmentID.String())

	req, err := http.NewRequest(http.MethodPut, u.String(), bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-Api-Key", c.ApiKey)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update government account, received status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (c *Client) DeleteGovernmentAccount(accountID uuid.UUID) error {
	// Prepare the request URL
	requestURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return err
	}
	requestURL.Path = path.Join(requestURL.Path, "government", accountID.String())

	// Prepare the request
	req, err := http.NewRequest(http.MethodDelete, requestURL.String(), nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-Api-Key", c.ApiKey)

	// Send the request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return errors.New(string(bodyBytes))
	}

	return nil
}

func (c *Client) ListGovernmentAccounts() ([]Government, error) {
	// Create the request URL from BaseURL
	requestURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}
	requestURL.Path = path.Join(requestURL.Path, "/api/government_accounts")

	// Create new HTTP request
	req, err := http.NewRequest(http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return nil, err
	}

	// Set request headers for authentication
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("X-API-KEY", c.ApiKey)

	// Send HTTP request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check HTTP response status code
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("expected status 200 OK, got %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response data
	var governmentAccounts []Government
	err = json.NewDecoder(resp.Body).Decode(&governmentAccounts)
	if err != nil {
		return nil, err
	}

	return governmentAccounts, nil
}

func (c *Client) AddMemberToGovernmentAccount(governmentID uuid.UUID, userID uuid.UUID, roleID uuid.UUID) error {
	// Create the request body
	reqBody := AddMemberToGovernmentAccountEvent{
		UserID:       userID,
		RoleID:       roleID,
		GovernmentID: governmentID,
	}

	// Marshal the request body to JSON
	jsonReqBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Construct the URL
	requestURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return fmt.Errorf("failed to parse base url: %w", err)
	}
	requestURL.Path = path.Join(requestURL.Path, "government", "addMember")

	// Create the HTTP request
	req, err := http.NewRequest(http.MethodPost, requestURL.String(), bytes.NewBuffer(jsonReqBody))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add necessary headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	// Send the request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check if the response status is successful
	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}
		return fmt.Errorf("received non-OK HTTP status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (c *Client) RemoveMemberFromGovernmentAccount(userID uuid.UUID, governmentID uuid.UUID) error {
	endpoint, err := url.Parse(c.BaseURL)
	if err != nil {
		return err
	}
	endpoint.Path = path.Join(endpoint.Path, fmt.Sprintf("/government/%s/member/%s", governmentID, userID))

	req, err := http.NewRequest(http.MethodDelete, endpoint.String(), nil)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Add("X-Api-Key", c.ApiKey)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error removing member from government account: %s", err)
		}
		return fmt.Errorf("error removing member from government account: %s", bodyBytes)
	}

	return nil
}

func (c *Client) GetMembersOfGovernmentAccount(governmentID uuid.UUID) ([]AccountMembership, error) {
	// Prepare the request URL
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, "api/government/members")

	// Create a request body
	reqBody := map[string]string{
		"government_id": governmentID.String(),
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	// Add the necessary headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("X-Api-Key", c.ApiKey)

	// Send the HTTP request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		// You may want to add more error handling here to deal with different HTTP status codes
		return nil, fmt.Errorf("API request failed with status code %d", resp.StatusCode)
	}

	// Parse the response body
	var memberships []AccountMembership
	err = json.NewDecoder(resp.Body).Decode(&memberships)
	if err != nil {
		return nil, err
	}

	return memberships, nil
}

func (event *UpdateMemberRoleInGovernmentAccountEvent) Validate() error {
	if event.UserID == uuid.Nil {
		return errors.New("user_id cannot be empty")
	}
	if event.GovernmentID == uuid.Nil {
		return errors.New("government_id cannot be empty")
	}
	if event.NewRoleID == uuid.Nil {
		return errors.New("new_role_id cannot be empty")
	}
	return nil
}

func (c *Client) UpdateMemberRoleInGovernmentAccount(event UpdateMemberRoleInGovernmentAccountEvent) error {
	// Check the validity of the event.
	if err := event.Validate(); err != nil {
		return err
	}

	// Build request URL from base URL.
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, "government", event.GovernmentID.String(), "users", event.UserID.String())

	// Prepare request body.
	requestBody, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Build request.
	req, err := http.NewRequest(http.MethodPut, u.String(), bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}

	// Add necessary headers.
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("X-Api-Key", c.ApiKey)

	// Perform the request.
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the response.
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.New(string(body))
	}

	return nil
}
