package handlers

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mugglemath/go-dew/internal/discord"
	"github.com/mugglemath/go-dew/internal/weather"
)

var outdoorDewpoint string = ""

func HandleOutdoorDewpoint(c *gin.Context) {
	response, err := weather.NwsAPIResponse(os.Getenv("OFFICE"), os.Getenv("GRID_X"), os.Getenv("GRID_Y"), os.Getenv("NWS_USER_AGENT"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	parsed, err := weather.ParseOutdoorDewpoint(response)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	outdoorDewpoint = fmt.Sprintf("%.2f", parsed)
	c.JSON(http.StatusOK, parsed)
}

func HandleSensorData(c *gin.Context) {
	var data map[string]interface{}
	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	indoor_temperature := data["indoor_temperature"].(float64)
	indoor_humidity := data["indoor_humidity"].(float64)
	indoor_dewpoint := data["indoor_dewpoint"].(float64)
	outdoor_dewpoint := data["outdoor_dewpoint"].(float64)
	dewpoint_delta := data["dewpoint_delta"].(float64)
	keep_windows := data["keep_windows"].(string)
	humidity_alert := data["humidity_alert"].(bool)
	isoTimestamp := time.Now().Format(time.RFC3339)

	message := fmt.Sprintf("%s\n"+
		"Sent from arDEWino-rs\n"+
		"Indoor Temperature: %.2f C\n"+
		"Indoor Humidity: %.2f %%\n"+
		"Indoor Dewpoint: %.2f C\n"+
		"Outdoor Dewpoint: %.2f C\n"+
		"Dewpoint Delta: %.2f C\n"+
		"Keep Windows: %s\n"+
		"Humidity Alert: %t",
		isoTimestamp, indoor_temperature, indoor_humidity, indoor_dewpoint,
		outdoor_dewpoint, dewpoint_delta, keep_windows, humidity_alert)

	discord.SendDiscordMessage(message, os.Getenv("DISCORD_SENSOR_FEED_WEBHOOK_URL"))
	fmt.Printf("Received data from C++ app: %v\n", data)

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "POST request received"})
}

func HandleWindowAlert(c *gin.Context) {
	var data map[string]interface{}
	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	indoor_dewpoint := data["indoor_dewpoint"].(float64)
	outdoor_dewpoint := data["outdoor_dewpoint"].(float64)
	dewpoint_delta := data["dewpoint_delta"].(float64)
	keep_windows := data["keep_windows"].(string)
	isoTimestamp := time.Now().Format(time.RFC3339)

	message := fmt.Sprintf("%s\n@everyone\n"+
		"Sent from arDEWino-rs\n"+
		"Indoor Dewpoint: %.2f C\n"+
		"Outdoor Dewpoint: %.2f C\n"+
		"Dewpoint Delta: %.2f C\n"+
		"Keep Windows: %s\n",
		isoTimestamp, indoor_dewpoint, outdoor_dewpoint, dewpoint_delta, keep_windows)

	discord.SendDiscordMessage(message, os.Getenv("DISCORD_WINDOW_ALERT_WEBHOOK_URL"))
	fmt.Printf("Received data from C++ app: %v\n", data)

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "POST request received"})
}

func HandleHumidityAlert(c *gin.Context) {
	var data map[string]interface{}
	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	indoor_humidity := data["indoor_humidity"].(float64)
	humidity_alert := data["humidity_alert"].(bool)
	isoTimestamp := time.Now().Format(time.RFC3339)

	message := fmt.Sprintf("Sent from arDEWino-rs\n+%s\n@everyone\nIndoor Humidity: %.2f %%\nHumidity Alert: %t",
		isoTimestamp, indoor_humidity, humidity_alert)

	discord.SendDiscordMessage(message, os.Getenv("DISCORD_HUMIDITY_ALERT_WEBHOOK_URL"))
	fmt.Printf("Received data from C++ app: %v\n", data)

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "POST request received"})
}
