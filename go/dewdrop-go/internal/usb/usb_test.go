package usb

import (
	"errors"
	"testing"

	"github.com/mugglemath/dewdrop-go/pkg/models"
)

type MockSerialPort struct {
	writeData []byte
	readData  []byte
	err       error
}

func (m *MockSerialPort) Write(data []byte) (int, error) {
	m.writeData = data
	return len(data), m.err
}

func (m *MockSerialPort) Read(buffer []byte) (int, error) {
	if m.err != nil {
		return 0, m.err
	}
	copy(buffer, m.readData)
	return len(m.readData), nil
}

func TestGetIndoorSensorData_Success(t *testing.T) {
	mockPort := &MockSerialPort{
		readData: []byte("123,25.55,60.01,1"),
	}
	usbComm := &usbClientImpl{port: mockPort}

	data, err := usbComm.GetIndoorSensorData()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	expectedData := models.IndoorSensorData{
		DeviceID:    123,
		Temperature: 25.55,
		Humidity:    60.01,
		LedState:    true,
	}
	if data != expectedData {
		t.Errorf("expected data %v, got %v", expectedData, data)
	}
}

func TestGetIndoorSensorData_InvalidFormat(t *testing.T) {
	mockPort := &MockSerialPort{
		readData: []byte("123,25.5"),
	}
	usbComm := &usbClientImpl{port: mockPort}

	expectedError := "invalid data format"
	_, err := usbComm.GetIndoorSensorData()
	if err == nil || err.Error() != expectedError {
		t.Errorf("expected error %v, got %v", expectedError, err)
	}
}

func TestGetIndoorSensorData_InvalidLEDState(t *testing.T) {
	mockPort := &MockSerialPort{
		readData: []byte("123,25.55,60.01,X"),
	}
	usbComm := &usbClientImpl{port: mockPort}

	expectedError := "invalid LED state value"
	_, err := usbComm.GetIndoorSensorData()
	if err == nil || err.Error() != expectedError {
		t.Errorf("expected error %v, got %v", expectedError, err)
	}
}

func TestToggleWarningLight_SuccessOn(t *testing.T) {
	mockPort := &MockSerialPort{}
	usbComm := &usbClientImpl{port: mockPort}

	err := usbComm.ToggleWarningLight(false)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestToggleWarningLight_SuccessOff(t *testing.T) {
	mockPort := &MockSerialPort{}
	usbComm := &usbClientImpl{port: mockPort}

	err := usbComm.ToggleWarningLight(true)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestToggleWarningLight_HTTPError(t *testing.T) {
	mockPort := &MockSerialPort{
		err: errors.New("write error"),
	}
	usbComm := &usbClientImpl{port: mockPort}

	expectedError := "write error"
	err := usbComm.ToggleWarningLight(false)
	if err == nil || err.Error() != expectedError {
		t.Errorf("expected error %v, got %v", expectedError, err)
	}
}
