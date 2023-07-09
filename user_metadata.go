// sky-accounts/pkg/clientlib/accountslib/user_metadata.go
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

// UserMetadata represents the structure of a user metadata.
type UserMetadata struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	KeyID     uuid.UUID `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// UserMetadataUpdate represents the input data for updating user metadata.
type UserMetadataUpdate struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`
	Key    string    `json:"key"`
	Value  string    `json:"value"`
}

func (u *UserMetadata) Validate() error {
	// Ensure the UserID is not nil
	if u.UserID == uuid.Nil {
		return errors.New("user ID is required")
	}

	// Ensure the KeyID is not nil
	if u.KeyID == uuid.Nil {
		return errors.New("key ID is required")
	}

	// Ensure the value is not empty
	if u.Value == "" {
		return errors.New("value is required")
	}

	return nil
}

func (c *Client) CreateUserMetadata(metadata *UserMetadata) error {
	// Validate the input
	err := metadata.Validate()
	if err != nil {
		return err
	}

	// Create the payload
	payload := UserMetadata{
		ID:        metadata.ID,
		UserID:    metadata.UserID,
		KeyID:     metadata.KeyID,
		Value:     metadata.Value,
		CreatedAt: metadata.CreatedAt,
		UpdatedAt: metadata.UpdatedAt,
	}

	// Marshal the payload
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/user-metadata", c.BaseURL), bytes.NewBuffer(jsonPayload))
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
	if res.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("unexpected status code: got %v, body: %s", res.StatusCode, body)
	}

	return nil
}

// GetUserMetadataByID retrieves user metadata by ID
func (c *Client) GetUserMetadataByID(metadataID uuid.UUID) (*UserMetadata, error) {
	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/user_metadata/%s", c.BaseURL, metadataID), nil)
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

	// Decode the response body
	var metadata UserMetadata
	if err := json.NewDecoder(res.Body).Decode(&metadata); err != nil {
		return nil, fmt.Errorf("unable to decode response body: %w", err)
	}

	return &metadata, nil
}

// GetUserMetadataByUserID fetches a user's metadata from the server.
func (c *Client) GetUserMetadataByUserID(userID uuid.UUID) (*[]UserMetadata, error) {
	// Validate the user ID
	if userID == uuid.Nil {
		return nil, errors.New("invalid user ID")
	}

	// Create a new HTTP request
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse base url: %w", err)
	}
	u.Path = path.Join(u.Path, fmt.Sprintf("api/users/%s/metadata", userID))
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
	var metadata []UserMetadata
	err = json.NewDecoder(res.Body).Decode(&metadata)
	if err != nil {
		return nil, fmt.Errorf("unable to decode response: %w", err)
	}

	return &metadata, nil
}

func (c *Client) GetUserMetadataByKey(userID, key string) (*UserMetadata, error) {
	// Construct the URL
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	u.Path = path.Join(u.Path, "api/usermetadata", userID, key)

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

	// Decode the response body
	var metadata UserMetadata
	if err := json.NewDecoder(res.Body).Decode(&metadata); err != nil {
		return nil, fmt.Errorf("unable to decode response body: %w", err)
	}

	return &metadata, nil
}

// Validate validates the UserMetadataUpdate fields.
func (u *UserMetadataUpdate) Validate() error {
	if u.ID == uuid.Nil {
		return errors.New("ID field is required")
	}
	if u.UserID == uuid.Nil {
		return errors.New("UserID field is required")
	}
	if u.Key == "" {
		return errors.New("key field is required")
	}
	if u.Value == "" {
		return errors.New("value field is required")
	}
	return nil
}

// UpdateUserMetadata sends an HTTP request to update user metadata.
func (c *Client) UpdateUserMetadata(userMetadata *UserMetadataUpdate) error {
	// Validate the payload
	err := userMetadata.Validate()
	if err != nil {
		return err
	}

	// Marshal the payload
	jsonPayload, err := json.Marshal(userMetadata)
	if err != nil {
		return err
	}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/users/%s/metadata/%s", c.BaseURL, userMetadata.UserID, userMetadata.ID), bytes.NewBuffer(jsonPayload))
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

// DeleteUserMetadataByID deletes user metadata by its ID.
func (c *Client) DeleteUserMetadataByID(id uuid.UUID) error {
	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/usermetadata/%s", c.BaseURL, id), nil)
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

func (c *Client) DeleteUserMetadataByUserID(userID uuid.UUID) error {
	// Check that provided UUID is not empty
	if userID == uuid.Nil {
		return errors.New("user ID is required")
	}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/users/%s/metadata", c.BaseURL, userID), nil)
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

func (c *Client) DeleteUserMetadataByKey(userID uuid.UUID, key string) error {
	// Check that provided UUID and key are not empty
	if userID == uuid.Nil {
		return errors.New("user ID is required")
	}
	if key == "" {
		return errors.New("metadata key is required")
	}

	// Create a new HTTP request
	endpoint, err := url.Parse(c.BaseURL)
	if err != nil {
		return fmt.Errorf("unable to parse base URL: %w", err)
	}
	endpoint.Path = path.Join(endpoint.Path, fmt.Sprintf("api/users/%s/metadata/%s", userID, key))

	req, err := http.NewRequest(http.MethodDelete, endpoint.String(), nil)
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
