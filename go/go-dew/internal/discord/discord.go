package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type Client struct {
	config Config
}

type Config struct {
	SensorFeedWebhook    string
	WindowAlertWebhook   string
	HumidityAlertWebhook string
	DebugWebhook         string
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

func (c *Client) PanicHandler(debugStack string, req *http.Request) {
	// TODO: re-implement me
	var buf bytes.Buffer
	tee := io.TeeReader(req.Body, &buf)
	body, _ := io.ReadAll(tee)
	req.Body = io.NopCloser(&buf)

	// send top two in Discord message
	// put all in file
	muchPrint := func() {
		log.Printf("Request Method: %s", req.Method)
		log.Printf("Request URL: %s", req.URL.String())
		log.Printf("Request Headers: %v", req.Header)
		log.Printf("Request Body: %s", string(body))
		log.Printf("Debug Stack:\n%s", debugStack)
	}
	err := sendMessageWithAttachment(c.config.DebugWebhook, "", "", "")
	if err != nil {
		log.Printf("failed to send panic to debug channel: %s", err)
		muchPrint()
	}
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

func sendMessageWithAttachment(webHookUrl string, message string, filename string, content string) error {
	// TODO: implement and make private
	return nil
}
