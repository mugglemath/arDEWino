package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/mugglemath/go-dew/internal/handlers"
)

var (
	office                         string
	gridX                          string
	gridY                          string
	nwsUserAgent                   string
	discordSensorFeedWebhookURL    string
	discordWindowAlertWebhookURL   string
	discordHumidityAlertWebhookURL string
	outdoorDewpoint                float64
	isoTimestamp                   string
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	office = os.Getenv("OFFICE")
	gridX = os.Getenv("GRID_X")
	gridY = os.Getenv("GRID_Y")
	nwsUserAgent = os.Getenv("NWS_USER_AGENT")
	discordSensorFeedWebhookURL = os.Getenv("DISCORD_SENSOR_FEED_WEBHOOK_URL")
	discordWindowAlertWebhookURL = os.Getenv("DISCORD_WINDOW_ALERT_WEBHOOK_URL")
	discordHumidityAlertWebhookURL = os.Getenv("DISCORD_HUMIDITY_ALERT_WEBHOOK_URL")
}

func main() {
	r := gin.Default()

	r.GET("/weather/outdoor-dewpoint", handlers.HandleOutdoorDewpoint)
	r.POST("/arduino/sensor-feed", handlers.HandleSensorData)

	r.Run(":5000")
}
