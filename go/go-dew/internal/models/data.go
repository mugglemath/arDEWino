package data

type SensorData struct {
	DeviceID          string  `json:"device_id"`
	IndoorTemperature float64 `json:"indoor_temperature"`
	IndoorHumidity    float64 `json:"indoor_humidity"`
	IndoorDewpoint    float64 `json:"indoor_dewpoint"`
	OutdoorDewpoint   float64 `json:"outdoor_dewpoint"`
	DewpointDelta     float64 `json:"dewpoint_delta"`
	KeepWindows       string  `json:"keep_windows"`
	HumidityAlert     bool    `json:"humidity_alert"`
}
