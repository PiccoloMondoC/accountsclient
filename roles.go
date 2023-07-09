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

	"github.com/google/uuid"
)

// Role represents the structure of a role.
type Role struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	CompanyDomainOnly bool      `json:"company_domain_only"`
	IsInternal        bool      `json:"is_internal"`
}

// RoleData represents the input data for a new role.
type RoleData struct {
	Name              string `json:"name"`
	Description       string `json:"description"`
	CompanyDomainOnly bool   `json:"company_domain_only"`
	IsInternal        bool   `json:"is_internal"`
}

type CreateRoleInput struct {
	Name              string
	Description       string
	CompanyDomainOnly bool
	IsInternal        bool
	UserID            uuid.UUID
}

type UpdateRoleInput struct {
	Role   *Role
	UserID uuid.UUID
}

type DeleteRoleInput struct {
	RoleID uuid.UUID
	UserID uuid.UUID
}

type DoesRoleExistInput struct {
	RoleID uuid.UUID
}

type GetRolesByUserIDInput struct {
	UserID uuid.UUID
}

type AssignPermissionToRoleEvent struct {
	RoleID       uuid.UUID `json:"role_id"`
	PermissionID uuid.UUID `json:"permission_id"`
	// You can add other fields if needed
}

type AssignPermissionToRoleInput struct {
	RoleID       uuid.UUID
	PermissionID uuid.UUID
}

type RemovePermissionFromRoleEvent struct {
	RoleID       uuid.UUID `json:"role_id"`
	PermissionID uuid.UUID `json:"permission_id"`
}

type RemovePermissionFromRoleInput struct {
	RoleID       uuid.UUID
	PermissionID uuid.UUID
}

type GetRolesByPermissionIDInput struct {
	PermissionID uuid.UUID
}

type IsPermissionAssignedToRoleInput struct {
	RoleID       uuid.UUID
	PermissionID uuid.UUID
}

func (input *CreateRoleInput) Validate() error {
	if input.Name == "" {
		return errors.New("role name is required")
	}
	return nil
}

// userIDKey is a type for context value for the userID key.
type userIDKey struct{}

func (c *Client) CreateRole(input *CreateRoleInput) (*Role, error) {
	// Validate the input data
	err := input.Validate()
	if err != nil {
		return nil, err
	}

	// Marshal the roleData
	jsonRoleData, err := json.Marshal(RoleData{
		Name:              input.Name,
		Description:       input.Description,
		CompanyDomainOnly: input.CompanyDomainOnly,
		IsInternal:        input.IsInternal,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to marshal roleData: %w", err)
	}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/roles", c.BaseURL), bytes.NewBuffer(jsonRoleData))
	if err != nil {
		return nil, fmt.Errorf("unable to create new request: %w", err)
	}

	// Set the appropriate headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-API-Key", c.ApiKey)

	// Add user id to the request context
	ctx := context.WithValue(req.Context(), userIDKey{}, input.UserID)
	req = req.WithContext(ctx)

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

	// Parse the response
	var createdRole Role
	err = json.NewDecoder(res.Body).Decode(&createdRole)
	if err != nil {
		return nil, fmt.Errorf("unable to parse response: %w", err)
	}

	return &createdRole, nil
}

func (c *Client) GetRoleByID(roleID uuid.UUID) (*Role, error) {
	// Create the URL
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base url: %w", err)
	}
	u.Path = path.Join(u.Path, "api", "roles", roleID.String())

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
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
	var role Role
	err = json.NewDecoder(res.Body).Decode(&role)
	if err != nil {
		return nil, fmt.Errorf("unable to parse response: %w", err)
	}

	return &role, nil
}

func (c *Client) GetRoleByName(roleName string) (*Role, error) {
	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/roles/%s", c.BaseURL, url.PathEscape(roleName)), nil)
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

	// Unmarshal the HTTP response body
	var role Role
	if err := json.NewDecoder(res.Body).Decode(&role); err != nil {
		return nil, fmt.Errorf("unable to decode response: %w", err)
	}

	return &role, nil
}

func (c *Client) UpdateRole(input *UpdateRoleInput) error {
	// Create the payload
	payload := Role{
		ID:                input.Role.ID,
		Name:              input.Role.Name,
		Description:       input.Role.Description,
		CompanyDomainOnly: input.Role.CompanyDomainOnly,
		IsInternal:        input.Role.IsInternal,
	}

	// Marshal the payload
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/roles/%s", c.BaseURL, input.Role.ID), bytes.NewBuffer(jsonPayload))
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

// DeleteRole deletes a role using the API.
func (c *Client) DeleteRole(input *DeleteRoleInput) error {
	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/roles/%s", c.BaseURL, input.RoleID), nil)
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

func (c *Client) ListRoles() ([]Role, error) {
	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/roles", c.BaseURL), nil)
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

	// Unmarshal the response into a slice of roles
	var roles []Role
	if err := json.NewDecoder(res.Body).Decode(&roles); err != nil {
		return nil, fmt.Errorf("unable to decode response body: %w", err)
	}

	return roles, nil
}

func (c *Client) DoesRoleExist(input DoesRoleExistInput) (bool, error) {
	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/roles/%s/does_exist", c.BaseURL, input.RoleID), nil)
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

	// Parse the response body
	var exists struct {
		Exists bool `json:"exists"`
	}
	err = json.NewDecoder(res.Body).Decode(&exists)
	if err != nil {
		return false, fmt.Errorf("unable to parse response body: %w", err)
	}

	return exists.Exists, nil
}

func (c *Client) GetRolesByUserID(input GetRolesByUserIDInput) ([]Role, error) {
	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/users/%s/roles", c.BaseURL, input.UserID), nil)
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

	// Parse the response body
	var roles []Role
	err = json.NewDecoder(res.Body).Decode(&roles)
	if err != nil {
		return nil, fmt.Errorf("unable to parse response body: %w", err)
	}

	return roles, nil
}

// You would also need to define the Validate method for AssignPermissionToRoleEvent
func (input *AssignPermissionToRoleInput) Validate() error {
	// Add validation logic here (e.g. checking if RoleID and PermissionID are not empty)
	if input.RoleID == uuid.Nil {
		return errors.New("RoleID cannot be empty")
	}
	if input.PermissionID == uuid.Nil {
		return errors.New("PermissionID cannot be empty")
	}
	return nil
}

func (c *Client) AssignPermissionToRole(input AssignPermissionToRoleInput) error {
	// Validate the payload
	err := input.Validate()
	if err != nil {
		return err
	}

	// Marshal the payload
	jsonPayload, err := json.Marshal(input)
	if err != nil {
		return err
	}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/roles/%s/permissions/%s", c.BaseURL, input.RoleID, input.PermissionID), bytes.NewBuffer(jsonPayload))
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

func (input *RemovePermissionFromRoleInput) Validate() error {
	if input.RoleID == uuid.Nil {
		return errors.New("role id cannot be empty")
	}

	if input.PermissionID == uuid.Nil {
		return errors.New("permission id cannot be empty")
	}

	return nil
}

func (c *Client) RemovePermissionFromRole(input RemovePermissionFromRoleInput) error {
	// Validate the event
	err := input.Validate()
	if err != nil {
		return err
	}

	// Create the endpoint URL
	endpoint := fmt.Sprintf("%s/api/roles/%s/permissions/%s", c.BaseURL, input.RoleID, input.PermissionID)

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodDelete, endpoint, nil)
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

// GetRolesByPermissionID retrieves all roles associated with a permission identified by its ID.
func (c *Client) GetRolesByPermissionID(input GetRolesByPermissionIDInput) ([]Role, error) {
	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/permissions/%s/roles", c.BaseURL, input.PermissionID), nil)
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

	// Unmarshal the response body into a slice of Role
	var roles []Role
	err = json.NewDecoder(res.Body).Decode(&roles)
	if err != nil {
		return nil, fmt.Errorf("unable to decode response body: %w", err)
	}

	return roles, nil
}

// IsPermissionAssignedToRole checks if a permission is assigned to a role.
func (c *Client) IsPermissionAssignedToRole(input IsPermissionAssignedToRoleInput) (bool, error) {
	// Create the URL for the request
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return false, fmt.Errorf("unable to parse base url: %w", err)
	}
	u.Path = path.Join(u.Path, fmt.Sprintf("/api/roles/%s/permissions/%s", input.RoleID, input.PermissionID))

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
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
	if res.StatusCode == http.StatusNotFound {
		return false, nil
	} else if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return false, fmt.Errorf("unexpected status code: got %v, body: %s", res.StatusCode, body)
	}

	return true, nil
}
