package clients

import (
	"fmt"
	"encoding/json"
	"strconv"
	"math/rand"
	"os"
	"log"

	"net/http"
	"time"
	"bytes"

	"github.com/transistxr/coach-assignment-server/src/internal/structs"
)


// AvailabilityClient is a client wrapper for the external Calendar API.
// It fetches raw availability data for a coach and lets you block and release slots
type AvailabilityClient struct {
	BaseURL string
	Client  *http.Client
}

func NewAvailabilityClient(baseURL string) *AvailabilityClient {
	return &AvailabilityClient{
		BaseURL: baseURL,
		Client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *AvailabilityClient) GetAvailability(coachID string, days int) (*structs.GetAvailabilityResponse, error) {

	log.Printf("Sending request to Calendar API for coach %s", coachID)

	url := fmt.Sprintf("%s/coaches/%s/availability?days=%d", c.BaseURL, coachID, days)

    maxAttempts, err := strconv.Atoi(os.Getenv("WEBHOOK_RETRY_ATTEMPTS"))
	if err != nil {
		return nil, err
	}

    webhookDelay, err := strconv.Atoi(os.Getenv("WEBHOOK_RETRY_DELAY_MS"))
	if err != nil {
		return nil, err
	}


	baseDelay := time.Duration(webhookDelay) * time.Millisecond

	var lastErr error
	var out structs.GetAvailabilityResponse
	for attempt := range maxAttempts {
		log.Printf("Attempt No. %d", attempt)
        resp, err := c.Client.Get(url)
        if err != nil {
            lastErr = err
        } else {
            defer resp.Body.Close()

            if resp.StatusCode >= 500 {
				log.Printf("Failed, retrying..")
            } else if resp.StatusCode >= 400 {
                return nil, err
            } else {
                if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
                    return nil, err
                }
				log.Printf("Recieved success!")

                return &out, nil
            }
        }

        sleep := baseDelay * (1 << attempt)
        jitter := time.Duration(rand.Int63n(int64(sleep / 2)))
        time.Sleep(sleep + jitter)
    }

	return nil, lastErr

}

func (c *AvailabilityClient) BlockSlot(coachID string, startTime time.Time, endTime time.Time) (*structs.BlockSlotResponse, error){
	url := fmt.Sprintf("%s/coaches/%s/block-slot", c.BaseURL, coachID)

	requestBody := structs.BlockSlotRequest{
		StartTime: startTime,
		EndTime: endTime,
	};

	requestBodyBytes, err := json.Marshal(requestBody)


    maxAttempts, err := strconv.Atoi(os.Getenv("WEBHOOK_RETRY_ATTEMPTS"))
	if err != nil {
		return nil, err
	}

    webhookDelay, err := strconv.Atoi(os.Getenv("WEBHOOK_RETRY_DELAY_MS"))
	if err != nil {
		return nil, err
	}


	baseDelay := time.Duration(webhookDelay) * time.Millisecond

	var lastErr error
	var response structs.BlockSlotResponse

	for attempt := range maxAttempts {
		log.Printf("Attempt No. %d", attempt)
		resp, err := c.Client.Post(url, "application/json", bytes.NewBuffer(requestBodyBytes))
		if err != nil {
			lastErr = err
		} else {
			defer resp.Body.Close()

			if resp.StatusCode > 500 {
				log.Printf("Failed, retrying..")
			} else if resp.StatusCode >= 400 {
				return nil, err
			} else {
                if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
                    return nil, err
                }
				log.Printf("Recieved success!")
                return &response, nil

			}
		}

        sleep := baseDelay * (1 << attempt)
        jitter := time.Duration(rand.Int63n(int64(sleep / 2)))
        time.Sleep(sleep + jitter)

	}

	return nil, lastErr
}

func (c *AvailabilityClient) ReleaseSlot(coachID string, blockID string, startTime time.Time, endTime time.Time) (*structs.ReleaseSlotResponse, error){
	url := fmt.Sprintf("%s/coaches/%s/release-slot", c.BaseURL, coachID)

	requestBody := structs.ReleaseSlotRequest{
		BlockId: blockID,
	};


	requestBodyBytes, err := json.Marshal(requestBody)


	resp, err := c.Client.Post(url, "application/json", bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("error calling API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var data structs.ReleaseSlotResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &data, nil

}

func (c *AvailabilityClient) GetCoachSettings(coachID string) (*structs.CoachSettingsResponse, error){
	url := fmt.Sprintf("%s/coaches/%s/settings", c.BaseURL, coachID)

	resp, err := c.Client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error calling API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var data structs.CoachSettingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &data, nil

}

// TODO: Webhook Calendar Update API -> Vague Body
