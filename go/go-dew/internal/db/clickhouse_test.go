package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/mugglemath/go-dew/internal/model"
	mockclickhouse "github.com/mugglemath/go-dew/mocks/mock_clickhouse"
	"github.com/stretchr/testify/assert"
)

func TestCheckRecentHumidityAlert_Error(t *testing.T) {
	// setup mock
	row := mockclickhouse.Row{}
	row.SetScan(func(dest ...any) error {
		alertExistsPtr := (dest[0]).(*bool)
		*alertExistsPtr = true
		return fmt.Errorf("error in Scan")
	})
	conn := mockclickhouse.Conn{}
	conn.SetQueryRow(func(ctx context.Context, query string, args ...any) driver.Row {
		return &row
	})

	// test implementation
	client := New(&conn)
	res, err := client.CheckRecentHumidityAlert(context.TODO())

	assert.Equal(t, false, res)
	assert.Equal(t, "failed to check recent humidity alert: error in Scan", err.Error())
}

func TestCheckRecentHumidityAlert_False(t *testing.T) {
	// setup mock
	row := mockclickhouse.Row{}
	row.SetScan(func(dest ...any) error {
		alertExistsPtr := (dest[0]).(*bool)
		*alertExistsPtr = false
		return nil
	})
	conn := mockclickhouse.Conn{}
	conn.SetQueryRow(func(ctx context.Context, query string, args ...any) driver.Row {
		return &row
	})

	// test implementation
	client := New(&conn)
	res, err := client.CheckRecentHumidityAlert(context.TODO())

	assert.Equal(t, false, res)
	assert.Equal(t, nil, err)
}

func TestCheckRecentHumidityAlert_True(t *testing.T) {
	// setup mock
	row := mockclickhouse.Row{}
	row.SetScan(func(dest ...any) error {
		alertExistsPtr := (dest[0]).(*bool)
		*alertExistsPtr = true
		return nil
	})
	conn := mockclickhouse.Conn{}
	conn.SetQueryRow(func(ctx context.Context, query string, args ...any) driver.Row {
		return &row
	})

	// test implementation
	client := New(&conn)
	res, err := client.CheckRecentHumidityAlert(context.TODO())

	assert.Equal(t, true, res)
	assert.Equal(t, nil, err)
}

func TestGetLastOpenWindowsValue_True(t *testing.T) {
	// setup mock
	row := mockclickhouse.Row{}
	row.SetScan(func(dest ...any) error {
		lastOpenWindowsPtr := (dest[0]).(*bool)
		*lastOpenWindowsPtr = true
		return nil
	})
	conn := mockclickhouse.Conn{}
	conn.SetQueryRow(func(ctx context.Context, query string, args ...any) driver.Row {
		return &row
	})

	// test implementation
	client := New(&conn)
	res, err := client.GetLastOpenWindowsValue(context.TODO())

	assert.Equal(t, true, res)
	assert.Equal(t, nil, err)
}

func TestGetLastOpenWindowsValue_False(t *testing.T) {
	// setup mock
	row := mockclickhouse.Row{}
	row.SetScan(func(dest ...any) error {
		lastOpenWindowsPtr := (dest[0]).(*bool)
		*lastOpenWindowsPtr = false
		return nil
	})
	conn := mockclickhouse.Conn{}
	conn.SetQueryRow(func(ctx context.Context, query string, args ...any) driver.Row {
		return &row
	})

	// test implementation
	client := New(&conn)
	res, err := client.GetLastOpenWindowsValue(context.TODO())

	assert.Equal(t, false, res)
	assert.Equal(t, nil, err)
}

func TestGetLastOpenWindowsValue_Error(t *testing.T) {
	// setup mock
	row := mockclickhouse.Row{}
	row.SetScan(func(dest ...any) error {
		lastOpenWindowsPtr := (dest[0]).(*bool)
		*lastOpenWindowsPtr = false
		return fmt.Errorf("error in Scan")
	})
	conn := mockclickhouse.Conn{}
	conn.SetQueryRow(func(ctx context.Context, query string, args ...any) driver.Row {
		return &row
	})

	// test implementation
	client := New(&conn)
	res, err := client.GetLastOpenWindowsValue(context.TODO())

	assert.Equal(t, false, res)
	assert.Equal(t, "failed to retrieve last keep windows value: error in Scan", err.Error())
}

func TestInsertSensorFeedData(t *testing.T) {
	// setup mock
	mockBatch := new(mockclickhouse.MockBatch)
	mockConn := new(mockclickhouse.Conn)

	mockConn.SetPrepareBatch(func(ctx context.Context, query string) (driver.Batch, error) {
		return mockBatch, nil
	})

	mockBatch.SetAppend(func(args ...any) error {
		assert.Equal(t, uint16(12345), args[0])
		assert.Equal(t, 22.5, args[1])
		assert.Equal(t, 55.0, args[2])
		assert.Equal(t, 10.0, args[3])
		assert.Equal(t, 5.0, args[4])
		assert.Equal(t, 2.0, args[5])
		assert.Equal(t, true, args[6])
		assert.Equal(t, true, args[7])
		return nil
	})

	mockBatch.SetSend(func() error {
		return nil
	})

	// test implementation
	client := New(mockConn)
	sensorData := model.SensorData{
		DeviceID:          uint16(12345),
		IndoorTemperature: 22.5,
		IndoorHumidity:    55.0,
		IndoorDewpoint:    10.0,
		OutdoorDewpoint:   5.0,
		DewpointDelta:     2.0,
		OpenWindows:       true,
		HumidityAlert:     true,
	}
	err := client.InsertSensorFeedData(context.TODO(), sensorData)
	assert.NoError(t, err)
}

func TestInsertSensorFeedData_BatchCreationFail(t *testing.T) {
	// Setup mock
	mockConn := new(mockclickhouse.Conn)

	mockConn.SetPrepareBatch(func(ctx context.Context, query string) (driver.Batch, error) {
		return nil, fmt.Errorf("set prepare batch error")
	})

	client := New(mockConn)
	sensorData := model.SensorData{
		DeviceID:          12345,
		IndoorTemperature: 22.5,
		IndoorHumidity:    55.0,
		IndoorDewpoint:    10.0,
		OutdoorDewpoint:   5.0,
		DewpointDelta:     2.0,
		OpenWindows:       true,
		HumidityAlert:     true,
	}

	err := client.InsertSensorFeedData(context.TODO(), sensorData)
	assert.Error(t, err)
	assert.Equal(t, "failed to create batch for sensor data: failed to create batch: set prepare batch error", err.Error())
}

func TestInsertSensorFeedData_BatchAppendFail(t *testing.T) {
	// setup mock
	mockBatch := new(mockclickhouse.MockBatch)
	mockConn := new(mockclickhouse.Conn)

	mockConn.SetPrepareBatch(func(ctx context.Context, query string) (driver.Batch, error) {
		return mockBatch, nil
	})

	mockBatch.SetAppend(func(args ...any) error {
		return fmt.Errorf("set append error")
	})

	mockBatch.SetSend(func() error {
		return nil
	})

	// test implementation
	client := New(mockConn)
	sensorData := model.SensorData{
		DeviceID:          12345,
		IndoorTemperature: 22.5,
		IndoorHumidity:    55.0,
		IndoorDewpoint:    10.0,
		OutdoorDewpoint:   5.0,
		DewpointDelta:     2.0,
		OpenWindows:       true,
		HumidityAlert:     true,
	}
	err := client.InsertSensorFeedData(context.TODO(), sensorData)
	assert.Error(t, err)
	assert.Equal(t, "failed to append sensor data to batch: failed to append batch: set append error", err.Error())
}

func TestInsertSensorFeedData_BatchSendFail(t *testing.T) {
	// setup mock
	mockBatch := new(mockclickhouse.MockBatch)
	mockConn := new(mockclickhouse.Conn)

	mockConn.SetPrepareBatch(func(ctx context.Context, query string) (driver.Batch, error) {
		return mockBatch, nil
	})

	mockBatch.SetAppend(func(args ...any) error {
		assert.Equal(t, uint16(12345), args[0])
		assert.Equal(t, 22.5, args[1])
		assert.Equal(t, 55.0, args[2])
		assert.Equal(t, 10.0, args[3])
		assert.Equal(t, 5.0, args[4])
		assert.Equal(t, 2.0, args[5])
		assert.Equal(t, true, args[6])
		assert.Equal(t, true, args[7])
		return nil
	})

	mockBatch.SetSend(func() error {
		return fmt.Errorf("set send error")
	})

	// test implementation
	client := New(mockConn)
	sensorData := model.SensorData{
		DeviceID:          uint16(12345),
		IndoorTemperature: 22.5,
		IndoorHumidity:    55.0,
		IndoorDewpoint:    10.0,
		OutdoorDewpoint:   5.0,
		DewpointDelta:     2.0,
		OpenWindows:       true,
		HumidityAlert:     true,
	}
	err := client.InsertSensorFeedData(context.TODO(), sensorData)
	assert.Error(t, err)
	assert.Equal(t, "failed to send batch of sensor data: failed to send batch: set send error", err.Error())
}
