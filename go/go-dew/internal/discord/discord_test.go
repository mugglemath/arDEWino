package discord

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupMockServer(mockStatusCode int, mockResponseBody string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Expected POST method", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(mockStatusCode)
		if mockResponseBody != "" {
			if _, err := w.Write([]byte(mockResponseBody)); err != nil {
				log.Printf("Error writing response: %v", err)
			}
		}
	}))
}

func TestSendSensorFeed_Success(t *testing.T) {
	mockServer := setupMockServer(http.StatusNoContent, "")
	defer mockServer.Close()

	config := &Config{
		SensorFeedWebhook: mockServer.URL,
	}
	client := New(config)

	err := client.SendSensorFeed("Test sensor feed message")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestSendWindowAlert_Success(t *testing.T) {
	mockServer := setupMockServer(http.StatusNoContent, "")
	defer mockServer.Close()

	config := &Config{
		WindowAlertWebhook: mockServer.URL,
	}
	client := New(config)

	err := client.SendWindowAlert("Test window alert message")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestSendHumidityAlert_Success(t *testing.T) {
	mockServer := setupMockServer(http.StatusNoContent, "")
	defer mockServer.Close()

	config := &Config{
		HumidityAlertWebhook: mockServer.URL,
	}
	client := New(config)

	err := client.SendHumidityAlert("Test humidity alert message")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestSendSensorFeed_BadRequest(t *testing.T) {
	mockServer := setupMockServer(http.StatusBadRequest, "")
	defer mockServer.Close()

	config := &Config{
		SensorFeedWebhook: mockServer.URL,
	}
	client := New(config)

	err := client.SendSensorFeed("Test sensor feed message")
	if err == nil {
		t.Fatal("Expected an error but got none")
	}
}

func TestSendWindowAlert_InternalServerError(t *testing.T) {
	mockServer := setupMockServer(http.StatusInternalServerError, "")
	defer mockServer.Close()

	config := &Config{
		WindowAlertWebhook: mockServer.URL,
	}
	client := New(config)

	err := client.SendWindowAlert("Test window alert message")
	if err == nil {
		t.Fatal("Expected an error but got none")
	}
}

func TestSendHumidityAlert_MalformedJSON(t *testing.T) {
	mockResponseBody := `{"invalid_json"` // Malformed JSON
	mockServer := setupMockServer(http.StatusOK, mockResponseBody)
	defer mockServer.Close()

	config := &Config{
		HumidityAlertWebhook: mockServer.URL,
	}
	client := New(config)

	err := client.SendHumidityAlert("Test humidity alert message")
	if err == nil {
		t.Error("Expected an error due to malformed JSON but got none")
	}
}

func TestSendSensorFeed_EmptyWebhookURL(t *testing.T) {
	config := &Config{
		SensorFeedWebhook: "",
	}
	client := New(config)

	err := client.SendSensorFeed("Test sensor feed message")
	if err == nil {
		t.Error("Expected an error due to empty webhook URL but got none")
	}
}
