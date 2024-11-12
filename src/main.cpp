#include <iostream>
#include <string>
#include <sstream>
#include <thread>
#include <chrono>
#include <dotenv.h>
#include <cmath>
#include <exception>
#include "UsbCommunication/UsbCommunication.h"
#include "RestApiHandler/RestApiHandler.h"

const char GET_URL[] = "localhost:5000/weather/outdoor-dewpoint";
const char POST_URL_SENSOR_FEED[] = "localhost:5000/discord/sensor-feed";
const char POST_URL_WINDOW_ALERT[] = "localhost:5000/discord/window-alert";
const char POST_URL_HUMIDITY_ALERT[] = "localhost:5000/discord/humidity-alert";

void parseSensorData(const std::string& response, double& temperature, double& humidity);
void logSensorData(double temperature, double humidity, double outdoorDewpoint, double indoorDewpoint);
void handleAlerts(UsbCommunication& usbComm, RestApiHandler& api, double indoorDewpoint, double outdoorDewpoint, double indoorHumidity);
void logDuration(const decltype(std::chrono::high_resolution_clock::now())& start, const decltype(std::chrono::high_resolution_clock::now())& stop);


int main() {
    const int maxAttempts = 3;
    int attempt = 0;

    while (attempt < maxAttempts) {
        try {
            auto start = std::chrono::high_resolution_clock::now();
            dotenv::init("../.env");
            RestApiHandler api("arDEWino/0.xx", POST_URL_SENSOR_FEED, POST_URL_WINDOW_ALERT, POST_URL_HUMIDITY_ALERT);
            const char* portname = std::getenv("ARDUINO_PORT");
            UsbCommunication usbComm(portname, B9600);

            // Establish serial connection
            if (!usbComm.openPort() || !usbComm.configurePort()) {
                throw std::runtime_error("Failed to open or configure USB port");
            }

            // Get outdoor dewpoint from REST API
            double outdoorDewpoint = std::stod(api.sendGetRequest(GET_URL));

            // Read sensor data
            std::string response = usbComm.getArduinoResponse(&usbComm, "d", 50);
            double indoorTemperature, indoorHumidity;
            parseSensorData(response, indoorTemperature, indoorHumidity);

            double indoorDewpoint = api.dewPointCalculator(indoorTemperature, indoorHumidity);

            // Debug output
            logSensorData(indoorTemperature, indoorHumidity, outdoorDewpoint, indoorDewpoint);

            // Send sensor data to REST API
            api.sendSensorFeed(indoorTemperature, indoorHumidity, indoorDewpoint);
            handleAlerts(usbComm, api, indoorDewpoint, outdoorDewpoint, indoorHumidity);

            usbComm.closePort();
            auto stop = std::chrono::high_resolution_clock::now();
            logDuration(start, stop);

            break;
        } catch (const std::exception& e) {
            attempt++;
            std::cerr << "Error occurred: " << e.what()
                      << ". Attempt " << attempt << " of " << maxAttempts << ". Restarting..." << std::endl;
            std::this_thread::sleep_for(std::chrono::milliseconds(1000));

            if (attempt == maxAttempts) {
                std::cerr << "Maximum attempts reached. Exiting program." << std::endl;
                return EXIT_FAILURE;
            }
        }
    }
    return EXIT_SUCCESS;
}


void parseSensorData(const std::string& response, double& temperature, double& humidity) {
    std::stringstream ss(response);
    std::string T, RH;
    std::getline(ss, T, ',');
    std::getline(ss, RH, ',');
    if (T.length() < 4 || RH.length() < 5) {
        throw std::invalid_argument("Invalid sensor data received.");
    }
    temperature = std::stod(T);
    humidity = std::stod(RH);
}

void logSensorData(double temperature, double humidity, double outdoorDewpoint, double indoorDewpoint) {
    std::cout << "Indoor Temperature: " << temperature << "\n"
              << "Indoor Humidity: " << humidity << "\n"
              << "Outdoor Dewpoint: " << outdoorDewpoint << "\n"
              << "Indoor Dewpoint: " << indoorDewpoint << "\n"
              << "Dewpoint Delta: " << (indoorDewpoint - outdoorDewpoint) << "\n";
}

void handleAlerts(UsbCommunication& usbComm, RestApiHandler& api, double indoorDewpoint, double outdoorDewpoint, double indoorHumidity) {
    if ((indoorDewpoint - outdoorDewpoint) > -1.0) {
        usbComm.getArduinoResponse(&usbComm, "0", 50);
        std::cout << "Warning light OFF" << std::endl;
    } else {
        usbComm.getArduinoResponse(&usbComm, "1", 50);
        api.sendWindowAlert(indoorDewpoint, outdoorDewpoint);
        std::cout << "Warning light ON" << std::endl;
    }

    if (indoorHumidity > 57) {
        api.sendHumidityAlert(indoorHumidity);
    }
}

void logDuration(const decltype(std::chrono::high_resolution_clock::now())& start,
                 const decltype(std::chrono::high_resolution_clock::now())& stop) {
    auto duration = std::chrono::duration_cast<std::chrono::milliseconds>(stop - start).count();
    std::cout << "Program took " << duration << " ms" << std::endl;
}
