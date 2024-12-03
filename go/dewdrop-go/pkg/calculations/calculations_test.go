package calculations

import (
	"math"
	"testing"
)

func TestDewPointCalculator_ValidInput(t *testing.T) {
	temperature := 20.0
	relativeHumidity := 60.0
	expectedDewpoint := 11.99
	actualDewpoint, _ := DewPointCalculator(temperature, relativeHumidity)
	if math.Abs(actualDewpoint-expectedDewpoint) > 0.01 {
		t.Errorf("DewpointCalculator returned an unexpected dew point: %f, expected %f", actualDewpoint, expectedDewpoint)
	}
}

func TestDewPointCalculator_InvalidTemperature(t *testing.T) {
	temperature := -300.0
	relativeHumidity := 60.0

	_, err := DewPointCalculator(temperature, relativeHumidity)
	if err == nil {
		t.Errorf("DewpointCalculator did not return an error for invalid temperature")
	}
}

func TestDewPointCalculator_InvalidRelativeHumidity(t *testing.T) {
	temperature := 20.0
	relativeHumidity := -10.0

	_, err := DewPointCalculator(temperature, relativeHumidity)
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
		actualDewpoint, err := DewPointCalculator(test.temperature, test.relativeHumidity)
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

	_, err := DewPointCalculator(temperature, relativeHumidity)
	if err == nil {
		t.Errorf("DewpointCalculator did not return an error for a NaN temperature")
	}
}
