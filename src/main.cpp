#include <iostream>
#include <string>
#include <sstream>
#include <thread>
#include <chrono>
#include <dotenv.h>
#include "UsbCommunication/UsbCommunication.h"
#include "WeatherApi/WeatherApi.h"


int main() {
    dotenv::init("../.env");
    std::string office = dotenv::getenv("OFFICE");
    std::string gridX = dotenv::getenv("GRID_X");
    std::string gridY = dotenv::getenv("GRID_Y");
    std::string userAgent = dotenv::getenv("USER_AGENT");
    const char* portname = std::getenv("ARDUINO_PORT");
    WeatherApi api(office, gridX, gridY, userAgent);
    std::string apiResponse = api.getWeatherData();
    double outdoorDewpoint = api.extractDewPoint(apiResponse);
    double indoorTemperature;
    double indoorHumidity;
    double indoorDewpoint;

    // establish serial connection
    UsbCommunication usbComm(portname, B9600);
    if (!usbComm.openPort()) {
        return EXIT_FAILURE;
    }
    if (!usbComm.configurePort()) {
        usbComm.closePort();
        return EXIT_FAILURE;
    }

    // read data
    std::string response = usbComm.getArduinoResponse(&usbComm, "d", 11, 50);

    // parse data
    std::stringstream ss(response);
    std::string T, RH;
    std::getline(ss, T, ',');
    std::getline(ss, RH, ',');

    indoorTemperature = std::stod(T);
    indoorHumidity = std::stod(RH);
    indoorDewpoint = api.dewPointCalculator(indoorTemperature, indoorHumidity);

    // debug
    std::cout << "indoorTemperature = " << indoorTemperature << std::endl;
    std::cout << "indoorHumidity = " << indoorHumidity << std::endl;
    std::cout << "outdoorDewpoint = " << outdoorDewpoint << std::endl;
    std::cout << "indoorDewpoint = " << indoorDewpoint << std::endl;
    std::cout << "diff = " << (indoorDewpoint - outdoorDewpoint) << std::endl;

    // write 1 if outdoor > indoor
    if ((indoorDewpoint - outdoorDewpoint) > 1.0) {
        std::string newResponse = usbComm.getArduinoResponse(&usbComm, "0", 1, 50);
        std::cout << "0 written" << std::endl;
    } else {
        std::string newerResponse = usbComm.getArduinoResponse(&usbComm, "1", 1, 50);
        std::cout << "1 written" << std::endl;
    }

    usbComm.closePort();
    return EXIT_SUCCESS;
}
