package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
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
	UpdateOutdoorDewPoint(ctx context.Context)
	Initialize(ctx context.Context) error
}

type handlerImpl struct {
	dbClient        db.Client
	discordClient   discord.Client
	weatherClient   weather.Client
	outdoorDewPoint atomic.Pointer[DewPoint]
}

type DewPoint struct {
	Value      float64
	LastUpdate time.Time
}

const (
	humidityAlertThreshold = 60.0
	updateInterval         = 15 * time.Minute
)

func New(dbClient db.Client, discordClient discord.Client, weatherClient weather.Client) Handler {
	return &handlerImpl{
		dbClient:      dbClient,
		discordClient: discordClient,
		weatherClient: weatherClient,
	}
}

func (h *handlerImpl) Initialize(ctx context.Context) error {
	return h.updateOutdoorDewPoint(ctx)
}

// UpdateOutdoorDewPoint asynchronously updates dewPoint if value is stale
func (h *handlerImpl) UpdateOutdoorDewPoint(ctx context.Context) {
	if time.Now().After(h.outdoorDewPoint.Load().LastUpdate.Add(updateInterval)) {
		go func() {
			_ = h.updateOutdoorDewPoint(ctx)
		}()
	}
}

// HandleOutdoorDewpoint may return a stale value up to twice the call interval
// (e.g. 2 minutes if called every 1 minute)
func (h *handlerImpl) HandleOutdoorDewpoint(ctx *gin.Context) {
	h.UpdateOutdoorDewPoint(ctx)
	dewPoint := h.outdoorDewPoint.Load().Value
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
		go func() {
			if err := h.discordClient.SendSensorFeed(data.FeedMessage()); err != nil {
				log.Println("failed to send data to Discord feed")
			}
		}()
	}

	// handle window alert with discord
	currentOpenWindows := data.OpenWindows
	lastOpenWindows, err := h.dbClient.GetLastOpenWindowsValue(ctx)
	if err != nil {
		fmt.Printf("failed to get last keep windows value: ")
		return
	}
	if currentOpenWindows != lastOpenWindows {
		go func() {
			if err := h.discordClient.SendSensorFeed(data.FeedMessage()); err != nil {
				log.Println("failed to send sensor feed to Discord")
			}
		}()
		go func() {
			if err := h.discordClient.SendWindowAlert(data.WindowAlertMessage()); err != nil {
				log.Println("failed to send window alert to Discord")
			}
		}()
	}

	// handle humidity alert with discord
	if data.IndoorHumidity > humidityAlertThreshold {
		recentHumidityAlert, err := h.dbClient.CheckRecentHumidityAlert(ctx)
		fmt.Printf("recent humidity alert: %t", recentHumidityAlert)
		if err != nil {
			log.Printf("failed to check recent humidity alert: %s", err)
			return
		}
		if !recentHumidityAlert {
			go func() {
				if err := h.discordClient.SendSensorFeed(data.FeedMessage()); err != nil {
					log.Println("failed to send sensor feed to Discord")
				}
			}()

			go func() {
				if err := h.discordClient.SendHumidityAlert(data.HumidityAlertMessage()); err != nil {
					log.Println("failed to send humidity alert to Discord")
				}
			}()
		}
	}

	if err := h.dbClient.InsertSensorFeedData(ctx, data); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to insert data into ClickHouse"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "POST request received"})
}

// updateOutdoorDewPoint atomically updates dewPoint from National Weather Service
func (h *handlerImpl) updateOutdoorDewPoint(ctx context.Context) (err error) {
	for i := 0; i < 10; i++ {
		var dewPoint DewPoint
		dewPoint.Value, err = h.weatherClient.GetOutdoorDewPoint(ctx)
		if err == nil {
			dewPoint.LastUpdate = time.Now()
			h.outdoorDewPoint.Store(&dewPoint)
			break
		}
		time.Sleep(time.Second * 5)
	}
	if h.outdoorDewPoint.Load() != nil && err == nil {
		log.Println("Updated outdoor dew point cache value:", h.outdoorDewPoint.Load().Value)
	}
	return
}
