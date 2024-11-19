package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/mugglemath/go-dew/internal/model"
)

type clientImpl struct {
	clickhouse.Conn
}

type Client interface {
	InsertSensorFeedData(ctx context.Context, sensorData model.SensorData) error
	GetLastKeepWindowsValue(ctx context.Context) (string, error)
	CheckRecentHumidityAlert(ctx context.Context) (bool, error)
}

func ConnectToClickHouse(addr []string, username, password string) (Client, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr[0]},
		Auth: clickhouse.Auth{
			Username: username,
			Password: password,
		},
	})
	if err != nil {
		return nil, err
	}
	fmt.Println("Successfully connected to ClickHouse!")
	return &clientImpl{conn}, nil
}

func (c *clientImpl) InsertSensorFeedData(ctx context.Context, sensorData model.SensorData) error {
	deviceID := sensorData.DeviceID
	indoorTemperature := sensorData.IndoorTemperature
	indoorHumidity := sensorData.IndoorHumidity
	indoorDewpoint := sensorData.IndoorDewpoint
	outdoorDewpoint := sensorData.OutdoorDewpoint
	dewpointDelta := sensorData.DewpointDelta
	keepWindows := sensorData.KeepWindows
	humidityAlert := sensorData.HumidityAlert
	batch, err := createBatch(c)
	if err != nil {
		log.Printf("failed to create batch: %v", err)
	}

	err = appendToBatch(batch, deviceID, indoorTemperature, indoorHumidity, indoorDewpoint,
		outdoorDewpoint, dewpointDelta, keepWindows, humidityAlert)
	if err != nil {
		log.Printf("failed to append to batch: %v", err)
	}

	err = sendBatch(batch)
	if err != nil {
		log.Printf("failed to send batch: %v", err)
	}
	fmt.Println("Successfully inserted batch to Clickhouse!")
	return nil
}

func (c *clientImpl) GetLastKeepWindowsValue(ctx context.Context) (string, error) {
	query := `
        SELECT keep_windows
        FROM my_database.indoor_environment
        ORDER BY isoTimestamp DESC
        LIMIT 1
    `

	var lastKeepWindows string
	err := c.QueryRow(ctx, query).Scan(&lastKeepWindows)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve last humidity value: %w", err)
	}

	return lastKeepWindows, nil
}

func (c *clientImpl) CheckRecentHumidityAlert(ctx context.Context) (bool, error) {
	query := `
        SELECT COUNT(*) > 0
        FROM my_database.indoor_environment
        WHERE humidity_alert = 1 AND isoTimestamp >= now() - toIntervalHour(1)
    `

	var alertExists bool
	err := c.QueryRow(ctx, query).Scan(&alertExists)
	if err != nil {
		return false, fmt.Errorf("failed to check recent humidity alert: %w", err)
	}

	return alertExists, nil
}

func createBatch(conn clickhouse.Conn) (driver.Batch, error) {
	ctx := context.Background()

	batch, err := conn.PrepareBatch(ctx, `
    INSERT INTO my_database.indoor_environment (
        device_id, 
        indoor_temperature, 
        indoor_humidity, 
        indoor_dewpoint, 
        outdoor_dewpoint, 
        dewpoint_delta, 
        keep_windows, 
        humidity_alert, 
        isoTimestamp
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		log.Printf("failed to prepare batch: %v", err)
	}
	return batch, nil
}

func appendToBatch(batch driver.Batch, deviceID string, indoorTemperature float64,
	indoorHumidity float64, indoorDewpoint float64, outdoorDewpoint float64,
	dewpointDelta float64, keepWindows string, humidityAlert bool) error {

	if err := batch.Append(deviceID, indoorTemperature, indoorHumidity,
		indoorDewpoint, outdoorDewpoint, dewpointDelta,
		keepWindows, humidityAlert,
		time.Now().Format("2006-01-02 15:04:05")); err != nil {
		return fmt.Errorf("failed to append data to batch: %w", err)
	}
	return nil
}

func sendBatch(batch driver.Batch) error {
	if err := batch.Send(); err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}
	return nil
}
