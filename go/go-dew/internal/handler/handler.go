package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mugglemath/go-dew/internal/db"
	"github.com/mugglemath/go-dew/internal/discord"
	"github.com/mugglemath/go-dew/internal/model"
	"github.com/mugglemath/go-dew/internal/weather"
)

type Handler interface {
	HandleOutdoorDewpoint(ctx *gin.Context)
	HandleSensorData(ctx *gin.Context)
	UpdateOutdoorDewPointCache(ctx context.Context)
}

type handlerImpl struct {
	dbClient      db.Client
	discordClient discord.Client
	weatherClient weather.Client
}

type CachedValue struct {
	OutdoorDewPoint float64
}

const (
	humidityAlertThreshold      = 60.0
	dewPointCacheUpdateInterval = 30 * time.Minute
)

var (
	cache      CachedValue
	cacheMutex sync.Mutex
)

func New(dbClient db.Client, discordClient discord.Client, weatherClient weather.Client) Handler {
	return &handlerImpl{
		dbClient:      dbClient,
		discordClient: discordClient,
		weatherClient: weatherClient,
	}
}

func (h *handlerImpl) UpdateOutdoorDewPointCache(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping UpdateOutdoorDewPointCache due to cancellation")
			return
		default:
			dewPoint, err := h.weatherClient.GetOutdoorDewPoint(ctx)
			if err != nil {
				log.Println("Error fetching data:", err)
				time.Sleep(5 * time.Second)
				continue
			}

			cacheMutex.Lock()
			cache.OutdoorDewPoint = dewPoint
			cacheMutex.Unlock()

			log.Println("Updated outdoor dew point cache value:", cache.OutdoorDewPoint)

			time.Sleep(30 * time.Minute)
		}
	}
}

func (h *handlerImpl) HandleOutdoorDewpoint(ctx *gin.Context) {
	cacheMutex.Lock()
	dewPoint := cache.OutdoorDewPoint
	cacheMutex.Unlock()

	ctx.JSON(http.StatusOK, dewPoint)
}

func (h *handlerImpl) HandleSensorData(ctx *gin.Context) {
	var data model.SensorData
	if err := ctx.ShouldBindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// if database is empty, initialize it
	empty, err := h.dbClient.CheckForEmptyTable(ctx, "data")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check row count"})
		return
	}

	if empty {
		if err := h.dbClient.InsertSensorFeedData(ctx, data); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initialize database with initial row"})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Database initialized with first entry", "data": data})
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
	currentOpenWindows := data.OpenWindows
	lastOpenWindows, err := h.dbClient.GetLastOpenWindowsValue(ctx)
	if err != nil {
		fmt.Printf("failed to get last keep windows value: ")
		return
	}
	if currentOpenWindows != lastOpenWindows {
		if err := h.discordClient.SendSensorFeed(data.FeedMessage()); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send sensor feed to Discord"})
			return
		}
		if err := h.discordClient.SendWindowAlert(data.WindowAlertMessage()); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to window alert to Discord"})
			return
		}
	}

	// handle humidity alert
	if data.IndoorHumidity > humidityAlertThreshold {
		recentHumidityAlert, err := h.dbClient.CheckRecentHumidityAlert(ctx)
		fmt.Printf("recent humidity alert: %t", recentHumidityAlert)
		if err != nil {
			fmt.Println("failed to check recent humidity alert: ")
			return
		}
		if !recentHumidityAlert {
			if err := h.discordClient.SendSensorFeed(data.FeedMessage()); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send sensor feed to Discord"})
				return
			}
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
