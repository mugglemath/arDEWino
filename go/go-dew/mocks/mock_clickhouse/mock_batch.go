//go:build ignore

package mockclickhouse

import (
	"github.com/ClickHouse/clickhouse-go/v2/lib/column"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type MockBatch struct {
	appendFunc func(args ...any) error
	sendFunc   func() error
	columnFunc func(int) driver.BatchColumn
}

func (m *MockBatch) SetAppend(fn func(args ...any) error) {
	m.appendFunc = fn
}

func (m *MockBatch) SetSend(fn func() error) {
	m.sendFunc = fn
}

func (m *MockBatch) Abort() error {
	return nil
}

func (m *MockBatch) Append(args ...any) error {
	if m.appendFunc != nil {
		return m.appendFunc(args...)
	}
	return nil
}

func (m *MockBatch) AppendStruct(v any) error {
	return nil
}

func (m *MockBatch) Column(i int) driver.BatchColumn {
	if m.columnFunc != nil {
		columnBatch := m.columnFunc(i)
		return columnBatch
	}
	return nil
}

func (m *MockBatch) Flush() error {
	return nil
}

func (m *MockBatch) IsSent() bool {
	return false
}

func (m *MockBatch) Rows() int {
	return 1
}

func (m *MockBatch) Send() error {
	if m.sendFunc != nil {
		return m.sendFunc()
	}
	return nil
}

func (m *MockBatch) Columns() []column.Interface {
	return nil
}
