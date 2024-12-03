package weather

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

func NewClient(office, gridX, gridY, userAgent string) (*clientImpl, error) {
	if office == "" {
		return nil, errors.New("office cannot be empty")
	}
	if gridX == "" {
		return nil, errors.New("gridX cannot be empty")
	}
	if gridY == "" {
		return nil, errors.New("gridY cannot be empty")
	}
	baseURL := fmt.Sprintf("https://api.weather.gov/gridpoints/%s/%s,%s", office, gridX, gridY)
	return &clientImpl{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    baseURL,
		userAgent:  userAgent,
	}, nil
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
