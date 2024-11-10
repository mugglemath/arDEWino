#include <iostream>
#include <string>
#include <sstream>
#include <thread>
#include <chrono>
#include <dotenv.h>
#include "UsbCommunication/UsbCommunication.h"
#include "RestApiHandler/RestApiHandler.h"

const std::string GET_URL = "localhost:5000/weather/outdoor-dewpoint";
const std::string POST_URL = "localhost:5000/arduino/sensor-values";

int main() {
    auto start = std::chrono::high_resolution_clock::now();
    dotenv::init("../.env");    

    RestApiHandler api("arDEWino/0.xx");
    const char* portname = std::getenv("ARDUINO_PORT");
    std::string getRequestResponse = api.sendGetRequest(GET_URL);
    std::cout << "GET Response from API: " << getRequestResponse << std::endl;
    double outdoorDewpoint = std::stod(getRequestResponse);
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

    // api test
    std::string jsonData = R"({"indoorTemperature": )" + std::to_string(indoorTemperature) + R"(, "indoorHumidity": )" + std::to_string(indoorHumidity) + R"(})";
    std::string postRequestResponse = api.sendPostRequest(POST_URL, jsonData);
    std::cout << "POST Response from API: " << postRequestResponse << std::endl;

    // debug
    std::cout << "indoorTemperature = " << indoorTemperature << std::endl;
    std::cout << "indoorHumidity = " << indoorHumidity << std::endl;
    std::cout << "outdoorDewpoint = " << outdoorDewpoint << std::endl;
    std::cout << "indoorDewpoint = " << indoorDewpoint << std::endl;
    std::cout << "diff = " << (indoorDewpoint - outdoorDewpoint) << std::endl;

    // write 1 if outdoor > indoor
    if ((indoorDewpoint - outdoorDewpoint) > -1.0) {
        std::string newResponse = usbComm.getArduinoResponse(&usbComm, "0", 1, 50);
        std::cout << "0 written" << std::endl;
    } else {
        std::string newerResponse = usbComm.getArduinoResponse(&usbComm, "1", 1, 50);
        std::cout << "1 written" << std::endl;
    }

    usbComm.closePort();
    auto stop = std::chrono::high_resolution_clock::now();
    auto duration = std::chrono::duration_cast<std::chrono::milliseconds>(stop - start).count();
    std::cout << "Program took " << duration << " ms" << std::endl;
    return EXIT_SUCCESS;
}
