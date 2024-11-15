package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	data "github.com/mugglemath/go-dew/internal/models"
)

type MessagePrepFunc func(data interface{}) string

func SendDiscordMessage(prepFunc MessagePrepFunc, webhookURL string) {
	message := prepFunc(data.SensorData{})
	data := map[string]string{"content": message}
	jsonData, _ := json.Marshal(data)

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Failed to send message to Discord: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("Failed to send message to Discord: %d, %s\n", resp.StatusCode, string(body))
	}
}

func PrepareSensorFeedMessage(data data.SensorData) string {
	isoTimestamp := time.Now().Format(time.RFC3339)
	return fmt.Sprintf("%s\n"+
		"Sent from: %s\n"+
		"Indoor Temperature: %.2f C\n"+
		"Indoor Humidity: %.2f %%\n"+
		"Indoor Dewpoint: %.2f C\n"+
		"Outdoor Dewpoint: %.2f C\n"+
		"Dewpoint Delta: %.2f C\n"+
		"Keep Windows: %s\n"+
		"Humidity Alert: %t",
		isoTimestamp, data.DeviceID, data.IndoorTemperature, data.IndoorHumidity,
		data.IndoorDewpoint, data.OutdoorDewpoint, data.DewpointDelta,
		data.KeepWindows, data.HumidityAlert)
}

func PrepareWindowAlertMessage(data data.SensorData) string {
	isoTimestamp := time.Now().Format(time.RFC3339)
	return fmt.Sprintf("%s\n@everyone\n"+
		"Sent from %s\n"+
		"Indoor Dewpoint: %.2f C\n"+
		"Outdoor Dewpoint: %.2f C\n"+
		"Dewpoint Delta: %.2f C\n"+
		"Keep Windows: %s\n",
		isoTimestamp, data.DeviceID, data.IndoorDewpoint, data.OutdoorDewpoint, data.DewpointDelta, data.KeepWindows)
}

func PrepareHumidityAlertMessage(data data.SensorData) string {
	isoTimestamp := time.Now().Format(time.RFC3339)
	return fmt.Sprintf("%s\n@everyone\n"+
		"Sent from %s\n"+
		"Indoor Humidity: %.2f %%\n"+
		"Humidity Alert: %t",
		isoTimestamp, data.DeviceID, data.IndoorHumidity, data.HumidityAlert)
}
