package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
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
		return nil, fmt.Errorf("failure loading .env file: %s", err)
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
	return &config, nil
}
