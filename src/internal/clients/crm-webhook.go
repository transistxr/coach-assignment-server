package clients

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"io"

	"bytes"
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/transistxr/coach-assignment-server/src/internal/structs"
)

// CRMClient sends appointment lifecycle events (e.g., created, cancelled)
// to a mocked CRM. Includes retry with exponential backoff and
// idempotency key support.
type CRMClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewCRMClient(baseURL string) *CRMClient {
	return &CRMClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *CRMClient) SendAppointmentCreated(ctx context.Context, req *structs.AppointmentCreatedRequest, idempotencyKey string) (*structs.AppointmentCreatedResponse, error) {
	url := fmt.Sprintf("%s/webhooks/appointment-created", c.baseURL)
	log.Println("Sending appointment creation request to CRM Webhook")
	response := &structs.AppointmentCreatedResponse{}
	err := c.postWithIdempotency(ctx, url, req, idempotencyKey, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (c *CRMClient) SendAppointmentUpdated(ctx context.Context, req *structs.AppointmentUpdatedRequest, idempotencyKey string) (*structs.AppointmentUpdatedResponse, error) {
	url := fmt.Sprintf("%s/webhooks/appointment-updated", c.baseURL)
	log.Println("Sending appointment updation request to CRM Webhook")
	response := &structs.AppointmentUpdatedResponse{}
	err := c.postWithIdempotency(ctx, url, req, idempotencyKey, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (c *CRMClient) SendAppointmentCancelled(ctx context.Context, req *structs.AppointmentCancelledRequest, idempotencyKey string) (*structs.AppointmentCancelledResponse, error) {
	url := fmt.Sprintf("%s/webhooks/appointment-cancelled", c.baseURL)
	log.Println("Sending appointment cancellation request to CRM Webhook")
	response := &structs.AppointmentCancelledResponse{}
	err := c.postWithIdempotency(ctx, url, req, idempotencyKey, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (c *CRMClient) GetContact(ctx context.Context, contactID string) (*structs.Contact, error) {
	url := fmt.Sprintf("%s/contacts/%s", c.baseURL, contactID)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("contact %s not found", contactID)
	}

	var out structs.Contact
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *CRMClient) postWithIdempotency(ctx context.Context, url string, body interface{}, key string, out interface{}) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Idempotency-Key", key)

	maxAttempts, err := strconv.Atoi(os.Getenv("WEBHOOK_RETRY_ATTEMPTS"))
	if err != nil {
		return err
	}

	webhookDelay, err := strconv.Atoi(os.Getenv("WEBHOOK_RETRY_DELAY_MS"))
	if err != nil {
		return err
	}

	baseDelay := time.Duration(webhookDelay) * time.Millisecond
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		log.Printf("Attempt No. %d", attempt)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			log.Printf("Request error: %v", err)
			lastErr = err
		} else {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			if resp.StatusCode >= 500 {
				log.Printf("Server error, retrying.. Status: %d, Body: %s", resp.StatusCode, string(body))
			} else if resp.StatusCode >= 400 {
				return fmt.Errorf("client error: %d, Body: %s", resp.StatusCode, string(body))
			} else {
				if err := json.Unmarshal(body, out); err != nil {
					return fmt.Errorf("failed to decode response: %w", err)
				}
				log.Printf("Received success!")
				return nil
			}
		}

		sleep := baseDelay * (1 << attempt)
		jitter := time.Duration(rand.Int63n(int64(sleep / 2)))
		log.Printf("Sleeping for %v", sleep+jitter)
		time.Sleep(sleep + jitter)
	}

	return lastErr
}
