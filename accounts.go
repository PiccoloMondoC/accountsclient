package accountsclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	nurl "net/url"
	"path"

	"github.com/google/uuid"
)

type Account struct {
	ID           uuid.UUID  `json:"id"`
	UserID       *uuid.UUID `json:"user_id"`
	AgencyID     *uuid.UUID `json:"agencyId,omitempty"`
	CelebrityID  *uuid.UUID `json:"celebrityId,omitempty"`
	BusinessID   *uuid.UUID `json:"businessId,omitempty"`
	EnterpriseID *uuid.UUID `json:"enterpriseId,omitempty"`
	GovernmentID *uuid.UUID `json:"governmentId,omitempty"`
}

type CreateAccountInput struct {
	UserID       *uuid.UUID `json:"user_id"`
	AgencyID     *uuid.UUID `json:"agencyId,omitempty"`
	CelebrityID  *uuid.UUID `json:"celebrityId,omitempty"`
	BusinessID   *uuid.UUID `json:"businessId,omitempty"`
	EnterpriseID *uuid.UUID `json:"enterpriseId,omitempty"`
	GovernmentID *uuid.UUID `json:"governmentId,omitempty"`
}

// Adding this method to CreateAccountInput struct
func (input *CreateAccountInput) GetAccountType() string {
	if input.UserID != nil {
		return "user"
	} else if input.AgencyID != nil {
		return "agency"
	} else if input.CelebrityID != nil {
		return "celebrity"
	} else if input.BusinessID != nil {
		return "business"
	} else if input.EnterpriseID != nil {
		return "enterprise"
	} else if input.GovernmentID != nil {
		return "government"
	} else {
		return ""
	}
}

// CreateAccount makes a POST request to create an account
func (c *Client) CreateAccount(input CreateAccountInput) (*Account, error) {
	url, err := nurl.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	// Use the GetAccountType method to get the account type
	accountType := input.GetAccountType()
	if accountType == "" {
		return nil, fmt.Errorf("could not determine the account type")
	}

	url.Path = path.Join(url.Path, accountType)

	requestBody, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url.String(), bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response from server: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var account Account
	err = json.Unmarshal(body, &account)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

// GetAccount retrieves an existing account.
func (ac *Client) GetAccount() {}

// GetAccountType retrieves the type of an existing account.
func (ac *Client) GetAccountType() {}

// UpdateAccount updates an existing account.
type UpdateAccountInput struct {
	UserID       *uuid.UUID `json:"user_id"`
	AgencyID     *uuid.UUID `json:"agencyId,omitempty"`
	CelebrityID  *uuid.UUID `json:"celebrityId,omitempty"`
	BusinessID   *uuid.UUID `json:"businessId,omitempty"`
	EnterpriseID *uuid.UUID `json:"enterpriseId,omitempty"`
	GovernmentID *uuid.UUID `json:"governmentId,omitempty"`
}

// Adding this method to UpdateAccountInput struct
func (input *UpdateAccountInput) GetAccountType() string {
	if input.UserID != nil {
		return "user"
	} else if input.AgencyID != nil {
		return "agency"
	} else if input.CelebrityID != nil {
		return "celebrity"
	} else if input.BusinessID != nil {
		return "business"
	} else if input.EnterpriseID != nil {
		return "enterprise"
	} else if input.GovernmentID != nil {
		return "government"
	} else {
		return ""
	}
}

// UpdateAccount makes a PUT request to update an account
func (c *Client) UpdateAccount(accountID uuid.UUID, input UpdateAccountInput) (*Account, error) {
	url, err := nurl.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	// Use the GetAccountType method to get the account type
	accountType := input.GetAccountType()
	if accountType == "" {
		return nil, fmt.Errorf("could not determine the account type")
	}

	url.Path = path.Join(url.Path, accountType, accountID.String())

	requestBody, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPut, url.String(), bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response from server: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var account Account
	err = json.Unmarshal(body, &account)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

// DeleteAccount makes a DELETE request to delete an account
func (c *Client) DeleteAccount(accountID uuid.UUID) error {
	accountTypes := []string{"user", "agency", "celebrity", "business", "enterprise", "government"}

	for _, accountType := range accountTypes {
		url, err := nurl.Parse(c.BaseURL)
		if err != nil {
			return err
		}

		url.Path = path.Join(url.Path, accountType, accountID.String())

		req, err := http.NewRequest(http.MethodDelete, url.String(), nil)
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.Token)

		resp, err := c.HttpClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return nil
		} else if resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("unexpected response from server: %s", resp.Status)
		}
	}

	return fmt.Errorf("account not found")
}

type AccountList struct {
	Accounts []Account `json:"accounts"`
}

// ListAccounts lists all accounts.
func (c *Client) ListAccounts() ([]Account, error) {
	accountTypes := []string{"user", "agency", "celebrity", "business", "enterprise", "government"}

	var accounts []Account

	for _, accountType := range accountTypes {
		url, err := nurl.Parse(c.BaseURL)
		if err != nil {
			return nil, err
		}

		url.Path = path.Join(url.Path, accountType)

		req, err := http.NewRequest(http.MethodGet, url.String(), nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.Token)

		resp, err := c.HttpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("bad response from server: %s", resp.Status)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var accountList AccountList
		err = json.Unmarshal(body, &accountList)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, accountList.Accounts...)
	}

	return accounts, nil
}

type SearchAccountInput struct {
	UserID       *uuid.UUID `json:"user_id"`
	AgencyID     *uuid.UUID `json:"agencyId,omitempty"`
	CelebrityID  *uuid.UUID `json:"celebrityId,omitempty"`
	BusinessID   *uuid.UUID `json:"businessId,omitempty"`
	EnterpriseID *uuid.UUID `json:"enterpriseId,omitempty"`
	GovernmentID *uuid.UUID `json:"governmentId,omitempty"`
}

// Add this method to your SearchAccountInput struct
func (input *SearchAccountInput) GetAccountType() string {
	if input.UserID != nil {
		return "user"
	} else if input.AgencyID != nil {
		return "agency"
	} else if input.CelebrityID != nil {
		return "celebrity"
	} else if input.BusinessID != nil {
		return "business"
	} else if input.EnterpriseID != nil {
		return "enterprise"
	} else if input.GovernmentID != nil {
		return "government"
	} else {
		return ""
	}
}

// SearchAccounts makes a GET request to search for accounts based on a query.
func (c *Client) SearchAccounts(input SearchAccountInput) ([]*Account, error) {
	url, err := nurl.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	// Use the GetAccountType method to get the account type
	accountType := input.GetAccountType()
	if accountType == "" {
		return nil, fmt.Errorf("could not determine the account type")
	}

	// Add the accountType to the URL
	url.Path = path.Join(url.Path, accountType, "search")

	// Convert the input to URL parameters
	params := nurl.Values{}
	if input.UserID != nil {
		params.Add("user_id", input.UserID.String())
	}
	if input.AgencyID != nil {
		params.Add("agency_id", input.AgencyID.String())
	}
	// Repeat this for all fields in the input
	// ...

	url.RawQuery = params.Encode()

	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response from server: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var accounts []*Account
	err = json.Unmarshal(body, &accounts)
	if err != nil {
		return nil, err
	}

	return accounts, nil
}

// VerifyAccountInput contains input parameters for VerifyAccount
type VerifyAccountInput struct {
	AccountID   uuid.UUID `json:"account_id"`
	AccountType string    `json:"account_type"`
}

// VerifyAccount makes a GET request to verify an account
func (c *Client) VerifyAccount(input VerifyAccountInput) (*Account, error) {
	url, err := nurl.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	accountType := input.AccountType
	if accountType == "" {
		return nil, fmt.Errorf("could not determine the account type")
	}

	// Modify the path to include the account type and ID
	url.Path = path.Join(url.Path, accountType, input.AccountID.String(), "verify")

	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response from server: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var account Account
	err = json.Unmarshal(body, &account)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

// GetAccountByField retrieves an account based on a field.
func (c *Client) GetAccountByField(fieldName string, fieldValue uuid.UUID) (*Account, error) {
	url, err := nurl.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	// Construct the URL with the specific field and value
	url.Path = path.Join(url.Path, "accounts", fieldName, fieldValue.String())

	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response from server: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var account Account
	err = json.Unmarshal(body, &account)
	if err != nil {
		return nil, err
	}

	return &account, nil
}
