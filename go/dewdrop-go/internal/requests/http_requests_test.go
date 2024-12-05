package requests

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/mugglemath/dewdrop-go/pkg/calculations"
	"github.com/mugglemath/dewdrop-go/pkg/models"
)

func TestGetOutdoorDewpoint_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("15.5")); err != nil {
			log.Printf("Error writing response: %v", err)
		}
	}))
	defer mockServer.Close()

	os.Setenv("GET_URL", mockServer.URL)
	defer os.Unsetenv("GET_URL")

	client := New()
	dewpoint, err := client.GetOutdoorDewpoint()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if dewpoint != 15.5 {
		t.Errorf("expected dewpoint to be 15.5, got %v", dewpoint)
	}
}

func TestGetOutdoorDewpoint_Failure(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mockServer.Close()

	os.Setenv("GET_URL", mockServer.URL)
	defer os.Unsetenv("GET_URL")

	client := New()
	_, err := client.GetOutdoorDewpoint()
	if err == nil {
		t.Fatal("expected an error, got none")
	}
}

func TestPrepareSensorFeedJSON(t *testing.T) {
	client := New()
	indoorData := &models.IndoorSensorData{
		DeviceID:    12345,
		Temperature: 22.5,
		Humidity:    60.0,
		LedState:    true,
	}

	jsonString, err := client.PrepareSensorFeedJSON(indoorData, 15.5, 10.0, 5.5, true, false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonString), &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if deviceID, ok := result["device_id"].(float64); !ok || uint64(deviceID) != indoorData.DeviceID {
		t.Errorf("expected device_id to be %d, got %v", indoorData.DeviceID, deviceID)
	}

	expectedTemperature := calculations.RoundTo2DecimalPlaces(indoorData.Temperature)
	if indoorTemperature, ok := result["indoor_temperature"].(float64); !ok || indoorTemperature != float64(expectedTemperature) {
		t.Errorf("expected indoor_temperature to be %.2f, got %.2f", expectedTemperature, indoorTemperature)
	}
}

func TestPostSensorFeed_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var data map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil || data["device_id"] != float64(12345) {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	os.Setenv("POST_URL_SENSOR_FEED", mockServer.URL)
	defer os.Unsetenv("POST_URL_SENSOR_FEED")

	client := New()
	sensorFeedJSON := `{"device_id":12345,
	"indoor_temperature":22.5,
	"indoor_humidity":60,
	"indoor_dewpoint":15.5,
	"outdoor_dewpoint":10,
	"dewpoint_delta":5,
	"open_windows":true,
	"humidity_alert":false}`

	err := client.PostSensorFeed(sensorFeedJSON)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestPostSensorFeed_Failure(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer mockServer.Close()

	os.Setenv("POST_URL_SENSOR_FEED", mockServer.URL)
	defer os.Unsetenv("POST_URL_SENSOR_FEED")

	client := New()
	sensorFeedJSON := `{"device_id":12345,
	"indoor_temperature":22.5,
	"indoor_humidity":60,
	"indoor_dewpoint":15.5,
	"outdoor_dewpoint":10,
	"dewpoint_delta":5,
	"open_windows":true,
	"humidity_alert":false}`

	err := client.PostSensorFeed(sensorFeedJSON)
	if err == nil {
		t.Fatal("expected an error, got none")
	}
}
