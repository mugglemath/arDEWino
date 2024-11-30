package wifi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/mugglemath/dewdrop-go/pkg/models"
	"github.com/mugglemath/dewdrop-go/pkg/utils"
)

// GetIndoorSensorData retrieves the sensor data from the Arduino
func GetIndoorSensorData(endpoint string) (models.IndoorSensorData, error) {
	resp, err := http.Get(endpoint)
	if err != nil {
		return models.IndoorSensorData{}, err
	}
	defer resp.Body.Close()

	var indoorData models.IndoorSensorData
	if resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return models.IndoorSensorData{}, errors.New("failed to read response body")
		}

		responseString := string(body)
		fmt.Println("Response body:", strings.TrimSpace(responseString))

		parts := utils.SplitAndTrim(responseString, ',')
		if len(parts) < 4 {
			return models.IndoorSensorData{}, errors.New("invalid data format")
		}

		deviceID, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			return models.IndoorSensorData{}, errors.New("invalid device ID format")
		}

		temperature, err := strconv.ParseFloat(parts[1], 32)
		if err != nil {
			return models.IndoorSensorData{}, errors.New("invalid temperature format")
		}

		humidity, err := strconv.ParseFloat(parts[2], 32)
		if err != nil {
			return models.IndoorSensorData{}, errors.New("invalid humidity format")
		}

		var ledState bool
		if parts[3] == "1" {
			ledState = true
		} else if parts[3] == "0" {
			ledState = false
		} else {
			return models.IndoorSensorData{}, errors.New("invalid LED state value")
		}

		indoorData = models.IndoorSensorData{
			DeviceID:    deviceID,
			Temperature: float32(temperature),
			Humidity:    float32(humidity),
			LedState:    ledState,
		}
	} else {
		fmt.Println("HTTP request failed with status:", resp.Status)
		return indoorData, fmt.Errorf("failed to fetch data: %s", resp.Status)
	}

	return indoorData, nil
}

// ToggleWarningLight toggles the blinking yellow light on the Arduino
func ToggleWarningLight(openWindows bool) error {
	arduinoIP := os.Getenv("ARDUINO_IP")
	if arduinoIP == "" {
		return errors.New("ARDUINO_IP environment variable is not set")
	}

	commandValue := "1"
	if openWindows {
		commandValue = "0"
	}

	arduinoLedEndpoint := fmt.Sprintf("%s/led?state=%s", arduinoIP, commandValue)

	jsonData, err := json.Marshal(struct{}{})
	if err != nil {
		return fmt.Errorf("failed to create JSON payload: %w", err)
	}

	resp, err := http.Post(arduinoLedEndpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send POST request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to toggle light: received status code %d", resp.StatusCode)
	}

	return nil
}
