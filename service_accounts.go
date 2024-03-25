package accountslib

import (
	"bytes"
	"context"
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

// ServiceAccount represents the structure of a service account.
type ServiceAccount struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	Secret      string     `db:"secret" json:"secret"`
	ServiceName string     `db:"service_name" json:"service_name"`
	Roles       []string   `json:"roles"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	ExpiresAt   *time.Time `db:"expires_at" json:"expires_at,omitempty"`
}

type RegisterServiceAccountInput struct {
	ServiceName string `json:"name"`
	Role        string `json:"role"`
}

// UpdateServiceAccountInput represents the input data for updating a service account.
type UpdateServiceAccountInput struct {
	ID          uuid.UUID  `json:"id"`
	ServiceName string     `json:"service_name"`
	Roles       []string   `json:"roles"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// AssignRoleInput represents the input data for assigning a role to a service account.
type AssignRoleInput struct {
	ServiceAccountID uuid.UUID `json:"service_account_id"`
	RoleID           uuid.UUID `json:"role_id"`
}

// RemoveRoleInput represents the input data for removing a role from a service account.
type RemoveRoleInput struct {
	ServiceAccountID uuid.UUID `json:"service_account_id"`
	RoleID           uuid.UUID `json:"role_id"`
}

// GetRolesInput represents the input data for retrieving roles of a service account.
type GetRolesInput struct {
	ServiceAccountID uuid.UUID `json:"service_account_id"`
}

// GetServiceAccountsInput represents the input data for retrieving service accounts by a role ID.
type GetServiceAccountsInput struct {
	RoleID uuid.UUID `json:"role_id"`
}

// RoleAssignmentInput represents the input data for checking a role assignment.
type RoleAssignmentInput struct {
	ServiceAccountID uuid.UUID `json:"service_account_id"`
	RoleID           uuid.UUID `json:"role_id"`
}

// RegisterServiceAccount registers a new service account with the provided name and roles.
func (c *Client) RegisterServiceAccount(ctx context.Context, input RegisterServiceAccountInput) (*ServiceAccount, error) {
	requestURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	requestURL.Path = path.Join(requestURL.Path, "/api/v1/service_account/")

	payload, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", requestURL.String(), bytes.NewBuffer(payload))
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

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to register service account: status code %d, message: %s", resp.StatusCode, string(bodyBytes))
	}

	var registeredServiceAccount ServiceAccount
	err = json.NewDecoder(resp.Body).Decode(&registeredServiceAccount)
	if err != nil {
		return nil, err
	}

	return &registeredServiceAccount, nil
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
func (c *Client) UpdateServiceAccount(input UpdateServiceAccountInput) error {
	serviceAccount := &ServiceAccount{
		ID:          input.ID,
		ServiceName: input.ServiceName,
		Roles:       input.Roles,
		ExpiresAt:   input.ExpiresAt,
	}

	// Validate the input
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

func (c *Client) AssignRoleToServiceAccount(input AssignRoleInput) error {
	// Marshal the payload
	jsonPayload, err := json.Marshal(input)
	if err != nil {
		return err
	}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/service-accounts/%s/roles", c.BaseURL, input.ServiceAccountID), bytes.NewBuffer(jsonPayload))
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
func (c *Client) RemoveRoleFromServiceAccount(input RemoveRoleInput) error {
	// Create the request URL
	requestURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return err
	}

	requestURL.Path = path.Join(requestURL.Path, fmt.Sprintf("/api/service_accounts/%s/roles/%s", input.ServiceAccountID, input.RoleID))

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
func (c *Client) GetRolesByServiceAccountID(input GetRolesInput) ([]Role, error) {
	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/serviceaccounts/%s/roles", c.BaseURL, input.ServiceAccountID), nil)
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

func (c *Client) GetServiceAccountsByRoleID(input GetServiceAccountsInput) ([]ServiceAccount, error) {
	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/roles/%s/service-accounts", c.BaseURL, input.RoleID.String()), nil)
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

func (c *Client) IsRoleAssignedToServiceAccount(input RoleAssignmentInput) (bool, error) {
	// Construct the request URL
	reqURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return false, fmt.Errorf("invalid base URL: %w", err)
	}

	reqURL.Path = path.Join(reqURL.Path, "api", "service_accounts", input.ServiceAccountID.String(), "roles", input.RoleID.String())

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
