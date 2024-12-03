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
	// set env variables
	config, err := NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	// set gin mode
	setGinMode(config.GinMode)

	// listen for SIGTERM (and SIGINT) signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	// initialize clients
	_, dbClient, err := db.ConnectToPostgres(dsn, nil)
	if err != nil {
		log.Fatalf("failed to connect to db: %s", err)
	}

	weatherClient, err := weather.NewClient(config.Office, config.GridX, config.GridY, config.NWSUserAgent)
	if err != nil {
		log.Fatalf("failed to initialize weather client: %s", err)
	}

	discordClient := discord.New(&discord.Config{
		SensorFeedWebhook:    config.DiscordSensorFeedWebhookURL,
		WindowAlertWebhook:   config.DiscordWindowAlertWebhookURL,
		HumidityAlertWebhook: config.DiscordHumidityAlertWebhookURL,
		DebugWebhook:         config.DiscordDebugWebhookURL,
	})

	handler := handler.New(dbClient, discordClient, weatherClient)
	err = handler.Initialize(ctx)
	if err != nil {
		log.Fatalf("failed to initialize app: %s", err)
	}

	// start server
	r := gin.Default()
	setPanicRecoveryMiddleware(r, discordClient.PanicHandler)
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

func setPanicRecoveryMiddleware(r *gin.Engine, fn RecoveryFn) {
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

func setGinMode(ginMode string) {
	switch ginMode {
	case "", "release":
		gin.SetMode(gin.ReleaseMode)
	case "debug":
		gin.SetMode(gin.DebugMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		log.Fatalf("Invalid GIN_MODE value: %s. Use 'debug', 'test', or 'release'.", ginMode)
	}
}
