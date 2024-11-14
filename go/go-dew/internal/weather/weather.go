package weather

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
)

func NwsAPIResponse(office string, gridX string, gridY string, nwsUserAgent string) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.weather.gov/gridpoints/%s/%s,%s", office, gridX, gridY)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(nwsUserAgent, "")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error fetching weather data: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func ParseOutdoorDewpoint(response map[string]interface{}) (float64, error) {
	properties, ok := response["properties"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("invalid response format")
	}

	dewpoint, ok := properties["dewpoint"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("invalid dewpoint data")
	}

	values, ok := dewpoint["values"].([]interface{})
	if !ok || len(values) == 0 {
		return 0, fmt.Errorf("no dewpoint values")
	}

	firstValue, ok := values[0].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("invalid first dewpoint value")
	}

	value, ok := firstValue["value"].(float64)
	if !ok {
		return 0, fmt.Errorf("invalid dewpoint value type")
	}

	return value, nil
}

func DewpointCalculator(T, RH float64) float64 {
	return (243.04 * (math.Log(RH/100) + ((17.625 * T) / (243.04 + T)))) /
		(17.625 - math.Log(RH/100) - ((17.625 * T) / (243.04 + T)))
}
