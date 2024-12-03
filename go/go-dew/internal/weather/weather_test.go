package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient_ValidInputs(t *testing.T) {
	client, err := NewClient("OFFICE", "100", "200", "test-agent")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	expectedURL := "https://api.weather.gov/gridpoints/OFFICE/100,200"
	if client.baseURL != expectedURL {
		t.Errorf("Expected baseURL %s, got %s", expectedURL, client.baseURL)
	}
	if client.userAgent != "test-agent" {
		t.Errorf("Expected userAgent %s, got %s", "test-agent", client.userAgent)
	}
	if client.httpClient.Timeout != 10*time.Second {
		t.Errorf("Expected Timeout to be %v, got %v", 10*time.Second, client.httpClient.Timeout)
	}
}

func TestNewClient_EmptyOffice(t *testing.T) {
	client, err := NewClient("", "100", "200", "test-agent")
	if err == nil {
		t.Fatal("Expected an error but got none")
	}
	if client != nil {
		t.Errorf("Expected client to be nil but got a valid client")
	}
}

func TestNewClient_EmptyGridX(t *testing.T) {
	client, err := NewClient("OFFICE", "", "200", "test-agent")
	if err == nil {
		t.Fatal("Expected an error but got none")
	}
	if client != nil {
		t.Errorf("Expected client to be nil but got a valid client")
	}
}

func TestNewClient_EmptyGridY(t *testing.T) {
	client, err := NewClient("OFFICE", "100", "", "test-agent")
	if err == nil {
		t.Fatal("Expected an error but got none")
	}
	if client != nil {
		t.Errorf("Expected client to be nil but got a valid client")
	}
}

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
