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

// MetadataKey represents a metadata key record in the database
type MetadataKey struct {
	ID      uuid.UUID `json:"id"`
	KeyName string    `json:"key_name"`
}

type CreateMetadataKeyInput struct {
	KeyName string `json:"key_name"`
}

type GetMetadataKeyByIDInput struct {
	MetadataKeyID uuid.UUID `json:"metadata_key_id"`
	UserID        uuid.UUID `json:"user_id"`
}

type GetMetadataKeyByKeyNameInput struct {
	KeyName string `json:"key_name"`
}

type UpdateMetadataKeyInput struct {
	ID      uuid.UUID `json:"id"`
	KeyName string    `json:"key_name"`
}

type DeleteMetadataKeyInput struct {
	ID uuid.UUID `json:"id"`
}

type MetadataKeyExistsInput struct {
	KeyName string
}

func (c *Client) CreateMetadataKey(input CreateMetadataKeyInput) (*MetadataKey, error) {
	// Prepare the metadata key payload
	payload := MetadataKey{
		ID:      uuid.New(),    // Generate a new UUID
		KeyName: input.KeyName, // Assign the provided key name
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshalling payload: %w", err)
	}

	// Prepare the request
	req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/metadatakeys", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-Api-Key", c.ApiKey)

	// Send the request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse the response
	var createdMetadataKey MetadataKey
	if err := json.NewDecoder(resp.Body).Decode(&createdMetadataKey); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &createdMetadataKey, nil
}

func (c *Client) GetMetadataKeyByID(input GetMetadataKeyByIDInput) (*UserMetadata, error) {
	endpoint := fmt.Sprintf("%s/metadata-keys/%s", c.BaseURL, metadataKeyID)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-API-KEY", c.ApiKey)
	req.Header.Set("UserID", userID.String()) // Pass UserID as a header

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %v", err)
		}
		return nil, fmt.Errorf("server responded with code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var metadataKey UserMetadata
	err = json.NewDecoder(resp.Body).Decode(&metadataKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response body: %v", err)
	}

	return &metadataKey, nil
}

// GetMetadataKeyByKeyName sends a GET request to the /metadata-keys/{keyName} endpoint
// of the account service to retrieve the metadata key by its key name.
func (c *Client) GetMetadataKeyByKeyName(input GetMetadataKeyByKeyNameInput) (*MetadataKey, error) {
	// Build the URL for the request
	reqURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	reqURL.Path = path.Join(reqURL.Path, "metadata-keys", keyName)

	// Create the HTTP request
	req, err := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+c.Token)
	req.Header.Add("X-Api-Key", c.ApiKey)

	// Send the request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse the response body
	var metadataKey MetadataKey
	err = json.NewDecoder(resp.Body).Decode(&metadataKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return &metadataKey, nil
}

func (c *Client) UpdateMetadataKey(input UpdateMetadataKeyInput) (*MetadataKey, error) {
	// Marshal MetadataKey to JSON
	data, err := json.Marshal(metadataKey)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal MetadataKey object: %v", err)
	}

	// Prepare HTTP request
	reqURL, err := url.Parse(path.Join(c.BaseURL, "metadatakey", metadataKey.ID.String()))
	if err != nil {
		return nil, fmt.Errorf("unable to parse URL: %v", err)
	}

	req, err := http.NewRequest(http.MethodPut, reqURL.String(), bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("unable to create HTTP request: %v", err)
	}

	// Add authorization and content-type headers
	req.Header.Add("Authorization", "Bearer "+c.Token)
	req.Header.Add("Content-Type", "application/json")

	// Send HTTP request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d. Body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var updatedMetadataKey MetadataKey
	err = json.NewDecoder(resp.Body).Decode(&updatedMetadataKey)
	if err != nil {
		return nil, fmt.Errorf("unable to decode response: %v", err)
	}

	return &updatedMetadataKey, nil
}

// DeleteMetadataKey deletes a metadata key by its id.
func (c *Client) DeleteMetadataKey(input DeleteMetadataKeyInput) error {
	// Create the url
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return fmt.Errorf("could not parse base URL: %v", err)
	}
	u.Path = path.Join(u.Path, fmt.Sprintf("metadata-keys/%s", metadataKeyID))

	// Create the request
	req, err := http.NewRequest(http.MethodDelete, u.String(), nil)
	if err != nil {
		return fmt.Errorf("could not create request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("X-Api-Key", c.ApiKey)

	// Perform the request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected response status: %d, body: %s", resp.StatusCode, body)
	}

	return nil
}

// ListAllMetadataKeys retrieves all metadata keys.
func (c *Client) ListAllMetadataKeys() ([]MetadataKey, error) {
	// Build the URL for fetching metadata keys.
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "metadata-keys")

	// Create a new request.
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Add the necessary headers.
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Add("X-Api-Key", c.ApiKey)

	// Send the request.
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Process the response.
	if resp.StatusCode != http.StatusOK {
		var errRes struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errRes); err != nil {
			return nil, fmt.Errorf("could not parse error response: %w", err)
		}
		return nil, errors.New(errRes.Error)
	}

	var keys []MetadataKey
	if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
		return nil, fmt.Errorf("could not parse metadata keys response: %w", err)
	}

	return keys, nil
}

// MetadataKeyExists sends a GET request to the service to determine if a metadata key exists.
// It returns true if it exists, false if not, and any error that occurred.
func (c *Client) MetadataKeyExists(keyName string) (bool, error) {
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return false, err
	}

	// Specify the path for the key existence check
	u.Path = path.Join(u.Path, "api/metadatakeys", keyName)

	// Create a new request
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return false, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("X-API-KEY", c.ApiKey)

	// Send the request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	} else if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	// Parse the response
	var result struct {
		Exists bool `json:"exists"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return false, err
	}

	return result.Exists, nil
}
