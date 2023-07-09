// sky-accounts/pkg/clientlib/accountslib/tokens.go
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

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/google/uuid"
)

// Token represents the structure of a token.
type Token struct {
	Plaintext string
	Hash      []byte
	UserID    uuid.UUID
	Expiry    time.Time
	Scope     string
	Error     error
}

// GetTokenByPlaintextInput represents the input data for retrieving a token by plaintext.
type GetTokenByPlaintextInput struct {
	Plaintext string `json:"plaintext"`
}

// CreateTokenInput represents the input data for creating a new token.
type CreateTokenInput struct {
	UserID uuid.UUID `json:"user_id"`
	Scope  string    `json:"scope"`
}

// GetTokensByUserIDInput represents the input data for retrieving a token by user ID.
type GetTokensByUserIDInput struct {
	UserID uuid.UUID `json:"user_id"`
}

// GetTokensByScopeInput represents the input data for retrieving a token by scope.
type GetTokensByScopeInput struct {
	Scope string `json:"scope"`
}

// DeleteTokenInput represents the input data for deleting a token.
type DeleteTokenInput struct {
	UserID  uuid.UUID `json:"userId"`
	TokenID uuid.UUID `json:"tokenId"`
}

// DeleteTokensByUserIDInput represents the input for deleting all tokens of a user.
type DeleteTokensByUserIDInput struct {
	UserID uuid.UUID `json:"userId"`
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

// CreateToken creates a new token for a user and returns it.
func (c *Client) CreateToken(input CreateTokenInput) (*Token, error) {
	// Create the payload
	payload := Token{
		UserID: input.UserID,
		Scope:  input.Scope,
		Expiry: time.Now().Add(time.Hour * 24), // token expires after 24 hours
	}

	// Validate the payload
	err := payload.Validate()
	if err != nil {
		return nil, err
	}

	// Marshal the payload
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/tokens", c.BaseURL), bytes.NewBuffer(jsonPayload))
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
	var createdToken Token
	if err := json.NewDecoder(res.Body).Decode(&createdToken); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return &createdToken, nil
}

func (c *Client) GetTokenByPlaintext(input GetTokenByPlaintextInput) (*Token, error) {
	// Create the new HTTP request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/tokens/%s", c.BaseURL, url.PathEscape(input.Plaintext)), nil)
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
	token := &Token{}
	if err := json.NewDecoder(res.Body).Decode(token); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return token, nil
}

// GetTokensByUserID gets all tokens associated with a user ID.
func (c *Client) GetTokensByUserID(input GetTokensByUserIDInput) ([]Token, error) {
	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/users/%s/tokens", c.BaseURL, input.UserID), nil)
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

	// Unmarshal the body into tokens
	var tokens []Token
	err = json.Unmarshal(body, &tokens)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal response body: %w", err)
	}

	return tokens, nil
}

// GetTokensByScope gets all tokens associated with a scope.
func (c *Client) GetTokensByScope(input GetTokensByScopeInput) ([]Token, error) {
	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/tokens/scope/%s", c.BaseURL, url.PathEscape(input.Scope)), nil)
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
	var tokens []Token
	err = json.NewDecoder(res.Body).Decode(&tokens)
	if err != nil {
		return nil, fmt.Errorf("unable to decode response body: %w", err)
	}

	return tokens, nil
}

// DeleteToken deletes a token associated with a user.
func (c *Client) DeleteToken(input DeleteTokenInput) error {
	// Validate the UserID and TokenID
	if input.UserID == uuid.Nil {
		return errors.New("user ID must be non-nil UUID")
	}

	if input.TokenID == uuid.Nil {
		return errors.New("token ID must be non-nil UUID")
	}

	// Create the URL to the token to be deleted
	reqURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return fmt.Errorf("unable to parse base URL: %w", err)
	}

	reqURL.Path = path.Join(reqURL.Path, "api", "users", input.UserID.String(), "tokens", input.TokenID.String())

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodDelete, reqURL.String(), nil)
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
	if res.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("unexpected status code: got %v, body: %s", res.StatusCode, body)
	}

	return nil
}

// DeleteTokensByUserID sends a request to the server to delete all tokens for the given user ID.
func (c *Client) DeleteTokensByUserID(input DeleteTokensByUserIDInput) error {
	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/tokens/%s", c.BaseURL, input.UserID.String()), nil)
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

// Validate validates the Client fields.
func (c *Client) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.Token, validation.Required),
		validation.Field(&c.ApiKey, validation.Required),
		validation.Field(&c.BaseURL, validation.Required, is.URL),
	)
}

func (c *Client) DeleteExpiredTokens() error {
	// Validate the client before proceeding
	err := c.Validate()
	if err != nil {
		return fmt.Errorf("client validation failed: %w", err)
	}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/tokens/expired", c.BaseURL), nil)
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

func (c *Client) VerifyToken(token string) (*Token, error) {
	// Validate the client before proceeding
	err := c.Validate()
	if err != nil {
		return nil, fmt.Errorf("client validation failed: %w", err)
	}

	// Prepare the payload
	tokenPayload := map[string]string{
		"token": token,
	}

	jsonPayload, err := json.Marshal(tokenPayload)
	if err != nil {
		return nil, fmt.Errorf("error marshaling token: %w", err)
	}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/tokens/verify", c.BaseURL), bytes.NewBuffer(jsonPayload))
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

	// Decode the response body into a Token
	var verifiedToken Token
	err = json.NewDecoder(res.Body).Decode(&verifiedToken)
	if err != nil {
		return nil, fmt.Errorf("error decoding response body: %w", err)
	}

	return &verifiedToken, nil
}
