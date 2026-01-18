package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"sma_event_log/internal/models"
)

const (
	searchPath  = "/api/v1/customermessages/search"
	componentID = "IGULD:SELF"
)

// Client handles HTTP communication with the SMA API
type Client struct {
	httpClient *http.Client
	baseURL    string
	username   string
	password   string
	token      string
}

// New creates a new Client instance
func New(url, username, password string) *Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	slog.Debug("initializing client", "url", url, "username", username)

	return &Client{
		httpClient: &http.Client{
			Transport: newLoggingTransport(transport),
		},
		baseURL:  url,
		username: username,
		password: password,
	}
}

// SearchMessages fetches messages from the API starting at the given marker
func (c *Client) SearchMessages(marker string, offset int) ([]models.Message, error) {
	return c.searchMessagesWithRetry(marker, offset, true)
}

func (c *Client) getToken() (string, error) {
	if c.token == "" {
		if err := c.fetchToken(); err != nil {
			return "", fmt.Errorf("failed to refresh token: %w", err)
		}
	}
	return c.token, nil
}

func (c *Client) searchMessagesWithRetry(marker string, offset int, retry bool) ([]models.Message, error) {
	searchURL := c.baseURL + searchPath

	reqBody := models.SearchRequest{
		ComponentID:      componentID,
		From:             nil,
		Until:            nil,
		MessageGroupTags: []int{},
		TraceLevels:      []string{},
		Marker:           marker,
		Offset:           offset,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, searchURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if token, err := c.getToken(); err == nil {
		req.Header.Set("Authorization", "Bearer "+token)
	} else {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode == http.StatusUnauthorized && retry {
		c.token = ""
		return c.searchMessagesWithRetry(marker, offset, false)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	var messages []models.Message
	if err := json.NewDecoder(resp.Body).Decode(&messages); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	slog.Debug("fetched messages", "count", len(messages), "marker", marker, "offset", offset)
	return messages, nil
}

// FetchAllMessages fetches all messages within the time range, calling the callback for each batch.
// Messages are returned newest to oldest. Stops fetching when messages are before the from time.
// The callback returns true to continue fetching, false to stop.
func (c *Client) FetchAllMessages(from, until time.Time, cb func(messages []models.Message) bool) error {
	marker := ""
	offset := 0

	for {
		messages, err := c.SearchMessages(marker, offset)
		if err != nil {
			return fmt.Errorf("failed to fetch messages: %w", err)
		}

		if len(messages) == 0 {
			break
		}

		// Filter messages within the time range
		var filtered []models.Message
		for _, msg := range messages {
			if !msg.Timestamp.Before(from) && msg.Timestamp.Before(until) {
				filtered = append(filtered, msg)
			}
		}

		if len(filtered) > 0 {
			if !cb(filtered) {
				break
			}
		}

		// Stop fetching if the last message is before the from time
		// (messages are ordered newest to oldest)
		lastMsg := messages[len(messages)-1]
		if lastMsg.Timestamp.Before(from) {
			break
		}

		offset += len(messages)
		marker = lastMsg.Marker
	}

	return nil
}
