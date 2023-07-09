// sky-accounts/pkg/clientlib/accountslib/users.go
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
	"regexp"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/google/uuid"
)

// User represents the structure of a user.
type User struct {
	ID               uuid.UUID `json:"id"`
	Email            string    `json:"email"`
	Username         string    `json:"username,omitempty"`
	PasswordHash     []byte    `json:"-"`
	FirstName        string    `json:"first_name,omitempty"`
	LastName         string    `json:"last_name,omitempty"`
	DisplayName      string    `json:"display_name,omitempty"`
	AvatarURL        string    `json:"avatar_url,omitempty"`
	PhoneNumber      string    `json:"phone_number,omitempty"`
	IsActive         bool      `json:"is_active"`
	IsEmailVerified  bool      `json:"is_email_verified"`
	IsPhoneVerified  bool      `json:"is_phone_verified"`
	TwoFactorEnabled bool      `json:"two_factor_enabled"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at,omitempty"`
}

// UserRegistrationData represents the input data for a new user registration.
type UserRegistrationData struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUserPayload struct {
	Email            *string `json:"email,omitempty"`
	Username         *string `json:"username,omitempty"`
	FirstName        *string `json:"first_name,omitempty"`
	LastName         *string `json:"last_name,omitempty"`
	DisplayName      *string `json:"display_name,omitempty"`
	IsActive         *bool   `json:"is_active,omitempty"`
	IsEmailVerified  *bool   `json:"is_email_verified,omitempty"`
	IsPhoneVerified  *bool   `json:"is_phone_verified,omitempty"`
	TwoFactorEnabled *bool   `json:"two_factor_enabled,omitempty"`
}

// SetUserActiveStatusEvent represents the event of a user's active status change.
type SetUserActiveStatusEvent struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type VerifyEmailEvent struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
}

type EnableTwoFactorAuthenticationInput struct {
	UserID           uuid.UUID `json:"user_id"`
	TwoFactorEnabled bool      `json:"two_factor_enabled"`
}

type DisableTwoFactorAuthenticationInput struct {
	UserID uuid.UUID `json:"user_id"`
}

// UserRemoveRoleEvent represents the payload structure for the RemoveRoleFromUser event.
type UserRemoveRoleEvent struct {
	UserID uuid.UUID `json:"user_id"`
	RoleID uuid.UUID `json:"role_id"`
}

type RolesForUserResponse struct {
	Roles []Role `json:"roles"`
}

type UserRoleData struct {
	UserID uuid.UUID `json:"user_id"`
	RoleID uuid.UUID `json:"role_id"`
}

// AssignRoleData represents the input data to assign a role to a user.
type AssignRoleData struct {
	UserID uuid.UUID `json:"user_id"`
	RoleID uuid.UUID `json:"role_id"`
}

type UserUnassignRoleInput struct {
	UserID uuid.UUID `json:"user_id"`
	RoleID uuid.UUID `json:"role_id"`
}

// UserInRoleCheckData represents the input data for a user role check.
type UserInRoleCheckData struct {
	UserID uuid.UUID `json:"user_id"`
	RoleID uuid.UUID `json:"role_id"`
}

var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

// Validate validates the UserRegistrationData fields.
func (d *UserRegistrationData) Validate() error {
	return validation.ValidateStruct(d,
		validation.Field(&d.Email, validation.Required, validation.Length(3, 255), validation.Match(emailRegex).Error("Invalid email")),
		validation.Field(&d.Password, validation.Required, validation.Length(8, 255)),
	)
}

func (c *Client) RegisterUser(data *UserRegistrationData) error {
	// Validate the input data
	err := data.Validate()
	if err != nil {
		return err
	}

	// Marshal the input data
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/api/users", bytes.NewBuffer(jsonData))
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

// Validate validates the User fields.
func (u *User) Validate() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.ID, validation.Required),
		validation.Field(&u.Email, validation.Required, validation.Match(emailRegex).Error("Invalid email")),
		// validation.Field(&u.Username, validation.Required),
		// ... add validation for other fields as needed
	)
}

func (c *Client) CreateUser(data *User) error {
	// Validate the input data
	err := data.Validate()
	if err != nil {
		return err
	}

	// Marshal the input data
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/api/users", bytes.NewBuffer(jsonData))
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

// GetUserByID fetches a user using the user's ID
func (c *Client) GetUserByID(id uuid.UUID) (*User, error) {
	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/users/%s", c.BaseURL, id), nil)
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
	var user User
	err = json.NewDecoder(res.Body).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("unable to decode response body: %w", err)
	}

	return &user, nil
}

func (c *Client) GetUserByEmail(email string) (*User, error) {
	// Validate email
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	if !emailRegex.MatchString(email) {
		return nil, errors.New("invalid email")
	}

	// Create the URL for the request
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse base URL: %w", err)
	}
	u.Path = path.Join(u.Path, "api", "users", email)

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

	// Unmarshal the response into a User struct
	var user User
	err = json.NewDecoder(res.Body).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &user, nil
}

func (u *UpdateUserPayload) Validate() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.Email, validation.NilOrNotEmpty, is.Email),
		// add more field validations here as per your requirement
	)
}

func (c *Client) UpdateUser(userID uuid.UUID, payload *UpdateUserPayload) error {
	// Validate the payload
	if err := payload.Validate(); err != nil {
		return err
	}

	// Marshal the payload
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/api/users/%s", c.BaseURL, userID), bytes.NewBuffer(jsonPayload))
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

func (c *Client) DeleteUser(userID uuid.UUID) error {
	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/users/%s", c.BaseURL, userID), nil)
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

// Validate validates the SetUserActiveStatusEvent data.
func (e *SetUserActiveStatusEvent) Validate() error {
	return validation.ValidateStruct(e,
		validation.Field(&e.UserID, validation.Required, is.UUID),
		validation.Field(&e.IsActive, validation.Required),
	)
}

func (c *Client) SetUserActiveStatus(event *SetUserActiveStatusEvent) error {
	// Validate the event
	err := event.Validate()
	if err != nil {
		return err
	}

	// Create the payload
	payload := map[string]bool{
		"is_active": event.IsActive,
	}

	// Marshal the payload
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Prepare the URL
	fullURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return err
	}

	fullURL.Path = path.Join(fullURL.Path, fmt.Sprintf("api/users/%s", event.UserID))

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPatch, fullURL.String(), bytes.NewBuffer(jsonPayload))
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

func (e *VerifyEmailEvent) Validate() error {
	return validation.ValidateStruct(e,
		validation.Field(&e.UserID, validation.Required),
		validation.Field(&e.Email, validation.Required, validation.Match(emailRegex).Error("Invalid email")),
	)
}

func (c *Client) VerifyEmail(event *VerifyEmailEvent) error {
	// Validate the event
	if err := event.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Marshal the event
	jsonPayload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("unable to marshal event: %w", err)
	}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/users/verify-email", c.BaseURL), bytes.NewBuffer(jsonPayload))
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

func (c *Client) VerifyPhoneNumber(user *User) error {
	// Check if phone number exists
	if user.PhoneNumber == "" {
		return errors.New("phone number is required")
	}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/users/%s/verifyphonenumber", c.BaseURL, user.ID), nil)
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

func (c *Client) EnableTwoFactorAuthentication(data EnableTwoFactorAuthenticationInput) error {
	// Create the payload
	payload := struct {
		UserID           uuid.UUID `json:"user_id"`
		TwoFactorEnabled bool      `json:"two_factor_enabled"`
	}{
		UserID:           data.UserID,
		TwoFactorEnabled: data.TwoFactorEnabled,
	}

	isTrue := func(value interface{}) error {
		if b, ok := value.(bool); !ok || !b {
			return errors.New("value must be true")
		}
		return nil
	}

	// Validate the payload
	err := validation.ValidateStruct(&payload,
		validation.Field(&payload.UserID, validation.Required),
		validation.Field(&payload.TwoFactorEnabled, validation.By(isTrue)),
	)
	if err != nil {
		return err
	}

	// Marshal the payload
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/users/%s/enableTwoFactorAuthentication", c.BaseURL, data.UserID), bytes.NewBuffer(jsonPayload)) // use data.UserID instead of userID
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

func (c *Client) DisableTwoFactorAuthentication(data DisableTwoFactorAuthenticationInput) error {
	// Create the payload
	payload := struct {
		UserID           uuid.UUID `json:"user_id"`
		TwoFactorEnabled bool      `json:"two_factor_enabled"`
	}{
		UserID:           data.UserID,
		TwoFactorEnabled: false,
	}

	// Validate the payload
	err := validation.ValidateStruct(&payload,
		validation.Field(&payload.UserID, validation.Required),
		validation.Field(&payload.TwoFactorEnabled, validation.In(false)), // Only accepts false values
	)
	if err != nil {
		return err
	}

	// Marshal the payload
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/users/%s/disableTwoFactorAuthentication", c.BaseURL, data.UserID), bytes.NewBuffer(jsonPayload))
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

// ListAllUsers sends a GET request to the accounts server to get a list of all users.
func (c *Client) ListAllUsers() ([]User, error) {
	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/users", c.BaseURL), nil)
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

	// Unmarshal the response body into a slice of User structs
	var users []User
	if err := json.Unmarshal(body, &users); err != nil {
		return nil, fmt.Errorf("unable to unmarshal response body: %w", err)
	}

	// Return the list of users
	return users, nil
}

// AddRoleToUser adds a role to a user.
func (c *Client) AddRoleToUser(userID, roleID uuid.UUID) error {
	// Validate the input
	if userID == uuid.Nil || roleID == uuid.Nil {
		return errors.New("invalid input: user ID and role ID are required")
	}

	// Create the endpoint URL
	endpointURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}

	endpointURL.Path = path.Join(endpointURL.Path, fmt.Sprintf("api/users/%s/roles/%s", userID, roleID))

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPost, endpointURL.String(), nil)
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

// Validate checks if the UserRemoveRoleEvent structure is valid.
func (e *UserRemoveRoleEvent) Validate() error {
	return validation.ValidateStruct(e,
		validation.Field(&e.UserID, validation.Required, validation.NotIn(uuid.Nil).Error("Invalid UserID")),
		validation.Field(&e.RoleID, validation.Required, validation.NotIn(uuid.Nil).Error("Invalid RoleID")),
	)
}

func (c *Client) RemoveRoleFromUser(userID uuid.UUID, roleID uuid.UUID) error {
	// Create the payload
	payload := UserRemoveRoleEvent{
		UserID: userID,
		RoleID: roleID,
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
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/users/%s/roles/%s", c.BaseURL, userID, roleID), bytes.NewBuffer(jsonPayload))
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

func (c *Client) GetRolesForUser(userID uuid.UUID) ([]Role, error) {
	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/users/%s/roles", c.BaseURL, userID), nil)
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
	var resp RolesForUserResponse
	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		return nil, fmt.Errorf("unable to parse response body: %w", err)
	}

	return resp.Roles, nil
}

// Validate checks if the AssignRoleData structure is valid.
func (d *AssignRoleData) Validate() error {
	return validation.ValidateStruct(d,
		validation.Field(&d.UserID, validation.Required, validation.NotIn(uuid.Nil).Error("Invalid UserID")),
		validation.Field(&d.RoleID, validation.Required, validation.NotIn(uuid.Nil).Error("Invalid RoleID")),
	)
}

// AssignRoleToUser assigns a role to a user
func (c *Client) AssignRoleToUser(data *AssignRoleData) error {
	// Validate the data
	if err := data.Validate(); err != nil {
		return err
	}

	// Marshal the data
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Create a new HTTP request
	endpoint, err := url.Parse(c.BaseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}

	endpoint.Path = path.Join(endpoint.Path, "api", "users", data.UserID.String(), "roles", data.RoleID.String())
	req, err := http.NewRequest(http.MethodPost, endpoint.String(), bytes.NewBuffer(jsonData))
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

// UnassignRoleFromUser removes a role from a user.
func (c *Client) UnassignRoleFromUser(input *UserUnassignRoleInput) error {
	// Validate the parameters
	if input.UserID == uuid.Nil {
		return errors.New("invalid user ID")
	}
	if input.RoleID == uuid.Nil {
		return errors.New("invalid role ID")
	}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/users/%s/roles/%s", c.BaseURL, input.UserID, input.RoleID), nil)
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

func (c *Client) IsUserInRole(data *UserInRoleCheckData) (bool, error) {
	// Create the URL
	url, err := url.Parse(c.BaseURL)
	if err != nil {
		return false, fmt.Errorf("unable to parse base url: %w", err)
	}
	url.Path = path.Join(url.Path, fmt.Sprintf("api/users/%s/roles/%s", data.UserID, data.RoleID))

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
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

	// Parse the response
	var response struct {
		InRole bool `json:"in_role"`
	}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return false, fmt.Errorf("unable to decode response: %w", err)
	}

	return response.InRole, nil
}
