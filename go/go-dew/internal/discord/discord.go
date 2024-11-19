package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type clientImpl struct {
	config Config
}

type Client interface {
	SendSensorFeed(message string) error
	SendWindowAlert(message string) error
	SendHumidityAlert(message string) error
	PanicHandler(debugStack string, req *http.Request)
}

type Config struct {
	SensorFeedWebhook    string
	WindowAlertWebhook   string
	HumidityAlertWebhook string
	DebugWebhook         string
}

func New(config *Config) Client {
	return &clientImpl{
		config: *config,
	}
}

func (c *clientImpl) PanicHandler(debugStack string, req *http.Request) {
	var buf bytes.Buffer
	tee := io.TeeReader(req.Body, &buf)
	body, _ := io.ReadAll(tee)
	req.Body = io.NopCloser(&buf)

	fileContent := fmt.Sprintf("Method: %s\nURL: %s\nHeaders: %v\nBody: %s\nDebug Stack:\n%s",
		req.Method, req.URL.String(), req.Header, string(body), debugStack)

	logDir := "./logs"
	filePath, err := createFile(fileContent, logDir)
	if err != nil {
		log.Printf("Failed to create log file: %s", err)
		log.Print(fileContent)
		return
	}

	message := fmt.Sprintf("Panic occurred! Method: %s, URL: %s", req.Method, req.URL.String())

	err = sendMessageWithAttachment(c.config.DebugWebhook, message, filePath)
	if err != nil {
		log.Printf("Failed to send panic to debug channel: %s", err)
		log.Print(fileContent)
	}
}

func (c *clientImpl) SendSensorFeed(message string) error {
	return sendMessage(c.config.SensorFeedWebhook, message)
}

func (c *clientImpl) SendWindowAlert(message string) error {
	return sendMessage(c.config.WindowAlertWebhook, message)
}

func (c *clientImpl) SendHumidityAlert(message string) error {
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

func createFile(content, directory string) (string, error) {
	if err := os.MkdirAll(directory, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("panic-%s.txt", timestamp)
	filePath := filepath.Join(directory, filename)

	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return "", fmt.Errorf("failed to write to file: %w", err)
	}

	return filePath, nil
}

func sendMessageWithAttachment(webHookUrl, message, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, _ := writer.CreateFormFile("file", filepath.Base(filePath))
	io.Copy(part, file)
	writer.WriteField("content", message)
	writer.Close()

	req, _ := http.NewRequest("POST", webHookUrl, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to send message. Status: %d", resp.StatusCode)
	}

	return nil
}
