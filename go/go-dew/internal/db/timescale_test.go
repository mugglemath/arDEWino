package db

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mugglemath/go-dew/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestConnectToPostgres(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm database: %v", err)
	}

	dsn := "host=localhost user=user password=pw dbname=db port=5432 sslmode=disable"
	resultDB, client, err := ConnectToPostgres(dsn, gormDB)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if resultDB == nil {
		t.Error("expected a valid gorm.DB instance, got nil")
	}

	if client == nil {
		t.Error("expected a valid Client instance, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestInsertSensorFeedData_Success(t *testing.T) {
	// setup mock
	client, mock := setupTestDB(t)

	sensorData := model.SensorData{
		DeviceID:          1,
		IndoorTemperature: 1.0,
		IndoorHumidity:    1.0,
		IndoorDewpoint:    1.0,
		OutdoorDewpoint:   1.0,
		DewpointDelta:     1.0,
		OpenWindows:       true,
		HumidityAlert:     false}

	// setup test
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO \"data\"").
		WithArgs(
			sensorData.DeviceID,
			sensorData.IndoorTemperature,
			sensorData.IndoorHumidity,
			sensorData.IndoorDewpoint,
			sensorData.OutdoorDewpoint,
			sensorData.DewpointDelta,
			sensorData.OpenWindows,
			sensorData.HumidityAlert).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := client.InsertSensorFeedData(context.Background(), sensorData)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestInsertSensorFeedData_Error(t *testing.T) {
	// setup mock
	client, mock := setupTestDB(t)

	sensorData := model.SensorData{
		DeviceID:          1,
		IndoorTemperature: 1.0,
		IndoorHumidity:    1.0,
		IndoorDewpoint:    1.0,
		OutdoorDewpoint:   1.0,
		DewpointDelta:     1.0,
		OpenWindows:       true,
		HumidityAlert:     false}

	// setup test
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO \"data\"").WithArgs(
		sensorData.DeviceID,
		sensorData.IndoorTemperature,
		sensorData.IndoorHumidity,
		sensorData.IndoorDewpoint,
		sensorData.OutdoorDewpoint,
		sensorData.DewpointDelta,
		sensorData.OpenWindows,
		sensorData.HumidityAlert).
		WillReturnError(errors.New("insert error"))
	mock.ExpectRollback()

	err := client.InsertSensorFeedData(context.Background(), sensorData)

	if err == nil {
		t.Errorf("expected an error but got none")
	} else if err.Error() != "failed to insert sensor data: insert error" {
		t.Errorf("expected failed to insert sensor data: insert error but got %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetLastOpenWindowsValue_True(t *testing.T) {
	// setup mock
	client, mock := setupTestDB(t)

	// setup test
	expectedOpenWindows := true

	mock.ExpectQuery(`SELECT "open_windows" FROM "data" ORDER BY time DESC LIMIT \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"open_windows"}).AddRow(expectedOpenWindows))

	openWindowsValue, err := client.GetLastOpenWindowsValue(context.Background())
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}

	if openWindowsValue != expectedOpenWindows {
		t.Errorf("expected %v but got %v", expectedOpenWindows, openWindowsValue)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetLastOpenWindowsValue_False(t *testing.T) {
	// setup mock
	client, mock := setupTestDB(t)

	// setup test
	expectedOpenWindows := false

	mock.ExpectQuery(`SELECT "open_windows" FROM "data" ORDER BY time DESC LIMIT \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"open_windows"}).AddRow(expectedOpenWindows))

	openWindowsValue, err := client.GetLastOpenWindowsValue(context.Background())
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}

	if openWindowsValue != expectedOpenWindows {
		t.Errorf("expected %v but got %v", expectedOpenWindows, openWindowsValue)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetLastOpenWindowsValue_Error(t *testing.T) {
	// setup mock
	client, mock := setupTestDB(t)

	// setup test
	mock.ExpectQuery(`SELECT "open_windows" FROM "data" ORDER BY time DESC LIMIT \$1`).
		WithArgs(1).
		WillReturnError(errors.New("query error"))

	_, err := client.GetLastOpenWindowsValue(context.Background())

	if err == nil {
		t.Errorf("expected an error but got none")
	} else if err.Error() != "failed to retrieve last open windows value: query error" {
		t.Errorf("expected failed to retrieve last open windows value: query error but got %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCheckRecentHumidityAlert_True(t *testing.T) {
	// setup mock
	client, mock := setupTestDB(t)

	// setup test
	expectedAlert := true

	mock.ExpectQuery(`SELECT COUNT\(\*\) > 0 FROM "data" WHERE humidity_alert = \$1 AND time >= NOW\(\) - INTERVAL '1 hour'`).
		WithArgs(true).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedAlert))

	alertValue, err := client.CheckRecentHumidityAlert(context.Background())
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}

	if alertValue != expectedAlert {
		t.Errorf("expected %v but got %v", expectedAlert, alertValue)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCheckRecentHumidityAlert_False(t *testing.T) {
	// setup mock
	client, mock := setupTestDB(t)

	// setup test
	expectedAlert := false

	mock.ExpectQuery(`SELECT COUNT\(\*\) > 0 FROM "data" WHERE humidity_alert = \$1 AND time >= NOW\(\) - INTERVAL '1 hour'`).
		WithArgs(true).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(expectedAlert))

	alertValue, err := client.CheckRecentHumidityAlert(context.Background())
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}

	if alertValue != expectedAlert {
		t.Errorf("expected %v but got %v", expectedAlert, alertValue)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCheckRecentHumidityAlert_Error(t *testing.T) {
	// setup mock
	client, mock := setupTestDB(t)

	// setup test
	mock.ExpectQuery(`SELECT COUNT\(\*\) > 0 FROM "data" WHERE humidity_alert = \$1 AND time >= NOW\(\) - INTERVAL '1 hour'`).
		WithArgs(true).
		WillReturnError(errors.New("query error"))

	_, err := client.CheckRecentHumidityAlert(context.Background())

	if err == nil {
		t.Errorf("expected an error but got none")
	} else if err.Error() != "failed to check recent humidity alert: query error" {
		t.Errorf("expected failed to check recent humidity alert: query error but got %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCheckForEmptyTable_True(t *testing.T) {
	// setup mock
	client, mock := setupTestDB(t)

	// setup test
	expectedExists := false
	expectedCheck := true

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM data LIMIT 1\)`).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(expectedExists))

	alertValue, err := client.CheckForEmptyTable(context.Background(), "data")
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}

	if alertValue != expectedCheck {
		t.Errorf("expected %v but got %v", expectedCheck, alertValue)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCheckForEmptyTable_False(t *testing.T) {
	// setup mock
	client, mock := setupTestDB(t)

	// setup test
	expectedExists := true
	expectedCheck := false

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM data LIMIT 1\)`).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(expectedExists))

	alertValue, err := client.CheckForEmptyTable(context.Background(), "data")
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}

	if alertValue != expectedCheck {
		t.Errorf("expected %v but got %v", expectedCheck, alertValue)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCheckForEmptyTable_Error(t *testing.T) {
	// setup mock
	client, mock := setupTestDB(t)

	// setup test
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM data LIMIT 1\)`).
		WillReturnError(errors.New("query error"))

	_, err := client.CheckForEmptyTable(context.Background(), "data")
	if err == nil {
		t.Errorf("expected an error but got none")
	} else if err.Error() != "error checking table size: query error" {
		t.Errorf("expected error checking table size: query error but got %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func setupTestDB(t *testing.T) (*clientImpl, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm DB: %v", err)
	}

	client := &clientImpl{db: gormDB}
	return client, mock
}
