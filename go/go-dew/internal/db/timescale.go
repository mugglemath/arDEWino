package db

import (
	"context"
	"fmt"

	"github.com/mugglemath/go-dew/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type clientImpl struct {
	db *gorm.DB
}

type Client interface {
	InsertSensorFeedData(ctx context.Context, sensorData model.SensorData) error
	GetLastOpenWindowsValue(ctx context.Context) (bool, error)
	CheckRecentHumidityAlert(ctx context.Context) (bool, error)
	CheckForEmptyTable(ctx context.Context, tableName string) (bool, error)
}

func New(db *gorm.DB) Client {
	return &clientImpl{db: db}
}

func ConnectToPostgres(dsn string) (*gorm.DB, Client, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}
	fmt.Println("Successfully connected to PostgreSQL!")
	return db, &clientImpl{db: db}, nil
}

func (c *clientImpl) InsertSensorFeedData(ctx context.Context, sensorData model.SensorData) error {
	if err := c.db.WithContext(ctx).Create(&sensorData).Error; err != nil {
		return fmt.Errorf("failed to insert sensor data: %w", err)
	}
	fmt.Println("Successfully inserted sensor data into PostgreSQL!")
	return nil
}

func (c *clientImpl) GetLastOpenWindowsValue(ctx context.Context) (bool, error) {
	var lastOpenWindows bool
	err := c.db.WithContext(ctx).Model(&model.SensorData{}).
		Select("open_windows").
		Order("time DESC").
		Limit(1).
		Scan(&lastOpenWindows).Error
	if err != nil {
		return false, fmt.Errorf("failed to retrieve last open windows value: %w", err)
	}
	return lastOpenWindows, nil
}

func (c *clientImpl) CheckRecentHumidityAlert(ctx context.Context) (bool, error) {
	var alertExists bool
	err := c.db.WithContext(ctx).Model(&model.SensorData{}).
		Select("COUNT(*) > 0").
		Where("humidity_alert = ? AND time >= NOW() - INTERVAL '1 hour'", true).
		Scan(&alertExists).Error
	if err != nil {
		return false, fmt.Errorf("failed to check recent humidity alert: %w", err)
	}
	return alertExists, nil
}

func (c *clientImpl) CheckForEmptyTable(ctx context.Context, tableName string) (bool, error) {
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s LIMIT 1)", tableName)
	err := c.db.WithContext(ctx).Raw(query).Scan(&exists).Error
	if err != nil {
		return false, fmt.Errorf("error checking table size: %w", err)
	}
	return !exists, nil
}
