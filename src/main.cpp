#include <iostream>
#include <string>
#include <sstream>
#include <thread>
#include <chrono>
#include <dotenv.h>
#include "UsbCommunication/UsbCommunication.h"
#include "RestApiHandler/RestApiHandler.h"

const char GET_URL[] = "localhost:5000/weather/outdoor-dewpoint";
const char POST_URL_SENSOR_FEED[] = "localhost:5000/discord/sensor-feed";
const char POST_URL_WINDOW_ALERT[] = "localhost:5000/discord/window-alert";
const char POST_URL_HUMIDITY_ALERT[] = "localhost:5000/discord/humidity-alert";

int main() {
    const int maxAttempts = 3;
    int attempt = 0;

    // TODO: fix Arduino wake issue
    while (attempt < maxAttempts) {
        try {
            auto start = std::chrono::high_resolution_clock::now();
            dotenv::init("../.env");
            RestApiHandler api("arDEWino/0.xx", POST_URL_SENSOR_FEED, POST_URL_WINDOW_ALERT, POST_URL_HUMIDITY_ALERT);
            const char* portname = std::getenv("ARDUINO_PORT");
            double indoorTemperature;
            double indoorHumidity;
            double indoorDewpoint;

            // get outdoor dewpoint from rest api
            std::string getRequestResponse = api.sendGetRequest(GET_URL);
            std::cout << "GET Response from API: " << getRequestResponse << std::endl;
            double outdoorDewpoint = std::stod(getRequestResponse);

            // establish serial connection
            UsbCommunication usbComm(portname, B9600);
            if (!usbComm.openPort()) {
                return EXIT_FAILURE;
            }
            if (!usbComm.configurePort()) {
                usbComm.closePort();
                return EXIT_FAILURE;
            }

            // read sensor data
            std::string response = usbComm.getArduinoResponse(&usbComm, "d", 11, 50);

            // parse sensor data
            std::stringstream ss(response);
            std::string T, RH;
            std::getline(ss, T, ',');
            std::getline(ss, RH, ',');
            indoorTemperature = std::stod(T);
            indoorHumidity = std::stod(RH);
            indoorDewpoint = api.dewPointCalculator(indoorTemperature, indoorHumidity);

            // debug output
            std::cout << "indoorTemperature = " << indoorTemperature << std::endl;
            std::cout << "indoorHumidity = " << indoorHumidity << std::endl;
            std::cout << "outdoorDewpoint = " << outdoorDewpoint << std::endl;
            std::cout << "indoorDewpoint = " << indoorDewpoint << std::endl;
            std::cout << "dewpointDelta = " << (indoorDewpoint - outdoorDewpoint) << std::endl;

            // send sensor data to the rest api
            api.sendSensorFeed(indoorTemperature, indoorHumidity, indoorDewpoint);

            // send window alert to the rest api and toggle arduino warning light
            if ((indoorDewpoint - outdoorDewpoint) > -1.0) {
                std::string newResponse = usbComm.getArduinoResponse(&usbComm, "0", 1, 50);
                std::cout << "0 written" << std::endl;
            } else {
                std::string newerResponse = usbComm.getArduinoResponse(&usbComm, "1", 1, 50);
                std::cout << "1 written" << std::endl;
                api.sendWindowAlert(indoorDewpoint, outdoorDewpoint);
            }

            // send humidity alert to the rest api
            if (indoorHumidity > 56) {
                api.sendHumidityAlert(indoorHumidity);
            }

            usbComm.closePort();
            auto stop = std::chrono::high_resolution_clock::now();
            auto duration = std::chrono::duration_cast<std::chrono::milliseconds>(stop - start).count();
            std::cout << "Program took " << duration << " ms" << std::endl;
            break;
        } catch (const std::exception& e) {
            attempt++;
            std::cerr << "Error occurred: " << e.what()
                      << ". Attempt " << attempt << " of " << maxAttempts
                      << ". Restarting..." << std::endl;

            if (attempt == maxAttempts) {
                std::cerr << "Maximum attempts reached. Exiting program." << std::endl;
                return EXIT_FAILURE;
            }
       }
    }
    return EXIT_SUCCESS;
}
