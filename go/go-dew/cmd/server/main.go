package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/mugglemath/go-dew/internal/db"
	"github.com/mugglemath/go-dew/internal/discord"
	"github.com/mugglemath/go-dew/internal/handler"
	"github.com/mugglemath/go-dew/internal/weather"
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func main() {
	config, err := NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	// listen for SIGTERM (and SIGINT) signals
	_, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	// initialize clients
	conn, dbClient, err := db.ConnectToClickHouse([]string{"clickhouse-dev:9000"}, "default", "")
	if err != nil {
		log.Fatalf("failed to connect to db: %s", err)
	}
	defer conn.Close()

	weatherClient := weather.NewClient(config.Office, config.GridX, config.GridY, config.NWSUserAgent)

	discordClient := discord.New(&discord.Config{
		SensorFeedWebhook:    config.DiscordSensorFeedWebhookURL,
		WindowAlertWebhook:   config.DiscordWindowAlertWebhookURL,
		HumidityAlertWebhook: config.DiscordHumidityAlertWebhookURL,
		DebugWebhook:         config.DiscordDebugWebhookURL,
	})

	handler := handler.New(dbClient, discordClient, weatherClient)
	if err != nil {
		log.Fatalf("failed to initialize app: %s", err)
	}

	// start server
	r := gin.Default()
	SetPanicRecoveryMiddleware(r, discordClient.PanicHandler)
	r.GET("/weather/outdoor-dewpoint", handler.HandleOutdoorDewpoint)
	r.POST("/arduino/sensor-feed", handler.HandleSensorData)

	go func() {
		if err := r.Run(":5000"); err != nil {
			log.Fatalf("failed to run server: %v", err)
		}
	}()

	<-sigs
	log.Println("Received shutdown signal. Exiting...")
	cancel()
}

type RecoveryFn func(debugStack string, req *http.Request)

func SetPanicRecoveryMiddleware(r *gin.Engine, fn RecoveryFn) {
	r.Use(func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				stack := debug.Stack()
				fn(string(stack), c.Request)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "Internal Server Error",
				})
			}
		}()
		c.Next()
	})
}
