package handlers

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/mugglemath/go-dew/internal/db"
	"github.com/mugglemath/go-dew/internal/discord"
	data "github.com/mugglemath/go-dew/internal/models"
	"github.com/mugglemath/go-dew/internal/weather"
)

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
	c.JSON(http.StatusOK, parsed)
}

func HandleSensorData(c *gin.Context) {
	var data data.SensorData
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// send to db
	conn, err := db.ConnectToClickHouse([]string{"localhost:9000"}, "default", "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to connect to ClickHouse"})
		return
	}
	defer conn.Close()

	if err := db.InsertSensorFeedData(conn, data.DeviceID, data.IndoorTemperature,
		data.IndoorHumidity, data.IndoorDewpoint, data.OutdoorDewpoint,
		data.DewpointDelta, data.KeepWindows, data.HumidityAlert); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to insert data into ClickHouse"})
		return
	}

	// TODO: implement discord notification logic
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "POST request received"})
}

func handleDiscordFeed(data data.SensorData) {
	discord.SendDiscordMessage(func(d interface{}) string {
		return discord.PrepareSensorFeedMessage(data)
	}, os.Getenv("DISCORD_SENSOR_FEED_WEBHOOK_URL"))
	fmt.Printf("Sent feed to Discord: %v\n", data)
}

func handleWindowAlert(data data.SensorData) {
	discord.SendDiscordMessage(func(d interface{}) string {
		return discord.PrepareWindowAlertMessage(data)
	}, os.Getenv("DISCORD_WINDOW_ALERT_WEBHOOK_URL"))
	fmt.Printf("Sent Window Alert to Discord: %v\n", data)
}

func handleHumidityAlert(data data.SensorData) {
	discord.SendDiscordMessage(func(d interface{}) string {
		return discord.PrepareHumidityAlertMessage(data)
	}, os.Getenv("DISCORD_HUMIDITY_ALERT_WEBHOOK_URL"))
	fmt.Printf("Sent Humidity Alert to Discord: %v\n", data)
}
