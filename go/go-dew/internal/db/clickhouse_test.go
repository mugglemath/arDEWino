package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	mockclickhouse "github.com/mugglemath/go-dew/mocks/mock_clickhouse"
	"github.com/stretchr/testify/assert"
)

func TestCheckRecentHumidityAlertError(t *testing.T) {
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

func TestCheckRecentHumidityAlertFalse(t *testing.T) {
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

func TestCheckRecentHumidityAlertTrue(t *testing.T) {
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

func TestGetLastKeepWindowsValueIsOpen(t *testing.T) {
	// setup mock
	row := mockclickhouse.Row{}
	row.SetScan(func(dest ...any) error {
		lastKeepWindowsPtr := (dest[0]).(*string)
		*lastKeepWindowsPtr = "Open"
		return nil
	})
	conn := mockclickhouse.Conn{}
	conn.SetQueryRow(func(ctx context.Context, query string, args ...any) driver.Row {
		return &row
	})

	// test implementation
	client := New(&conn)
	res, err := client.GetLastKeepWindowsValue(context.TODO())

	assert.Equal(t, "Open", res)
	assert.Equal(t, nil, err)
}

func TestGetLastKeepWindowsValueIsClosed(t *testing.T) {
	// setup mock
	row := mockclickhouse.Row{}
	row.SetScan(func(dest ...any) error {
		lastKeepWindowsPtr := (dest[0]).(*string)
		*lastKeepWindowsPtr = "Closed"
		return nil
	})
	conn := mockclickhouse.Conn{}
	conn.SetQueryRow(func(ctx context.Context, query string, args ...any) driver.Row {
		return &row
	})

	// test implementation
	client := New(&conn)
	res, err := client.GetLastKeepWindowsValue(context.TODO())

	assert.Equal(t, "Closed", res)
	assert.Equal(t, nil, err)
}

func TestGetLastKeepWindowsValueError(t *testing.T) {
	// setup mock
	row := mockclickhouse.Row{}
	row.SetScan(func(dest ...any) error {
		lastKeepWindowsPtr := (dest[0]).(*string)
		*lastKeepWindowsPtr = ""
		return fmt.Errorf("error in Scan")
	})
	conn := mockclickhouse.Conn{}
	conn.SetQueryRow(func(ctx context.Context, query string, args ...any) driver.Row {
		return &row
	})

	// test implementation
	client := New(&conn)
	res, err := client.GetLastKeepWindowsValue(context.TODO())

	assert.Equal(t, "", res)
	assert.Equal(t, "failed to retrieve last humidity value: error in Scan", err.Error())
}
