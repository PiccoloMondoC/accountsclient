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
	"strings"
	"time"

	"github.com/google/uuid"
)

// ServiceAccount represents the structure of a service account.
type ServiceAccount struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	Secret      string     `db:"secret" json:"-"`
	ServiceName string     `db:"service_name" json:"service_name"`
	Roles       []string   `json:"roles"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	ExpiresAt   *time.Time `db:"expires_at" json:"expires_at,omitempty"`
}

// ServiceAccountData represents the input data for a new service account.
type ServiceAccountData struct {
	ServiceName string     `json:"service_name"`
	Roles       []string   `json:"roles"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// CreateServiceAccount creates a new service account by making a POST request to the server.
func (c *Client) CreateServiceAccount(serviceAccount *ServiceAccountData) (*ServiceAccount, error) {
	// Validate the input
	if strings.TrimSpace(serviceAccount.ServiceName) == "" {
		return nil, errors.New("service name is required")
	}
	if len(serviceAccount.Roles) == 0 {
		return nil, errors.New("at least one role is required")
	}

	// Marshal the service account data
	jsonPayload, err := json.Marshal(serviceAccount)
	if err != nil {
		return nil, err
	}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/service-accounts", c.BaseURL), bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("unable to create new request: %w", err)
	}

	// Set the appropriate headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-API-Key", c.ApiKey)

	// Send the HTTP request
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to send request: %w", err)
	}
	defer res.Body.Close()

	// Check the status code
	if res.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("unexpected status code: got %v, body: %s", res.StatusCode, body)
	}

	// Decode the response body
	var createdServiceAccount ServiceAccount
	err = json.NewDecoder(res.Body).Decode(&createdServiceAccount)
	if err != nil {
		return nil, fmt.Errorf("unable to decode response body: %w", err)
	}

	// Return the created service account
	return &createdServiceAccount, nil
}

// GetServiceAccountByID sends a GET request to the server to retrieve a service account by its ID
func (c *Client) GetServiceAccountByID(id uuid.UUID) (*ServiceAccount, error) {
	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/serviceaccounts/%s", c.BaseURL, id), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create new request: %w", err)
	}

	// Set the appropriate headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-API-Key", c.ApiKey)

	// Send the HTTP request
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to send request: %w", err)
	}
	defer res.Body.Close()

	// Check the status code
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("unexpected status code: got %v, body: %s", res.StatusCode, body)
	}

	// Parse the response
	var serviceAccount ServiceAccount
	if err := json.NewDecoder(res.Body).Decode(&serviceAccount); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &serviceAccount, nil
}

func (c *Client) GetServiceAccountByName(serviceName string) (*ServiceAccount, error) {
	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/service_accounts/%s", c.BaseURL, url.PathEscape(serviceName)), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create new request: %w", err)
	}

	// Set the appropriate headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-API-Key", c.ApiKey)

	// Send the HTTP request
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to send request: %w", err)
	}
	defer res.Body.Close()

	// Check the status code
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("unexpected status code: got %v, body: %s", res.StatusCode, body)
	}

	// Read the response body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body: %w", err)
	}

	// Unmarshal the response body
	var serviceAccount ServiceAccount
	if err = json.Unmarshal(body, &serviceAccount); err != nil {
		return nil, fmt.Errorf("unable to unmarshal response body: %w", err)
	}

	return &serviceAccount, nil
}

// UpdateServiceAccount sends a request to update a service account.
func (c *Client) UpdateServiceAccount(serviceAccount *ServiceAccount) error {
	// Validate the input
	if serviceAccount == nil {
		return errors.New("serviceAccount cannot be nil")
	}
	if serviceAccount.ID == uuid.Nil {
		return errors.New("service account ID is required")
	}
	if serviceAccount.ServiceName == "" {
		return errors.New("service account name is required")
	}
	if len(serviceAccount.Roles) == 0 {
		return errors.New("at least one role is required for the service account")
	}

	// Marshal the serviceAccount to JSON
	jsonPayload, err := json.Marshal(serviceAccount)
	if err != nil {
		return fmt.Errorf("unable to marshal service account: %w", err)
	}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/service-accounts/%s", c.BaseURL, serviceAccount.ID), bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("unable to create new request: %w", err)
	}

	// Set the appropriate headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-API-Key", c.ApiKey)

	// Send the HTTP request
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to send request: %w", err)
	}
	defer res.Body.Close()

	// Check the status code
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("unexpected status code: got %v, body: %s", res.StatusCode, body)
	}

	return nil
}

// DeleteServiceAccount deletes a service account by its ID
func (c *Client) DeleteServiceAccount(serviceAccountID uuid.UUID) error {
	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/service-accounts/%s", c.BaseURL, serviceAccountID), nil)
	if err != nil {
		return fmt.Errorf("unable to create new request: %w", err)
	}

	// Set the appropriate headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-API-Key", c.ApiKey)

	// Send the HTTP request
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to send request: %w", err)
	}
	defer res.Body.Close()

	// Check the status code
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("unexpected status code: got %v, body: %s", res.StatusCode, body)
	}

	return nil
}

func (c *Client) ListServiceAccounts() ([]ServiceAccount, error) {
	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/service_accounts", c.BaseURL), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create new request: %w", err)
	}

	// Set the appropriate headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-API-Key", c.ApiKey)

	// Send the HTTP request
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to send request: %w", err)
	}
	defer res.Body.Close()

	// Check the status code
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("unexpected status code: got %v, body: %s", res.StatusCode, body)
	}

	// Parse the response
	var serviceAccounts []ServiceAccount
	if err = json.NewDecoder(res.Body).Decode(&serviceAccounts); err != nil {
		return nil, fmt.Errorf("unable to decode response: %w", err)
	}

	return serviceAccounts, nil
}

func (c *Client) AssignRoleToServiceAccount(serviceAccountID, roleID uuid.UUID) error {
	// Create the payload
	payload := map[string]uuid.UUID{
		"service_account_id": serviceAccountID,
		"role_id":            roleID,
	}

	// Marshal the payload
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/service-accounts/%s/roles", c.BaseURL, serviceAccountID), bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("unable to create new request: %w", err)
	}

	// Set the appropriate headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-API-Key", c.ApiKey)

	// Send the HTTP request
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to send request: %w", err)
	}
	defer res.Body.Close()

	// Check the status code
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("unexpected status code: got %v, body: %s", res.StatusCode, body)
	}

	return nil
}

// RemoveRoleFromServiceAccount removes a role from a service account.
func (c *Client) RemoveRoleFromServiceAccount(serviceAccountID uuid.UUID, roleID uuid.UUID) error {
	// Create the request URL
	requestURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return err
	}

	requestURL.Path = path.Join(requestURL.Path, fmt.Sprintf("/api/service_accounts/%s/roles/%s", serviceAccountID, roleID))

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodDelete, requestURL.String(), nil)
	if err != nil {
		return fmt.Errorf("unable to create new request: %w", err)
	}

	// Set the appropriate headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-API-Key", c.ApiKey)

	// Send the HTTP request
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to send request: %w", err)
	}
	defer res.Body.Close()

	// Check the status code
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("unexpected status code: got %v, body: %s", res.StatusCode, body)
	}

	return nil
}

// GetRolesByServiceAccountID retrieves roles associated with a specific service account ID
func (c *Client) GetRolesByServiceAccountID(serviceAccountID uuid.UUID) ([]Role, error) {
	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/serviceaccounts/%s/roles", c.BaseURL, serviceAccountID), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create new request: %w", err)
	}

	// Set the appropriate headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-API-Key", c.ApiKey)

	// Send the HTTP request
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to send request: %w", err)
	}
	defer res.Body.Close()

	// Check the status code
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("unexpected status code: got %v, body: %s", res.StatusCode, body)
	}

	// Parse the response
	var roles []Role
	err = json.NewDecoder(res.Body).Decode(&roles)
	if err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return roles, nil
}

func (c *Client) GetServiceAccountsByRoleID(roleID uuid.UUID) ([]ServiceAccount, error) {
	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/roles/%s/service-accounts", c.BaseURL, roleID.String()), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create new request: %w", err)
	}

	// Set the appropriate headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-API-Key", c.ApiKey)

	// Send the HTTP request
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to send request: %w", err)
	}

	defer res.Body.Close()

	// Check the status code
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("unexpected status code: got %v, body: %s", res.StatusCode, body)
	}

	var serviceAccounts []ServiceAccount

	// Parse the response
	if err := json.NewDecoder(res.Body).Decode(&serviceAccounts); err != nil {
		return nil, fmt.Errorf("unable to parse response: %w", err)
	}

	return serviceAccounts, nil
}

func (c *Client) IsRoleAssignedToServiceAccount(serviceAccountID, roleID uuid.UUID) (bool, error) {
	// Construct the request URL
	reqURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return false, fmt.Errorf("invalid base URL: %w", err)
	}

	reqURL.Path = path.Join(reqURL.Path, "api", "service_accounts", serviceAccountID.String(), "roles", roleID.String())

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return false, fmt.Errorf("unable to create new request: %w", err)
	}

	// Set the appropriate headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-API-Key", c.ApiKey)

	// Send the HTTP request
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("unable to send request: %w", err)
	}
	defer res.Body.Close()

	// Check the status code
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return false, fmt.Errorf("unexpected status code: got %v, body: %s", res.StatusCode, body)
	}

	// Unmarshal the response body
	var result struct {
		IsRoleAssigned bool `json:"is_role_assigned"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("error decoding response body: %w", err)
	}

	return result.IsRoleAssigned, nil
}

func (t *Token) Validate() error {
	if t.UserID == uuid.Nil {
		return errors.New("user ID is required")
	}
	if t.Scope == "" {
		return errors.New("scope is required")
	}
	if t.Expiry.Before(time.Now()) {
		return errors.New("expiry must be a future time")
	}
	return nil
}
