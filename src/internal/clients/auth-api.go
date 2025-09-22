package clients

import (
	//"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"log"

	"github.com/transistxr/coach-assignment-server/src/internal/structs"
)

// AuthClient validates API keys and verifies contact identities
// by calling the external Auth service. Used for security checks
// before booking appointments.
type AuthClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewAuthClient(baseURL string) *AuthClient {
	return &AuthClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}


func (c *AuthClient) ValidateKey(ctx context.Context, apiKey string) (*structs.ValidateResponse, error) {
	url := fmt.Sprintf("%s/validate", c.baseURL)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}

	log.Println("Request to Auth Service: ", req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	log.Println("Response from Auth Service: ", resp, resp.Body)


	var out structs.ValidateResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	log.Println("ValidateResponse: ", out.Valid)
	return &out, nil
}

func (c *AuthClient) GetKeyInfo(ctx context.Context, keyID string) (*structs.KeyInfo, error) {
	url := fmt.Sprintf("%s/keys/%s/info", c.baseURL, keyID)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("key %s not found", keyID)
	}

	var out structs.KeyInfo
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *AuthClient) RotateKey(ctx context.Context, keyID string) (*structs.RotateKeyResponse, error) {
	url := fmt.Sprintf("%s/keys/%s/rotate", c.baseURL, keyID)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("key %s not found", keyID)
	}

	var out structs.RotateKeyResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
