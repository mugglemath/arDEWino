package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/mugglemath/go-dew/internal/weather"
)

type Config struct {
	Latitude     string
	Longitude    string
	Office       string
	GridX        string
	GridY        string
	NWSUserAgent string

	DiscordSensorFeedWebhookURL    string
	DiscordWindowAlertWebhookURL   string
	DiscordHumidityAlertWebhookURL string
	DiscordDebugWebhookURL         string
}

func NewConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Printf("failed to load .env file: %s", err)
	}
	var config Config

	config.Latitude = os.Getenv("LATITUDE")
	config.Longitude = os.Getenv("LONGITUDE")
	config.Office = os.Getenv("OFFICE")
	config.GridX = os.Getenv("GRID_X")
	config.GridY = os.Getenv("GRID_Y")
	config.NWSUserAgent = os.Getenv("NWS_USER_AGENT")
	config.DiscordSensorFeedWebhookURL = os.Getenv("DISCORD_SENSOR_FEED_WEBHOOK_URL")
	config.DiscordWindowAlertWebhookURL = os.Getenv("DISCORD_WINDOW_ALERT_WEBHOOK_URL")
	config.DiscordHumidityAlertWebhookURL = os.Getenv("DISCORD_HUMIDITY_ALERT_WEBHOOK_URL")
	config.DiscordDebugWebhookURL = os.Getenv("DISCORD_DEBUG_WEBHOOK_URL")

	hasLatLong := config.Latitude != "" && config.Longitude != ""
	hasOfficeGrid := config.Office != "" && config.GridX != "" && config.GridY != ""

	if !hasLatLong && !hasOfficeGrid {
		return nil, fmt.Errorf("must provide either {LATITUDE, LONGITUDE} or {OFFICE, GRID_X, GRID_Y}")
	}

	if !hasOfficeGrid {
		office, gridX, gridY, err := weather.GetGridData(config.Latitude, config.Longitude, config.NWSUserAgent)
		if err != nil {
			log.Fatal("Error retrieving grid data:", err)
		}
		config.Office = office
		config.GridX = fmt.Sprintf("%d", gridX)
		config.GridY = fmt.Sprintf("%d", gridY)
	}

	return &config, nil
}
