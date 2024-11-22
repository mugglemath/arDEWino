package mockclickhouse

type MockBatchColumn struct {
	batchColumnCallback func(any) error
}

func (m *MockBatchColumn) SetAppend(cb func(any) error) {
	m.batchColumnCallback = cb
}

func (m *MockBatchColumn) Append(any) error {
	return nil
}

func (m *MockBatchColumn) AppendRow(any) error {
	return nil
}
