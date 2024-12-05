package wifi

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/mugglemath/dewdrop-go/pkg/models"
)

func createMockServer(response string, statusCode int) *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		if _, err := w.Write([]byte(response)); err != nil {
			log.Printf("Error writing response: %v", err)
		}
	})
	return httptest.NewServer(handler)
}

func TestGetIndoorSensorData_Success(t *testing.T) {
	mockResponse := "123,25.5,60.0,1"
	server := createMockServer(mockResponse, http.StatusOK)
	defer server.Close()

	expectedData := models.IndoorSensorData{
		DeviceID:    123,
		Temperature: 25.5,
		Humidity:    60.0,
		LedState:    true,
	}

	wifiComm := &wifiClientImpl{}

	data, err := wifiComm.GetIndoorSensorData(server.URL)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if data != expectedData {
		t.Errorf("expected data %v, got %v", expectedData, data)
	}
}

func TestGetIndoorSensorData_InvalidFormat(t *testing.T) {
	mockResponse := "123,25.5"
	server := createMockServer(mockResponse, http.StatusOK)
	defer server.Close()

	expectedError := "invalid data format"

	wifiComm := &wifiClientImpl{}

	_, err := wifiComm.GetIndoorSensorData(server.URL)
	if err == nil || err.Error() != expectedError {
		t.Errorf("expected error %v, got %v", expectedError, err)
	}
}

func TestGetIndoorSensorData_HTTPError(t *testing.T) {
	server := createMockServer("", http.StatusInternalServerError)
	defer server.Close()

	expectedError := "failed to fetch data: 500 Internal Server Error"

	wifiComm := &wifiClientImpl{}

	_, err := wifiComm.GetIndoorSensorData(server.URL)
	if err == nil || err.Error() != expectedError {
		t.Errorf("expected error %v, got %v", expectedError, err)
	}
}

func TestToggleWarningLight_SuccessOn(t *testing.T) {
	os.Setenv("ARDUINO_IP", "http://localhost")
	mockResponse := "{}"
	server := createMockServer(mockResponse, http.StatusOK)
	defer server.Close()

	originalArduinoIP := os.Getenv("ARDUINO_IP")
	os.Setenv("ARDUINO_IP", server.URL)

	wifiComm := &wifiClientImpl{}

	err := wifiComm.ToggleWarningLight(false)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	os.Setenv("ARDUINO_IP", originalArduinoIP)
}

func TestToggleWarningLight_SuccessOff(t *testing.T) {
	os.Setenv("ARDUINO_IP", "http://localhost")
	mockResponse := "{}"
	server := createMockServer(mockResponse, http.StatusOK)
	defer server.Close()

	originalArduinoIP := os.Getenv("ARDUINO_IP")
	os.Setenv("ARDUINO_IP", server.URL)

	wifiComm := &wifiClientImpl{}

	err := wifiComm.ToggleWarningLight(true)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	os.Setenv("ARDUINO_IP", originalArduinoIP)
}

func TestToggleWarningLight_HTTPError(t *testing.T) {
	os.Setenv("ARDUINO_IP", "http://localhost")
	mockResponse := "{}"
	server := createMockServer(mockResponse, http.StatusInternalServerError)
	defer server.Close()

	originalArduinoIP := os.Getenv("ARDUINO_IP")
	os.Setenv("ARDUINO_IP", server.URL)

	expectedError := "failed to toggle light: received status code 500"

	wifiComm := &wifiClientImpl{}

	err := wifiComm.ToggleWarningLight(false)
	if err == nil || err.Error() != expectedError {
		t.Errorf("expected error %v, got %v", expectedError, err)
	}

	os.Setenv("ARDUINO_IP", originalArduinoIP)
}
