package weather

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"time"
)

type Client interface {
	GetOutdoorDewPoint(ctx context.Context) (float64, error)
}

type clientImpl struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string
}

func NewClient(office, gridX, gridY, userAgent string) *clientImpl {
	baseURL := fmt.Sprintf("https://api.weather.gov/gridpoints/%s/%s,%s", office, gridX, gridY)
	return &clientImpl{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    baseURL,
		userAgent:  userAgent,
	}
}

func NewClientFromLatLong(latitude, longitude, userAgent string) *clientImpl {
	baseURL := fmt.Sprintf("https://api.weather.gov/points/%s,%s", latitude, longitude)
	return &clientImpl{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    baseURL,
		userAgent:  userAgent,
	}
}

type WeatherResponse struct {
	Properties struct {
		Dewpoint struct {
			Values []struct {
				Value float64 `json:"value"`
			} `json:"values"`
		} `json:"dewpoint"`
	} `json:"properties"`
}

func (c *clientImpl) GetOutdoorDewPoint(ctx context.Context) (float64, error) {
	req, err := http.NewRequest("GET", c.baseURL, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Add("User-Agent", c.userAgent)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("error fetching weather data: %d", resp.StatusCode)
	}

	var response WeatherResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return 0, err
	}

	if len(response.Properties.Dewpoint.Values) == 0 {
		return 0, fmt.Errorf("no dewpoint values")
	}

	return response.Properties.Dewpoint.Values[0].Value, nil
}

// uses Magnus-Tetens formula for dew point
func DewPointCalculator(temperature, relativeHumidity float64) (float64, error) {
	if temperature < -273.15 {
		return temperature, errors.New("temperature must be greater than or equal to -273.15")
	}
	if relativeHumidity < 0 || relativeHumidity > 100 {
		return temperature, errors.New("relative humidity must be between 0 and 100")
	}
	if math.IsNaN(temperature) || math.IsNaN(relativeHumidity) {
		return temperature, errors.New("input values must not be NaN")
	}
	if math.IsInf(temperature, 0) || math.IsInf(relativeHumidity, 0) {
		return temperature, errors.New("input values must not be infinite")
	}

	t := temperature
	rh := relativeHumidity
	dewPoint := (243.04 * (math.Log(rh/100) + ((17.625 * t) / (243.04 + t)))) /
		(17.625 - math.Log(rh/100) - ((17.625 * t) / (243.04 + t)))

	if math.IsNaN(dewPoint) {
		return 0, errors.New("result is NaN")
	}
	return dewPoint, nil
}
