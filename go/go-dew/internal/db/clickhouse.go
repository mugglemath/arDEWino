package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

func ConnectToClickHouse(addr []string, username, password string) (clickhouse.Conn, error) {
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
	return conn, nil
}

func InsertSensorFeedData(conn clickhouse.Conn, deviceID string, indoorTemperature float64,
	indoorHumidity float64, indoorDewpoint float64, outdoorDewpoint float64,
	dewpointDelta float64, keepWindows string, humidityAlert bool) error {

	batch, err := createBatch(conn)
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
