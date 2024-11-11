# arDEWino: An Arduino-based Dewpoint Monitor
## Overview

arDEWino is an Arduino-based dewpoint monitor designed to help maintain optimal indoor humidity levels and prevent mold growth. The system utilizes an Arduino Nano microcontroller and a SHT31 temperature and humidity sensor to measure indoor conditions. It then compares the indoor dewpoint to the outdoor dewpoint retrieved from the National Weather Service.
arDEWino can notify users when the outdoor dewpoint is higher than the indoor dewpoint, indicating that it's best to keep windows closed to prevent excess humidity from entering the home. The system can also be used for real-time temperature and humidity alerts, making it suitable for monitoring conditions for pets, instruments, or anything else sensitive to climate changes.

## Current Functionality
Computes indoor dewpoint and compares to outdoor dewpoint. Retrieves indoor temperature and humidity data from an Arduino Nano + SHT31 sensor over USB every 10 minutes. Onboard sensor readings are collected every few seconds and the device returns a simple moving average.
Communicates with a [REST API](https://github.com/mugglemath/dew-serve) to retrieve outdoor weather information and pass sensor data.
Routes notifications and sensor data via Discord webhooks.
Also, the microcontroller will blink yellow if outdoor dewpoint is more than 1 degree Celsius higher than indoor dewpoint.

## TODO
### Hardware
* Upgrade to Nano 33 BLE for communication over Bluetooth Low Energy
* Switch to battery power

### Software

#### Arduino
* Use sleep and interrupts on the new board

#### API
* Validate data
* Set up a database
* Log power usage over time
* Log sensor readings over time

#### CPP App
* Validate data
* Improve error handling
* Add tests after Nano BLE 33 upgrade

#### Visualizations
* Graph energy consumption
* Graph sensor data
