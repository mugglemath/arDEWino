package mockclickhouse

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type Conn struct {
	queryRowCallback func(context.Context, string, ...any) driver.Row
}

func (c *Conn) SetQueryRow(cb func(ctx context.Context, query string, args ...any) driver.Row) {
	c.queryRowCallback = cb
}

func (c *Conn) Contributors() []string {
	return nil
}
func (c *Conn) ServerVersion() (*driver.ServerVersion, error) {
	return nil, nil
}
func (c *Conn) Select(ctx context.Context, dest any, query string, args ...any) error {
	return nil
}
func (c *Conn) Query(ctx context.Context, query string, args ...any) (driver.Rows, error) {
	return nil, nil
}
func (c *Conn) QueryRow(ctx context.Context, query string, args ...any) driver.Row {
	if c.queryRowCallback != nil {
		row := c.queryRowCallback(ctx, query, args)
		return row
	}
	return nil
}
func (c *Conn) PrepareBatch(ctx context.Context, query string, opts ...driver.PrepareBatchOption) (driver.Batch, error) {
	return nil, nil
}
func (c *Conn) Exec(ctx context.Context, query string, args ...any) error {
	return nil
}
func (c *Conn) AsyncInsert(ctx context.Context, query string, wait bool, args ...any) error {
	return nil
}
func (c *Conn) Ping(context.Context) error {
	return nil
}
func (c *Conn) Stats() driver.Stats {
	return driver.Stats{}
}
func (c *Conn) Close() error {
	return nil
}
