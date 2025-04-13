package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client is responsible for sending requests to the notification service.
type Client struct {
	baseURL string
	client  *http.Client
}

// NewClient returns a new instance of the notification client.
// You can pass in custom config, e.g., timeouts, etc.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		client:  &http.Client{}, // or inject your own
	}
}

// SendMessageEvent sends a message event to the notification service.
// It will retry up to 3 times before returning an error.
func (c *Client) SendMessageEvent(payload Payload) error {
	url := fmt.Sprintf("%s/messages", c.baseURL)

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var lastErr error
	for i := 0; i < 3; i++ {
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("failed to call notification service: %w", err)
		} else {
			// Close response body properly.
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
			lastErr = fmt.Errorf("notification service returned non-OK status: %s", resp.Status)
		}
		// Wait a second before retrying.
		time.Sleep(1 * time.Second)
	}
	return lastErr
}

// CreateUserEvent calls POST /users and expects a 201 Created status in response.
// It will retry up to 3 times before returning an error.
func (c *Client) CreateUserEvent(name, mobile string) error {
	url := fmt.Sprintf("%s/users", c.baseURL)

	payload := map[string]string{
		"name":   name,
		"mobile": mobile,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal user payload: %w", err)
	}

	var lastErr error
	for i := 0; i < 3; i++ {
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
		if err != nil {
			return fmt.Errorf("failed to create user request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("failed to call user creation endpoint: %w", err)
		} else {
			resp.Body.Close()
			if resp.StatusCode == http.StatusCreated {
				return nil
			}
			lastErr = fmt.Errorf("notification service returned non-Created status: %s", resp.Status)
		}
		time.Sleep(1 * time.Second)
	}
	return lastErr
}
