package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mugglemath/go-dew/internal/db"
	"github.com/mugglemath/go-dew/internal/discord"
	"github.com/mugglemath/go-dew/internal/model"
	"github.com/mugglemath/go-dew/internal/weather"
)

type Handler struct {
	dbClient      *db.Client
	discordClient *discord.Client
	weatherClient *weather.Client
}

func New(dbClient *db.Client, discordClient *discord.Client, weatherClient *weather.Client) *Handler {
	return &Handler{
		dbClient:      dbClient,
		discordClient: discordClient,
		weatherClient: weatherClient,
	}
}

func (h *Handler) HandleOutdoorDewpoint(ctx *gin.Context) {
	dewPoint, err := h.weatherClient.GetOutdoorDewPoint(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, dewPoint)
}

func (h *Handler) HandleSensorData(ctx *gin.Context) {
	var data model.SensorData
	if err := ctx.ShouldBindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// send Discord feed if it's time
	now := time.Now()
	if now.Minute() == 0 {
		if err := h.discordClient.SendSensorFeed(data.FeedMessage()); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send data to Discord feed"})
			return
		}
	}

	// handle window alert toggle
	currentKeepWindows := data.KeepWindows
	lastKeepWindows, err := h.dbClient.GetLastKeepWindowsValue(ctx)
	if err != nil {
		fmt.Printf("failed to get last keep windows value: ")
	}
	if currentKeepWindows != lastKeepWindows {
		h.discordClient.SendSensorFeed(data.FeedMessage())
		if err := h.discordClient.SendWindowAlert(data.WindowAlertMessage()); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to window alert to Discord"})
			return
		}
	}

	// handle humidity alert
	fmt.Printf("indoor_humidity: %f", data.IndoorHumidity)
	if data.IndoorHumidity > 60.0 {
		recentHumidityAlert, err := h.dbClient.CheckRecentHumidityAlert(ctx)
		fmt.Printf("recent humidity alert: %t", recentHumidityAlert)
		if err != nil {
			fmt.Println("failed to check recent humidity alert: ")
		}
		if !recentHumidityAlert {
			h.discordClient.SendSensorFeed(data.FeedMessage())
			if err := h.discordClient.SendHumidityAlert(data.HumidityAlertMessage()); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send humidity alert to Discord"})
				return
			}
		}
	}

	if err := h.dbClient.InsertSensorFeedData(ctx, data); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to insert data into ClickHouse"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "POST request received"})
}
