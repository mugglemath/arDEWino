#include "RestApiHandler.h"
#include <iostream>
#include <math.h>
#include <string>
#include <cmath>

RestApiHandler::RestApiHandler(const std::string& userAgent, const std::string& postUrlSensorFeed, const std::string& postUrlWindowAlert, const std::string& postUrlHumidityAlert)
    : userAgent(userAgent), postUrlSensorFeed(postUrlSensorFeed), postUrlWindowAlert(postUrlWindowAlert), postUrlHumidityAlert(postUrlHumidityAlert) {}
RestApiHandler::RestApiHandler() : userAgent("cpp app") {}


size_t RestApiHandler::WriteCallback(void* contents, size_t size, size_t nmemb, std::string* userp) {
    size_t totalSize = size * nmemb;
    userp->append((char*)contents, totalSize);
    return totalSize;
}


std::string RestApiHandler::sendGetRequest(const std::string& url) {
    CURL* curl;
    CURLcode res;
    std::string readBuffer;

    curl = curl_easy_init();
    if (curl) {
        curl_easy_setopt(curl, CURLOPT_URL, url.c_str());
        curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, WriteCallback);
        curl_easy_setopt(curl, CURLOPT_WRITEDATA, &readBuffer);
        curl_easy_setopt(curl, CURLOPT_USERAGENT, userAgent.c_str());

        res = curl_easy_perform(curl);
        if (res != CURLE_OK) {
            std::cerr << "curl_easy_perform() failed: " << curl_easy_strerror(res) << std::endl;
        }

        curl_easy_cleanup(curl);
    }
    return readBuffer;
}


std::string RestApiHandler::sendPostRequest(const std::string& url, const std::string& jsonData) {
    CURL* curl;
    CURLcode res;
    std::string readBuffer;

    curl = curl_easy_init();
    if (curl) {
        curl_easy_setopt(curl, CURLOPT_URL, url.c_str());
        curl_easy_setopt(curl, CURLOPT_POSTFIELDS, jsonData.c_str());

        struct curl_slist *headers = NULL;
        headers = curl_slist_append(headers, "Content-Type: application/json");
        curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);

        curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, WriteCallback);
        curl_easy_setopt(curl, CURLOPT_WRITEDATA, &readBuffer);
        curl_easy_setopt(curl, CURLOPT_USERAGENT, userAgent.c_str());

        res = curl_easy_perform(curl);
        if (res != CURLE_OK) {
            std::cerr << "curl_easy_perform() failed: " << curl_easy_strerror(res) << std::endl;
        }

        curl_slist_free_all(headers);
        curl_easy_cleanup(curl);
    }
    return readBuffer;
}


double RestApiHandler::dewPointCalculator(double T, double RH) {
    return (243.04 * (log(RH / 100) + ((17.625 * T) / (243.04 + T)))) /
           (17.625 - log(RH / 100) - ((17.625 * T) / (243.04 + T)));
}


double RestApiHandler::roundToTwoDecimals(double value) {
    return std::round(value * 100.0) / 100.0;
}


void RestApiHandler::sendSensorFeed(double indoorTemperature, double indoorHumidity, double indoorDewpoint, double outdoorDewpoint) {
    double dewpointDelta = indoorDewpoint - outdoorDewpoint;
    bool openWindows = (dewpointDelta > -1);
    bool humidityAlert = (indoorHumidity > 57);
    std::string jsonDataSensorFeed =
        R"({"indoorTemperature": )" + std::to_string(indoorTemperature) +
        R"(, "indoorHumidity": )" + std::to_string(indoorHumidity) +
        R"(, "indoorDewpoint": )" + std::to_string(indoorDewpoint) +
        R"(, "outdoorDewpoint": )" + std::to_string(outdoorDewpoint) +
        R"(, "dewpointDelta": )" + std::to_string(dewpointDelta) +
        R"(, "openWindows": )" + (openWindows ? "true" : "false") +
        R"(, "humidityAlert": )" + (humidityAlert ? "true" : "false") +
        R"(})";

    // std::cout << "JSON Data to be sent: " << jsonDataSensorFeed << std::endl;
    std::string postRequestResponse = sendPostRequest(postUrlSensorFeed, jsonDataSensorFeed);
    std::cout << "POST Response from Sensor Feed API: " << postRequestResponse << std::endl;
}


void RestApiHandler::sendWindowAlert(double indoorDewpoint, double outdoorDewpoint) {
    double dewpointDelta = indoorDewpoint - outdoorDewpoint;
    bool openWindows = (dewpointDelta > -1);
    std::string jsonDataWindowAlert =
        R"({"indoorDewpoint": )" + std::to_string(indoorDewpoint) +
        R"(, "outdoorDewpoint": )" + std::to_string(outdoorDewpoint) +
        R"(, "dewpointDelta": )" + std::to_string(dewpointDelta) +
        R"(, "openWindows": )" + (openWindows ? "true" : "false") + R"(})";
    
    // std::cout << "JSON Data to be sent: " << jsonDataSensorFeed << std::endl;
    std::string postRequestResponse = sendPostRequest(postUrlWindowAlert, jsonDataWindowAlert);
    std::cout << "POST Response from Window Alert API: " << postRequestResponse << std::endl;
}


void RestApiHandler::sendHumidityAlert(double indoorHumidity) {
    bool humidityAlert = (indoorHumidity > 57);
    std::string jsonDataHumidityAlert =
        R"({"indoorHumidity": )" + std::to_string(indoorHumidity) + 
        R"(, "humidityAlert": )" + (humidityAlert ? "true" : "false") + R"(})";
    
    // std::cout << "JSON Data to be sent: " << jsonDataSensorFeed << std::endl;
    std::string postRequestResponse = sendPostRequest(postUrlHumidityAlert, jsonDataHumidityAlert);
    std::cout << "POST Response from Humidity Alert API: " << postRequestResponse << std::endl;
}
