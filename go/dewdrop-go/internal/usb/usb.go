package usb

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/mugglemath/dewdrop-go/pkg/models"
	"github.com/mugglemath/dewdrop-go/pkg/utils"
	"github.com/tarm/serial"
)

type Client interface {
	GetIndoorSensorData() (models.IndoorSensorData, error)
	ToggleWarningLight(openWindows bool) error
}

type SerialPort interface {
	Write(data []byte) (int, error)
	Read(buffer []byte) (int, error)
}

type usbClientImpl struct {
	port SerialPort
}

func NewUsbCommunication(portName string) (*usbClientImpl, error) {
	c := &serial.Config{
		Name: portName,
		Baud: 115200,
	}
	port, err := serial.OpenPort(c)
	if err != nil {
		return nil, err
	}
	return &usbClientImpl{port: port}, nil
}

// GetIndoorSensorData retrieves the sensor data from the Arduino
func (usb *usbClientImpl) GetIndoorSensorData() (models.IndoorSensorData, error) {
	command := "d"
	response, err := usb.getArduinoResponse(command, 50*time.Millisecond)
	if err != nil {
		return models.IndoorSensorData{}, err
	}

	parts := utils.SplitAndTrim(response, ',')
	if len(parts) < 4 {
		return models.IndoorSensorData{}, errors.New("invalid data format")
	}

	deviceID, _ := strconv.ParseUint(parts[0], 10, 64)
	temperature, _ := strconv.ParseFloat(parts[1], 32)
	humidity, _ := strconv.ParseFloat(parts[2], 32)
	var ledState bool
	if parts[3] == "1" {
		ledState = true
	} else if parts[3] == "0" {
		ledState = false
	} else {
		return models.IndoorSensorData{}, errors.New("invalid LED state value")
	}

	return models.IndoorSensorData{
		DeviceID:    deviceID,
		Temperature: float32(temperature),
		Humidity:    float32(humidity),
		LedState:    ledState,
	}, nil
}

// ToggleWarningLight toggles the blinking yellow light on the Arduino
func (usb *usbClientImpl) ToggleWarningLight(openWindows bool) error {
	command := "1"
	if openWindows {
		command = "0"
	}

	_, err := usb.getArduinoResponse(command, 50*time.Millisecond)
	if err != nil {
		return err
	}

	if openWindows {
		fmt.Println("Warning light OFF")
	} else {
		fmt.Println("Warning light ON")
	}

	return nil
}

// getArduinoResponse sends the Arduino a command and retrieves the response
func (usb *usbClientImpl) getArduinoResponse(command string, sleepDuration time.Duration) (string, error) {
	startTime := time.Now()
	maxDuration := time.Millisecond * 500

	fmt.Printf("Waiting for command '%s' ack...\n", command)

	for {
		if err := usb.writeData(command); err != nil {
			return "", err
		}

		response, err := usb.readData()
		if err != nil {
			return "", err
		}

		if utils.IsValidResponse(response) {
			fmt.Printf("Arduino says: %s\n", response)
			return response, nil
		}

		if time.Since(startTime) >= maxDuration {
			return response, nil
		}

		time.Sleep(sleepDuration)
	}
}

func (usb *usbClientImpl) readData() (string, error) {
	buffer := make([]byte, 32)
	n, err := usb.port.Read(buffer)
	if err != nil {
		return "", err
	}
	if n > 0 {
		return string(bytes.TrimSpace(buffer[:n])), nil
	}
	return "", nil
}

func (usb *usbClientImpl) writeData(data string) error {
	_, err := usb.port.Write([]byte(data))
	return err
}
