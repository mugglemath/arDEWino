package requests

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/mugglemath/dewdrop-go/pkg/calculations"
	"github.com/mugglemath/dewdrop-go/pkg/models"
)

// GetOutdoorDewpoint retrieves the outdoor dewpoint from a configured URL asynchronously.
func GetOutdoorDewpoint() (float32, error) {
	getURL := os.Getenv("GET_URL")
	getResponse, err := getRequestAsync(getURL)
	if err != nil {
		return 0, err
	}

	fmt.Printf("GET outdoor dewpoint response: %s\n", getResponse)

	dewpoint, err := strconv.ParseFloat(getResponse, 32)
	if err != nil {
		return 0, errors.New("invalid float format")
	}

	return float32(dewpoint), nil
}

// PrepareSensorFeedJSON prepares the JSON payload for the sensor feed.
func PrepareSensorFeedJSON(
	indoorData *models.IndoorSensorData,
	indoorDewpoint float32,
	outdoorDewpoint float32,
	dewpointDelta float32,
	openWindows bool,
	humidityAlert bool,
) (string, error) {
	sensorFeed := map[string]interface{}{
		"device_id":          indoorData.DeviceID,
		"indoor_temperature": calculations.RoundTo2DecimalPlaces(indoorData.Temperature),
		"indoor_humidity":    calculations.RoundTo2DecimalPlaces(indoorData.Humidity),
		"indoor_dewpoint":    calculations.RoundTo2DecimalPlaces(indoorDewpoint),
		"outdoor_dewpoint":   calculations.RoundTo2DecimalPlaces(outdoorDewpoint),
		"dewpoint_delta":     calculations.RoundTo2DecimalPlaces(dewpointDelta),
		"open_windows":       openWindows,
		"humidity_alert":     humidityAlert,
	}

	jsonData, err := json.Marshal(sensorFeed)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// PostSensorFeed posts the sensor feed JSON data to a configured URL asynchronously.
func PostSensorFeed(jsonString string) error {
	sensorFeedURL := os.Getenv("POST_URL_SENSOR_FEED")
	data := make(map[string]interface{})

	err := json.Unmarshal([]byte(jsonString), &data)
	if err != nil {
		return err
	}

	_, err = postRequestAsync(sensorFeedURL, data)
	return err
}

// getRequestAsync performs an asynchronous GET request.
func getRequestAsync(url string) (string, error) {
	resultChan := make(chan string)
	errChan := make(chan error)

	go func() {
		resp, err := http.Get(url)
		if err != nil {
			errChan <- err
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errChan <- fmt.Errorf("GET request failed with status: %s", resp.Status)
			return
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			errChan <- err
			return
		}

		resultChan <- string(body)
	}()

	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errChan:
		return "", err
	}
}

// postRequestAsync performs an asynchronous POST request.
func postRequestAsync(url string, data interface{}) (string, error) {
	resultChan := make(chan string)
	errChan := make(chan error)

	go func() {
		jsonData, err := json.Marshal(data)
		if err != nil {
			errChan <- err
			return
		}

		resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			errChan <- err
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errChan <- fmt.Errorf("POST request failed with status: %s", resp.Status)
			return
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			errChan <- err
			return
		}

		resultChan <- string(body)
	}()

	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errChan:
		return "", err
	}
}
