// sky-accounts/pkg/clientlib/accountslib/permissions.go
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

	"github.com/google/uuid"
)

// Permission represents the structure of a permission.
type Permission struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

func (c *Client) CreatePermission(permission *Permission) (*Permission, error) {
	// Marshall the Permission struct to JSON
	payloadBuf := new(bytes.Buffer)
	err := json.NewEncoder(payloadBuf).Encode(permission)
	if err != nil {
		return nil, err
	}

	// Construct the URL
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "permissions") // I assumed permissions is the correct endpoint

	// Create a new request
	req, err := http.NewRequest("POST", u.String(), payloadBuf)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token) // Assuming the server uses Bearer token authentication
	req.Header.Set("X-Api-Key", c.ApiKey)              // Assuming the server requires API Key

	// Do the request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check for a successful status code
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, errors.New(string(body))
	}

	// Unmarshal the response into a Permission struct
	var createdPermission Permission
	err = json.NewDecoder(resp.Body).Decode(&createdPermission)
	if err != nil {
		return nil, err
	}

	return &createdPermission, nil
}

func (c *Client) GetPermissionByID(permissionID uuid.UUID) (*Permission, error) {
	// Define the endpoint URL
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %v", err)
	}
	u.Path = path.Join(u.Path, "permissions", permissionID.String())

	// Create the request
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Add the necessary headers to the request
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Add("X-Api-Key", c.ApiKey)
	req.Header.Add("Content-Type", "application/json")

	// Send the request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get permission, status: %d, response: %s", resp.StatusCode, body)
	}

	// Decode the response body into the Permission struct
	var permission Permission
	err = json.NewDecoder(resp.Body).Decode(&permission)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response body: %v", err)
	}

	return &permission, nil
}

func (c *Client) GetPermissionByName(permissionName string) (*Permission, error) {
	// Create new URL
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse base URL: %w", err)
	}

	// Set the endpoint path
	u.Path = path.Join(u.Path, "permissions", permissionName)

	// Prepare request
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to prepare request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("X-API-Key", c.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var permission Permission
	if err := json.NewDecoder(resp.Body).Decode(&permission); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}
	return &permission, nil
}

func (p *Permission) Validate() error {
	// Check if ID is not empty
	if p.ID == uuid.Nil {
		return fmt.Errorf("permission ID cannot be empty")
	}

	// Check if Name is not empty
	if p.Name == "" {
		return fmt.Errorf("permission Name cannot be empty")
	}

	// Add other validation logic if needed...

	return nil
}

func (c *Client) UpdatePermission(permission *Permission) error {
	// Create the payload
	payload := Permission{
		ID:          permission.ID,
		Name:        permission.Name,
		Description: permission.Description,
	}

	// Validate the payload
	err := payload.Validate()
	if err != nil {
		return err
	}

	// Marshal the payload
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/permissions/%s", c.BaseURL, permission.ID), bytes.NewBuffer(jsonPayload))
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

func (c *Client) DeletePermission(permissionID uuid.UUID) error {
	// Validate permissionID
	if permissionID == uuid.Nil {
		return errors.New("invalid permissionID")
	}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/permissions/%s", c.BaseURL, permissionID), nil)
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

func (c *Client) ListPermissions() ([]Permission, error) {
	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/permissions", c.BaseURL), nil)
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
	var permissions []Permission
	if err := json.NewDecoder(res.Body).Decode(&permissions); err != nil {
		return nil, fmt.Errorf("unable to decode response: %w", err)
	}

	return permissions, nil
}

func (c *Client) DoesPermissionExist(permissionID uuid.UUID) (bool, error) {
	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/permissions/%s", c.BaseURL, permissionID.String()), nil)
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
	// For this method, we assume that a status of 200 means the permission exists,
	// a 404 means it does not, and any other status is an error.
	switch res.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		body, _ := io.ReadAll(res.Body)
		return false, fmt.Errorf("unexpected status code: got %v, body: %s", res.StatusCode, body)
	}
}

// GetPermissionsByUserID sends a GET request to retrieve permissions by user ID.
func (c *Client) GetPermissionsByUserID(userID uuid.UUID) ([]Permission, error) {
	// Prepare the API endpoint with the user ID
	endpoint := fmt.Sprintf("%s/api/permissions/user/%s", c.BaseURL, userID.String())

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
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

	// Decode the response body into a slice of Permission
	var permissions []Permission
	err = json.NewDecoder(res.Body).Decode(&permissions)
	if err != nil {
		return nil, fmt.Errorf("unable to decode response body: %w", err)
	}

	return permissions, nil
}

// GetPermissionsByRoleID fetches the permissions associated with the provided role ID.
func (c *Client) GetPermissionsByRoleID(roleID uuid.UUID) ([]Permission, error) {
	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/roles/%s/permissions", c.BaseURL, roleID), nil)
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

	// Decode the response
	var permissions []Permission
	err = json.NewDecoder(res.Body).Decode(&permissions)
	if err != nil {
		return nil, fmt.Errorf("unable to decode response: %w", err)
	}

	return permissions, nil
}
