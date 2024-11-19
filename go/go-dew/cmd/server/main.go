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
	discord := ProvideDiscordClient(config)
	handler, err := InitializeApp(config)
	if err != nil {
		log.Fatalf("failed to initialize app: %s", err)
	}

	r := gin.Default()

	SetPanicRecoveryMiddleware(r, discord.PanicHandler)

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

func ProvideDB(config *Config) (*db.Client, error) {
	return db.ConnectToClickHouse([]string{"localhost:9000"}, "default", "")
}

func ProvideWeatherClient(config *Config) *weather.Client {
	return weather.NewClient(config.Office, config.GridX, config.GridY, config.NWSUserAgent)
}

func ProvideDiscordClient(config *Config) *discord.Client {
	return discord.NewClient(discord.Config{
		SensorFeedWebhook:    config.DiscordSensorFeedWebhookURL,
		WindowAlertWebhook:   config.DiscordWindowAlertWebhookURL,
		HumidityAlertWebhook: config.DiscordHumidityAlertWebhookURL,
		DebugWebhook:         config.DiscordDebugWebhookURL,
	})
}

func ProvideHandler(conn *db.Client, discordClient *discord.Client, weatherClient *weather.Client) *handler.Handler {
	return handler.New(conn, discordClient, weatherClient)
}
