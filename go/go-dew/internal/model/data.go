package model

import (
	"fmt"
	"time"
)

type SensorData struct {
	DeviceID          uint64  `json:"device_id"`
	IndoorTemperature float64 `json:"indoor_temperature"`
	IndoorHumidity    float64 `json:"indoor_humidity"`
	IndoorDewpoint    float64 `json:"indoor_dewpoint"`
	OutdoorDewpoint   float64 `json:"outdoor_dewpoint"`
	DewpointDelta     float64 `json:"dewpoint_delta"`
	OpenWindows       bool    `json:"open_windows"`
	HumidityAlert     bool    `json:"humidity_alert"`
}

func (s *SensorData) FeedMessage() string {
	isoTimestamp := time.Now().Format(time.RFC3339)
	return fmt.Sprintf("%s\n"+
		"Sent from: %d\n"+
		"Indoor Temperature: %.2f C\n"+
		"Indoor Humidity: %.2f %%\n"+
		"Indoor Dewpoint: %.2f C\n"+
		"Outdoor Dewpoint: %.2f C\n"+
		"Dewpoint Delta: %.2f C\n"+
		"Open Windows: %t\n"+
		"Humidity Alert: %t",
		isoTimestamp, s.DeviceID, s.IndoorTemperature, s.IndoorHumidity,
		s.IndoorDewpoint, s.OutdoorDewpoint, s.DewpointDelta,
		s.OpenWindows, s.HumidityAlert)
}

func (s *SensorData) WindowAlertMessage() string {
	isoTimestamp := time.Now().Format(time.RFC3339)
	return fmt.Sprintf("%s\n@everyone\n"+
		"Sent from %d\n"+
		"Indoor Dewpoint: %.2f C\n"+
		"Outdoor Dewpoint: %.2f C\n"+
		"Dewpoint Delta: %.2f C\n"+
		"Open Windows: %t\n",
		isoTimestamp, s.DeviceID, s.IndoorDewpoint, s.OutdoorDewpoint, s.DewpointDelta, s.OpenWindows)
}

func (s *SensorData) HumidityAlertMessage() string {
	isoTimestamp := time.Now().Format(time.RFC3339)
	return fmt.Sprintf("%s\n@everyone\n"+
		"Sent from %d\n"+
		"Indoor Humidity: %.2f %%\n"+
		"Humidity Alert: %t",
		isoTimestamp, s.DeviceID, s.IndoorHumidity, s.HumidityAlert)
}
