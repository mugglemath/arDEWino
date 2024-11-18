package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	config Config
}

type Config struct {
	SensorFeedWebhook    string
	WindowAlertWebhook   string
	HumidityAlertWebhook string
}

func NewClient(config Config) *Client {
	return &Client{
		config: config,
	}
}

func (c *Client) SendSensorFeed(message string) error {
	return sendMessage(c.config.SensorFeedWebhook, message)
}

func (c *Client) SendWindowAlert(message string) error {
	return sendMessage(c.config.WindowAlertWebhook, message)
}

func (c *Client) SendHumidityAlert(message string) error {
	return sendMessage(c.config.HumidityAlertWebhook, message)
}

func sendMessage(webHookURL string, message string) error {
	data := map[string]string{"content": message}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal message data: %w", err)
	}

	resp, err := http.Post(webHookURL, "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to send message to Discord: %w\n%s", err, body)
		}
	}
	return nil
}
