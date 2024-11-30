package calculations

import (
	"errors"
	"math"
)

// DewPointCalculator uses Magnus-Tetens formula for dew point
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

func RoundTo2DecimalPlaces(value float32) float32 {
	return float32(math.Round(float64(value)*100.0)) / 100.0
}
