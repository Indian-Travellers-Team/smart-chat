package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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

// MessagePair represents user and bot messages.
type MessagePair struct {
	User string `json:"user"`
	Bot  string `json:"bot"`
}

// Payload is the JSON body sent to the notification service.
type Payload struct {
	ConversationID uint        `json:"conversation_id"`
	Mobile         string      `json:"mobile"`
	MessagePair    MessagePair `json:"message_pair"`
}

func (c *Client) SendMessageEvent(payload Payload) error {
	url := fmt.Sprintf("%s/messages", c.baseURL)

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call notification service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("notification service returned non-OK status: %s", resp.Status)
	}

	return nil
}
