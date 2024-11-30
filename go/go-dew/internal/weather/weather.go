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

type GridResponse struct {
	Properties struct {
		Dewpoint struct {
			Values []struct {
				Value float64 `json:"value"`
			} `json:"values"`
		} `json:"dewpoint"`
	} `json:"properties"`
}

type PointResponse struct {
	Properties struct {
		Office string `json:"gridId"`
		GridX  int    `json:"gridX"`
		GridY  int    `json:"gridY"`
	} `json:"properties"`
}

func NewClient(office, gridX, gridY, userAgent string) *clientImpl {
	baseURL := fmt.Sprintf("https://api.weather.gov/gridpoints/%s/%s,%s", office, gridX, gridY)
	return &clientImpl{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    baseURL,
		userAgent:  userAgent,
	}
}

// GetGridData parses the NWS points response and returns the gridpoints variables
func GetGridData(latitude, longitude, userAgent string) (string, int, int, error) {
	baseURL := fmt.Sprintf("https://api.weather.gov/points/%s,%s", latitude, longitude)

	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		return "", 0, 0, err
	}

	req.Header.Add("User-Agent", userAgent)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", 0, 0, fmt.Errorf("error fetching grid data: %d", resp.StatusCode)
	}

	var pointResponse PointResponse
	err = json.NewDecoder(resp.Body).Decode(&pointResponse)
	if err != nil {
		return "", 0, 0, err
	}

	return pointResponse.Properties.Office, pointResponse.Properties.GridX, pointResponse.Properties.GridY, nil
}

// GetOutdoorDewPoint retrieves the outdoor dew point from NWS using the gridpoints variables
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

	var response GridResponse
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
// this function is used in dewpoint-go and unused in go-dew for now
func dewPointCalculator(temperature, relativeHumidity float64) (float64, error) {
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
