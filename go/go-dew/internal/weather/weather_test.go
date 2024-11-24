package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetOutdoorDewPoint(t *testing.T) {
	// create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// return a mock response
		resp := GridResponse{
			Properties: struct {
				Dewpoint struct {
					Values []struct {
						Value float64 `json:"value"`
					} `json:"values"`
				} `json:"dewpoint"`
			}{
				Dewpoint: struct {
					Values []struct {
						Value float64 `json:"value"`
					} `json:"values"`
				}{
					Values: []struct {
						Value float64 `json:"value"`
					}{
						{Value: 10.5},
					},
				},
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("Error encoding JSON response: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}))
	defer ts.Close()

	// create a client with the test server's URL
	c := &clientImpl{
		baseURL:   ts.URL,
		userAgent: "test-agent",
	}

	// call the function being tested
	dewPoint, err := c.GetOutdoorDewPoint(context.Background())
	if err != nil {
		t.Errorf("GetOutdoorDewPoint returned an error: %v", err)
	}

	// check the result
	if dewPoint != 10.5 {
		t.Errorf("GetOutdoorDewPoint returned an unexpected dew point: %f", dewPoint)
	}
}

func TestGetOutdoorDewPoint_InvalidRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Invalid request", http.StatusBadRequest)
	}))
	defer ts.Close()

	c := &clientImpl{
		baseURL:   ts.URL,
		userAgent: "test-agent",
	}

	_, err := c.GetOutdoorDewPoint(context.Background())
	if err == nil {
		t.Errorf("GetOutdoorDewPoint did not return an error for an invalid request")
	}
}

func TestGetOutdoorDewPoint_Non200Response(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not found", http.StatusNotFound)
	}))
	defer ts.Close()

	c := &clientImpl{
		baseURL:   ts.URL,
		userAgent: "test-agent",
	}

	_, err := c.GetOutdoorDewPoint(context.Background())
	if err == nil {
		t.Errorf("GetOutdoorDewPoint did not return an error for a non-200 response")
	}
}

func TestGetOutdoorDewPoint_JSONDecodeError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Invalid JSON")
	}))
	defer ts.Close()

	c := &clientImpl{
		baseURL:   ts.URL,
		userAgent: "test-agent",
	}

	_, err := c.GetOutdoorDewPoint(context.Background())
	if err == nil {
		t.Errorf("GetOutdoorDewPoint did not return an error for a JSON decoding error")
	}
}

func TestGetOutdoorDewPoint_EmptyDewPointValues(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := GridResponse{
			Properties: struct {
				Dewpoint struct {
					Values []struct {
						Value float64 `json:"value"`
					} `json:"values"`
				} `json:"dewpoint"`
			}{
				Dewpoint: struct {
					Values []struct {
						Value float64 `json:"value"`
					} `json:"values"`
				}{
					Values: []struct {
						Value float64 `json:"value"`
					}{},
				},
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("Error encoding JSON response: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}))
	defer ts.Close()

	c := &clientImpl{
		baseURL:   ts.URL,
		userAgent: "test-agent",
	}

	_, err := c.GetOutdoorDewPoint(context.Background())
	if err == nil {
		t.Errorf("GetOutdoorDewPoint did not return an error for empty dew point values")
	}
}

func TestGetOutdoorDewPoint_NilContext(t *testing.T) {
	c := &clientImpl{
		baseURL:   "https://example.com",
		userAgent: "test-agent",
	}

	_, err := c.GetOutdoorDewPoint(context.TODO())
	if err == nil {
		t.Errorf("GetOutdoorDewPoint did not return an error for a nil context")
	}
}

func TestGetOutdoorDewPoint_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	c := &clientImpl{
		baseURL:   "https://example.com",
		userAgent: "test-agent",
	}

	_, err := c.GetOutdoorDewPoint(ctx)
	if err == nil {
		t.Errorf("GetOutdoorDewPoint did not return an error for a cancelled context")
	}
}

func TestDewPointCalculator_ValidInput(t *testing.T) {
	temperature := 20.0
	relativeHumidity := 60.0
	expectedDewpoint := 11.99
	actualDewpoint, _ := dewPointCalculator(temperature, relativeHumidity)
	if math.Abs(actualDewpoint-expectedDewpoint) > 0.01 {
		t.Errorf("DewpointCalculator returned an unexpected dew point: %f, expected %f", actualDewpoint, expectedDewpoint)
	}
}

func TestDewPointCalculator_InvalidTemperature(t *testing.T) {
	temperature := -300.0
	relativeHumidity := 60.0

	_, err := dewPointCalculator(temperature, relativeHumidity)
	if err == nil {
		t.Errorf("DewpointCalculator did not return an error for invalid temperature")
	}
}

func TestDewPointCalculator_InvalidRelativeHumidity(t *testing.T) {
	temperature := 20.0
	relativeHumidity := -10.0

	_, err := dewPointCalculator(temperature, relativeHumidity)
	if err == nil {
		t.Errorf("DewpointCalculator did not return an error for invalid relative humidity")
	}
}

func TestDewPointCalculator_EdgeCases(t *testing.T) {
	tests := []struct {
		temperature      float64
		relativeHumidity float64
		expectedDewpoint float64
	}{
		{0, 0, -273.15},
		{0, 100, 0},
		{20, 0, -273.15},
		{20, 100, 20},
	}
	for _, test := range tests {
		actualDewpoint, err := dewPointCalculator(test.temperature, test.relativeHumidity)
		if err != nil {
			if test.relativeHumidity == 0 {
				continue
			}
			t.Errorf("DewpointCalculator returned an unexpected error: %v", err)
		}
		if math.Abs(actualDewpoint-test.expectedDewpoint) > 0.01 {
			t.Errorf("DewpointCalculator returned an unexpected dew point: %f, expected %f", actualDewpoint, test.expectedDewpoint)
		}
	}
}

func TestDewPointCalculator_NaNInput(t *testing.T) {
	temperature := math.NaN()
	relativeHumidity := 60.0

	_, err := dewPointCalculator(temperature, relativeHumidity)
	if err == nil {
		t.Errorf("DewpointCalculator did not return an error for a NaN temperature")
	}
}
