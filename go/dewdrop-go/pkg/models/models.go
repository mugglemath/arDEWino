package models

// IndoorSensorData is the response from the Arduino
type IndoorSensorData struct {
	DeviceID    uint64  `json:"device_id"`
	Temperature float32 `json:"temperature"`
	Humidity    float32 `json:"humidity"`
	LedState    bool    `json:"led_state"`
}
