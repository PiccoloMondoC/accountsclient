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

// Celebrity represents the structure of an artist, band, sports personality, or other public figure.
type Celebrity struct {
	ID            uuid.UUID `json:"id"`
	UserAccountID uuid.UUID `json:"user_account_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CreateCelebrityAccountInput represents the information needed to create a celebrity account.
type CreateCelebrityAccountInput struct {
	UserID        uuid.UUID `json:"user_id"`
	CelebrityName string    `json:"celebrity_name"`
}

// CreateCelebrityAccountResponse represents the response from creating a celebrity account.
type CreateCelebrityAccountResponse struct {
	CelebrityID uuid.UUID `json:"celebrity_id"`
	Status      string    `json:"status"`
	Message     string    `json:"message,omitempty"`
}

type UpdateCelebrityAccountEvent struct {
	CelebrityID uuid.UUID `json:"celebrity_id"`
	UserID      uuid.UUID `json:"user_id,omitempty"`
	NewName     string    `json:"new_name,omitempty"`
}

// UpdateMemberRoleInCelebrityAccountEvent represents the structure of an update member role in celebrity account event.
type UpdateMemberRoleInCelebrityAccountEvent struct {
	CelebrityID uuid.UUID `json:"celebrity_id,omitempty"`
	UserID      uuid.UUID `json:"user_id,omitempty"`
	NewRole     string    `json:"new_role,omitempty"`
}
type AddMemberToCelebrityAccountInput struct {
	UserID      uuid.UUID `json:"user_id"`
	RoleID      uuid.UUID `json:"role_id"`
	CelebrityID uuid.UUID `json:"celebrity_id"`
}

// CreateCelebrityAccount creates a new celebrity account.
func (c *Client) CreateCelebrityAccount(input CreateCelebrityAccountInput) (*Celebrity, error) {
	requestURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	requestURL.Path = path.Join(requestURL.Path, "/api/v1/celebrity/")

	payload, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", requestURL.String(), bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("X-Api-Key", c.ApiKey)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create celebrity account: %s", string(bodyBytes))
	}

	var newCelebrityAccount Celebrity
	err = json.NewDecoder(resp.Body).Decode(&newCelebrityAccount)
	if err != nil {
		return nil, err
	}

	return &newCelebrityAccount, nil
}

// GetCelebrityAccountByID fetches celebrity account data by ID from the API.
func (c *Client) GetCelebrityAccountByID(celebrityID uuid.UUID) (*Celebrity, error) {
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, "celebrities", celebrityID.String())
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("X-API-Key", c.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request error: got status %d with message %s", resp.StatusCode, string(bodyBytes))
	}

	var celebrity Celebrity
	if err = json.NewDecoder(resp.Body).Decode(&celebrity); err != nil {
		return nil, err
	}

	return &celebrity, nil
}

// GetCelebrityAccountsByUserID sends a GET request to the server to retrieve celebrity accounts by user ID.
func (c *Client) GetCelebrityAccountsByUserID(userID uuid.UUID) ([]Celebrity, error) {
	// Construct the URL
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, fmt.Sprintf("celebrity/accounts/%s", userID))

	// Create a new request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Add necessary headers to the request
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("X-API-KEY", c.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check the status code
	if resp.StatusCode != http.StatusOK {
		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("server responded with status code %d: %s", resp.StatusCode, string(body))
	}

	// Decode the response body
	var celebrities []Celebrity
	if err := json.NewDecoder(resp.Body).Decode(&celebrities); err != nil {
		return nil, err
	}

	return celebrities, nil
}

func (e *UpdateCelebrityAccountEvent) Validate() error {
	if e.CelebrityID == uuid.Nil {
		return errors.New("missing CelebrityID")
	}
	if e.NewName == "" {
		return errors.New("missing new name")
	}
	return nil
}

func (c *Client) UpdateCelebrityAccount(event *UpdateCelebrityAccountEvent) (*Celebrity, error) {
	// First, validate the event
	if err := event.Validate(); err != nil {
		return nil, err
	}

	// Construct the URL for the request
	url, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}
	url.Path = path.Join(url.Path, "celebrity", event.CelebrityID.String())

	// Create the JSON body from the event
	body, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	// Create the request
	req, err := http.NewRequest(http.MethodPut, url.String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	// Add necessary headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-Api-Key", c.ApiKey)

	// Send the request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// If the status code is not 200, something went wrong
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(respBody))
	}

	// Decode the response body into a Celebrity object
	var updatedCelebrity Celebrity
	if err := json.Unmarshal(respBody, &updatedCelebrity); err != nil {
		return nil, err
	}

	return &updatedCelebrity, nil
}

func (c *Client) DeleteCelebrityAccount(userID uuid.UUID, celebrityID uuid.UUID) error {
	// Create the url
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, "celebrity_account")

	// Add the user and celebrity IDs as query parameters
	q := u.Query()
	q.Set("userID", userID.String())
	q.Set("celebrityID", celebrityID.String())
	u.RawQuery = q.Encode()

	// Create the request
	req, err := http.NewRequest(http.MethodDelete, u.String(), nil)
	if err != nil {
		return err
	}

	// Add authorization headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))

	// Send the request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("could not delete celebrity account: %v, %s", resp.Status, body)
	}

	// Unmarshal the response body
	var response CreateCelebrityAccountResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("could not parse response: %v", err)
	}

	if response.Status != "success" {
		return fmt.Errorf("failed to delete celebrity account: %s", response.Message)
	}

	return nil
}

func (c *Client) ListCelebrityAccounts() ([]Celebrity, error) {
	// construct the url
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "celebrities")

	// create the request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	// set the headers
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-API-Key", c.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	// send the request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// check the status code
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unexpected status: " + resp.Status)
	}

	// decode the response body
	var celebrities []Celebrity
	err = json.NewDecoder(resp.Body).Decode(&celebrities)
	if err != nil {
		return nil, err
	}

	return celebrities, nil
}

// AddMemberToCelebrityAccount adds a new member to a celebrity account.
func (c *Client) AddMemberToCelebrityAccount(input AddMemberToCelebrityAccountInput) error {
	// Create the URL
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, "memberships")

	// Create the request body
	reqBody := &AccountLinkRequest{
		UserID:      input.UserID,
		AccountType: "celebrity",
		AccountID:   input.CelebrityID,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	// Create the request
	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	// Set the headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	// Make the request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode >= 400 {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.New(string(bodyBytes))
	}

	return nil
}

func (c *Client) RemoveMemberFromCelebrityAccount(celebrityID uuid.UUID, userID uuid.UUID) error {
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return err
	}

	u.Path = path.Join(u.Path, "celebrities", celebrityID.String(), "members", userID.String())

	req, err := http.NewRequest(http.MethodDelete, u.String(), nil)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Add("x-api-key", c.ApiKey)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) GetMembersOfCelebrityAccount(celebrityID uuid.UUID) ([]AccountMembership, error) {
	// Build the request URL
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, fmt.Sprintf("celebrity/%s/members", celebrityID))

	// Create the request
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Set the request headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("X-API-Key", c.ApiKey)

	// Send the request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, errors.New(string(bodyBytes))
	}

	// Decode the response body
	var memberships []AccountMembership
	err = json.NewDecoder(resp.Body).Decode(&memberships)
	if err != nil {
		return nil, err
	}

	return memberships, nil
}

func (c *Client) UpdateMemberRoleInCelebrityAccount(e *UpdateMemberRoleInCelebrityAccountEvent) error {
	// Step 1: Serialize the data to JSON
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}

	// Step 2: Create a new HTTP request
	req, err := http.NewRequest(http.MethodPut, c.BaseURL+"/celebrities/memberships", bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	// Step 3: Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	// Step 4: Send the request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Step 5: Check the HTTP status code
	if resp.StatusCode != http.StatusOK {
		// Read the response body
		body, _ := io.ReadAll(resp.Body)

		// Create an error message
		errMsg := fmt.Sprintf("HTTP request failed with status code %d and body %s", resp.StatusCode, string(body))

		// Return an error
		return errors.New(errMsg)
	}

	return nil
}
