# arDEWino: An Arduino-based Dew Point Monitor
## Overview

<div align="center">
    <img src="images/overview.png" alt="Overview" style="width:30%; margin-right:10%;">
    <img src="images/temp-humidity.png" alt="Temp / Humidity" style="width:30%; margin-right:10%;">
    <img src="images/discord-feed.jpeg" alt="Discord Feed" style="width:12%;">
</div>


arDEWino is an Arduino-based dew point monitor designed to help maintain optimal indoor humidity levels and prevent mold growth. The system utilizes an Arduino Nano ESP32 microcontroller and a SHT31 temperature and humidity sensor to measure indoor conditions. It then compares the indoor dew point to the outdoor dew point retrieved from the National Weather Service.
arDEWino can notify users when the outdoor dew point is higher than the indoor dew point, indicating that it's best to keep windows closed to prevent excess humidity from entering the home. The system can also be used for real-time temperature and humidity alerts, making it suitable for monitoring conditions for pets, instruments, or anything else sensitive to climate changes.

## Current Components/Services
* arDEWino-rs:
  * Retrieves sensor data from Arduino over WiFi/USB
* go-dew:
  * REST API that handles outdoor weather information, sending Discord notifications/alerts, and the database
* [TimescaleDB](https://www.timescale.com)
  * Open source time series database that extends Postgres
* [Grafana](https://grafana.com):
  * Open source interactive visualization and monitoring web app that connects to your data sources

## How to Use
This app is currently in early development and can be used locally with either Docker if you prefer to use containers, or compile the Rust and Go apps and install TimescaleDB and Grafana, or some combination of both. A Release/Package coming soon!

### Requirements
* Arduino Nano ESP32 Microcontroller
* SHT31 Temperature/Humidity Sensor
* Arduino IDE
* Docker

### Local Setup with Docker
1. Connect sensor to microcontroller (use jumper cables to an Arduino with headers if you don't want to solder)
2. Set the SSID and Password variables in [arduino/sketches/nano_esp32_usb_wifi.ino](/arduino/sketches/nano_esp32_usb_wifi.ino)
3. Upload this sketch file with your WiFi information to your Arduino using Arduino IDE
4. Set the environment variables in `docker/compose.yml` or use the `env_file:` directive with .env files in their respective directories

> [!IMPORTANT]  
> arDEWino uses the National Weather Service API to retrieve the outdoor dewpoint nearest you using *decimal degrees*.
> When you drop a pin, the format can vary depending on the mapping application.
> NWS requires values like the Google Maps format which includes the negative sign.

```
Map App:                                         Environment Variables:

Apple Maps: 40.74867° N,73.98628° W              LATITUDE=40.74867
Google Maps: (40.74867,-73.98628)                LONGITUDE=-73.98628
```
> [!TIP]
> [Here is how you generate and use Discord webhook URLs](https://support.discord.com/hc/en-us/articles/228383668-Intro-to-Webhooks)
5. Change into the docker directory and run `docker compose up -d`
6. Open `localhost:3000` in your web browser

> [!WARNING]
> TimescaleDB and Grafana come with default usernames and passwords. Change before deploying.
