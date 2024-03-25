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

// Enterprise represents the structure of an enterprise.
type Enterprise struct {
	ID            uuid.UUID `json:"id"`
	UserAccountID uuid.UUID `json:"user_account_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateEnterpriseAccountInput struct {
	UserID         uuid.UUID `json:"user_id"`
	EnterpriseName string    `json:"enterprise_name"`
}

// UpdateEnterpriseAccountInput is the data structure for the request to update an enterprise account.
type UpdateEnterpriseAccountInput struct {
	UserID               uuid.UUID `json:"user_id"`
	EnterpriseID         uuid.UUID `json:"enterprise_id"`
	UpdatedUserAccountID uuid.UUID `json:"updated_user_account_id"`
}

type AddMemberToEnterpriseAccountEvent struct {
	UserID       uuid.UUID `json:"user_id"`
	RoleID       uuid.UUID `json:"role_id"`
	EnterpriseID uuid.UUID `json:"enterprise_id"`
}

// AddMemberToEnterpriseAccountInput is the data structure for the request to add a member to an enterprise account.
type AddMemberToEnterpriseAccountInput struct {
	EnterpriseID uuid.UUID `json:"enterprise_id"`
	UserID       uuid.UUID `json:"user_id"`
	RoleID       uuid.UUID `json:"role_id"`
}

type EnterpriseMembers struct {
	Members []AccountMembership `json:"members"`
}

// UpdateMemberRoleRequest represents the request payload to update the role of the member in an enterprise account.
type UpdateMemberRoleInEnterpriseAccountRequest struct {
	UserID       uuid.UUID `json:"user_id"`
	EnterpriseID uuid.UUID `json:"enterprise_id"`
	NewRoleID    uuid.UUID `json:"new_role_id"`
}

func (c *Client) CreateEnterpriseAccount(input CreateEnterpriseAccountInput) (*Enterprise, error) {
	enterpriseID := uuid.New() // generate a new UUID for the enterprise account

	enterprise := &Enterprise{
		ID:            enterpriseID,
		UserAccountID: input.UserID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	jsonData, err := json.Marshal(enterprise)
	if err != nil {
		return nil, fmt.Errorf("error marshaling data: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/enterprise", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating new request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %v", err)
		}
		return nil, fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, string(body))
	}

	var result Enterprise
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &result, nil
}

func (c *Client) GetEnterpriseAccountByID(enterpriseID uuid.UUID) (*Enterprise, error) {
	endpoint := "/enterprise/" + enterpriseID.String()
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("parsing base url: %w", err)
	}

	u.Path = path.Join(u.Path, endpoint)
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-Api-Key", c.ApiKey)
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected status OK, got %v", res.StatusCode)
	}

	var enterprise Enterprise
	err = json.NewDecoder(res.Body).Decode(&enterprise)
	if err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &enterprise, nil
}

func (c *Client) GetEnterpriseAccountsByUserID(userID uuid.UUID) ([]Enterprise, error) {
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, fmt.Sprintf("/enterprise/%s", userID))
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("x-api-key", c.ApiKey)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("unexpected status: %s", resp.Status)
		}
		return nil, fmt.Errorf("unexpected response: %s", string(body))
	}

	var enterprises []Enterprise
	err = json.NewDecoder(resp.Body).Decode(&enterprises)
	if err != nil {
		return nil, err
	}

	return enterprises, nil
}

func (c *Client) UpdateEnterpriseAccount(input UpdateEnterpriseAccountInput) error {
	// Validate input
	if input.UserID == uuid.Nil || input.EnterpriseID == uuid.Nil || input.UpdatedUserAccountID == uuid.Nil {
		return errors.New("invalid input parameters")
	}

	// Prepare data for the PUT request
	jsonData, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("failed to marshal request data: %w", err)
	}

	// Create URL
	relativePath := path.Join("api", "enterprise", input.UserID.String())
	url, err := url.Parse(c.BaseURL)
	if err != nil {
		return fmt.Errorf("failed to parse base URL: %w", err)
	}
	url.Path = path.Join(url.Path, relativePath)

	// Create HTTP request
	req, err := http.NewRequest("PUT", url.String(), bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	// Send request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("received non-OK HTTP status: %s, %s", resp.Status, string(body))
	}

	return nil
}

func (c *Client) DeleteEnterpriseAccount(enterpriseID uuid.UUID) error {
	// Construct the URL
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, "enterprise", enterpriseID.String())

	// Create the request
	req, err := http.NewRequest(http.MethodDelete, u.String(), nil)
	if err != nil {
		return err
	}

	// Add the Authorization header
	req.Header.Add("Authorization", "Bearer "+c.Token)

	// Send the request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check for successful status code
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("DeleteEnterpriseAccount failed: %d %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// EnterpriseAccountsResponse represents the structure of the response for the ListEnterpriseAccounts function.
type EnterpriseAccountsResponse struct {
	EnterpriseAccounts []*Enterprise `json:"enterprise_accounts"`
}

func (c *Client) ListEnterpriseAccounts() ([]*Enterprise, error) {
	// Prepare request
	reqURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	reqURL.Path = path.Join(reqURL.Path, "/enterprise") // replace with actual API endpoint path
	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, err
	}

	// Add headers
	req.Header.Add("Authorization", "Bearer "+c.Token)
	req.Header.Add("X-API-Key", c.ApiKey)
	req.Header.Add("Content-Type", "application/json")

	// Send request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check HTTP response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: got %v", resp.Status)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal response body into target structure
	var enterpriseAccountsResp EnterpriseAccountsResponse
	err = json.Unmarshal(body, &enterpriseAccountsResp)
	if err != nil {
		return nil, err
	}

	return enterpriseAccountsResp.EnterpriseAccounts, nil
}

func (c *Client) AddMemberToEnterpriseAccount(input AddMemberToEnterpriseAccountInput) error {
	// Create the URL for the API endpoint
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return err
	}

	u.Path = path.Join(u.Path, "api", "enterprise", input.EnterpriseID.String(), "member")

	// Create a struct for the request body
	reqBody := AddMemberToEnterpriseAccountEvent{
		UserID:       input.UserID,
		RoleID:       input.RoleID,
		EnterpriseID: input.EnterpriseID,
	}

	// Convert the request body to JSON
	jsonReqBody, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	// Create a new request
	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewBuffer(jsonReqBody))
	if err != nil {
		return err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("x-api-key", c.ApiKey)

	// Send the request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check for a successful status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return errors.New(string(body))
	}

	return nil
}

func (c *Client) RemoveMemberFromEnterpriseAccount(enterpriseID, userID uuid.UUID) error {
	// Prepare the URL
	endpoint, err := url.Parse(c.BaseURL)
	if err != nil {
		return err
	}
	endpoint.Path = path.Join(endpoint.Path, fmt.Sprintf("/api/enterprise/%s/member/%s", enterpriseID, userID))

	// Create the request
	req, err := http.NewRequest(http.MethodDelete, endpoint.String(), nil)
	if err != nil {
		return err
	}

	// Add headers
	req.Header.Add("Authorization", "Bearer "+c.Token)
	req.Header.Add("Content-Type", "application/json")

	// Send the request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return errors.New(string(body))
	}

	return nil
}

// GetMembersOfEnterpriseAccount makes a request to the server to get the members of a given enterprise account.
func (c *Client) GetMembersOfEnterpriseAccount(enterpriseID uuid.UUID) (*EnterpriseMembers, error) {
	if c.BaseURL == "" {
		return nil, errors.New("base URL not set")
	}

	// Build the URL
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %w", err)
	}
	u.Path = path.Join(u.Path, fmt.Sprintf("v1/enterprise/%s/members", enterpriseID))

	// Create a new request
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set the request headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("X-API-Key", c.ApiKey)

	// Send the request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check the status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned non-OK status code: %d", resp.StatusCode)
	}

	// Decode the response body
	var members EnterpriseMembers
	if err := json.NewDecoder(resp.Body).Decode(&members); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return &members, nil
}

func (c *Client) UpdateMemberRoleInEnterpriseAccount(req UpdateMemberRoleInEnterpriseAccountRequest) error {
	url, err := url.Parse(c.BaseURL)
	if err != nil {
		return err
	}
	url.Path = path.Join(url.Path, "your-endpoint") // Replace "your-endpoint" with your actual endpoint.

	jsonReq, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest(http.MethodPut, url.String(), bytes.NewBuffer(jsonReq))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))

	resp, err := c.HttpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return errors.New(string(bodyBytes))
	}

	return nil
}
