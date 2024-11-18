package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string
}

func NewClient(office, gridX, gridY, userAgent string) *Client {
	baseURL := fmt.Sprintf("https://api.weather.gov/gridpoints/%s/%s,%s", office, gridX, gridY)
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    baseURL,
		userAgent:  userAgent,
	}
}

func NewClientFromLatLong(latitude, longitude, userAgent string) *Client {
	baseURL := fmt.Sprintf("https://api.weather.gov/points/%s,%s", latitude, longitude)
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    baseURL,
		userAgent:  userAgent,
	}
}

// Define structs to match the JSON structure
type WeatherResponse struct {
	Properties struct {
		Dewpoint struct {
			Values []struct {
				Value float64 `json:"value"`
			} `json:"values"`
		} `json:"dewpoint"`
	} `json:"properties"`
}

func (c *Client) GetOutdoorDewPoint(ctx context.Context) (float64, error) {
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

func DewpointCalculator(temperature, relativeHumidity float64) float64 {
	t := temperature
	rh := relativeHumidity
	return (243.04 * (math.Log(rh/100) + ((17.625 * t) / (243.04 + t)))) /
		(17.625 - math.Log(rh/100) - ((17.625 * t) / (243.04 + t)))
}
