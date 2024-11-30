package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/mugglemath/dewdrop-go/internal/requests"
	"github.com/mugglemath/dewdrop-go/internal/usb"
	"github.com/mugglemath/dewdrop-go/internal/wifi"
	"github.com/mugglemath/dewdrop-go/pkg/calculations"
	"github.com/mugglemath/dewdrop-go/pkg/models"
)

func main() {
	startTime := time.Now()
	err := godotenv.Load()
	if err != nil {
		fmt.Println("error loading env variables")
		return
	}

	mode := os.Args[1]
	fmt.Printf("Running in %s mode\n", mode)

	var indoorData models.IndoorSensorData
	var outdoorDewpoint float32
	var wg sync.WaitGroup

	wg.Add(2)

	// fetch indoor data asynchronously
	go func() {
		defer wg.Done()
		if mode == "wifi" {
			arduinoIP := os.Getenv("ARDUINO_IP")
			endpoint := fmt.Sprintf("%s/data", arduinoIP)
			data, err := wifi.GetIndoorSensorData(endpoint)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			indoorData = data
		} else if mode == "usb" {
			portName := os.Getenv("ARDUINO_PORT")
			usbComm, err := usb.NewUsbCommunication(portName)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			data, err := usbComm.GetIndoorSensorData()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			indoorData = data
		} else {
			fmt.Printf("Invalid mode: %s\n", mode)
			os.Exit(1)
		}
	}()

	// fetch outdoor dewpoint asynchronously
	go func() {
		defer wg.Done()
		dewpoint, err := requests.GetOutdoorDewpoint()
		if err != nil {
			fmt.Println("Error fetching outdoor dewpoint:", err)
			os.Exit(1)
		}
		outdoorDewpoint = dewpoint
	}()

	wg.Wait()

	// prepare sensor feed data
	ledState := indoorData.LedState
	indoorDewpoint, err := calculations.DewPointCalculator(float64(indoorData.Temperature),
		float64(indoorData.Humidity))
	if err != nil {
		fmt.Println("dew point calculation error")
	}
	dewpointDelta := indoorDewpoint - float64(outdoorDewpoint)
	openWindows := dewpointDelta > -1.0
	humidityAlert := indoorData.Humidity > 60.0

	payload, err := requests.PrepareSensorFeedJSON(&indoorData, float32(indoorDewpoint),
		outdoorDewpoint, float32(dewpointDelta), openWindows, humidityAlert)
	if err != nil {
		fmt.Println(err)
		return
	}

	wg.Add(2)

	// post sensor feed data asynchronously
	go func() {
		defer wg.Done()
		if err = requests.PostSensorFeed(payload); err != nil {
			fmt.Println("Error posting sensor feed:", err)
		}
	}()

	// toggle the warning light asynchronously
	go func() {
		defer wg.Done()
		if openWindows == ledState {
			if mode == "usb" {
				var usbComm *usb.UsbCommunication
				usbComm, err = usb.NewUsbCommunication(os.Getenv("ARDUINO_PORT"))
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				if err = usbComm.ToggleWarningLight(openWindows); err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			} else if mode == "wifi" {
				if err = wifi.ToggleWarningLight(openWindows); err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}
		}
	}()
	wg.Wait()

	fmt.Printf("Indoor Temperature: %.2f\n", indoorData.Temperature)
	fmt.Printf("Indoor Humidity: %.2f\n", indoorData.Humidity)
	fmt.Printf("Outdoor Dewpoint: %.2f\n", outdoorDewpoint)
	fmt.Printf("Indoor Dewpoint: %.2f\n", indoorDewpoint)
	fmt.Printf("Dewpoint Delta: %.2f\n", dewpointDelta)
	fmt.Printf("Open Windows: %v\n", openWindows)
	fmt.Printf("Humidity Alert: %v\n", humidityAlert)
	fmt.Printf("Sensor Feed JSON Data: %s\n", string(payload))

	fmt.Println("Program completed successfully.")
	elapsedTime := time.Since(startTime)
	fmt.Printf("Total execution time: %v\n", elapsedTime)
}