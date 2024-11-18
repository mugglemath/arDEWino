package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/mugglemath/go-dew/internal/db"
	"github.com/mugglemath/go-dew/internal/discord"
	"github.com/mugglemath/go-dew/internal/handler"
	"github.com/mugglemath/go-dew/internal/weather"
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
	log.SetFlags(log.Lshortfile | log.LstdFlags)
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
	// connect to db
	conn, err := db.ConnectToClickHouse([]string{"localhost:9000"}, "default", "")
	if err != nil {
		log.Fatalf("failed to connect to db: %s", err)
	}
	defer conn.Close()

	weatherClient := weather.NewClient(office, gridX, gridY, nwsUserAgent)

	discordClient := discord.NewClient(discord.Config{
		SensorFeedWebhook:    discordSensorFeedWebhookURL,
		WindowAlertWebhook:   discordWindowAlertWebhookURL,
		HumidityAlertWebhook: discordHumidityAlertWebhookURL,
	})

	h := handler.New(conn, discordClient, weatherClient)

	r := gin.Default()
	SetPanicRecoveryMiddleware(r, logRequestDetails)

	r.GET("/weather/outdoor-dewpoint", h.HandleOutdoorDewpoint)
	r.POST("/arduino/sensor-feed", h.HandleSensorData)

	r.Run(":5000")
}

// send to discord as well
func logRequestDetails(debugStack string, req *http.Request) {
	log.Printf("Debug Stack:\n%s", debugStack)
	var buf bytes.Buffer
	tee := io.TeeReader(req.Body, &buf)
	body, _ := io.ReadAll(tee)
	req.Body = io.NopCloser(&buf)

	log.Printf("Request Method: %s", req.Method)
	log.Printf("Request URL: %s", req.URL.String())
	log.Printf("Request Headers: %v", req.Header)
	log.Printf("Request Body: %s", string(body))
}

type RecoveryFn func(debugStack string, req *http.Request)

func SetPanicRecoveryMiddleware(r *gin.Engine, fn RecoveryFn) {
	r.Use(func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				log.Printf("Panic recovered: %v", err)

				// Get and log the debug stack
				stack := debug.Stack()
				fn(string(stack), c.Request)

				// Return a 500 Internal Server Error response
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "Internal Server Error",
				})
			}
		}()
		c.Next()
	})
}
