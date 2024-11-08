#include <curl/curl.h>
#include <iostream>
#include <cmath>
#include <string>
#include "WeatherApi.h"
#include <nlohmann/json.hpp>

using json = nlohmann::json;

WeatherApi::WeatherApi(const std::string& office, const std::string& gridX, const std::string& gridY, const std::string& userAgent)
    : office(office), gridX(gridX), gridY(gridY), userAgent(userAgent) {}
WeatherApi::WeatherApi() : office("default_office"), gridX("0"), gridY("0"), userAgent("undefined") {}


size_t WeatherApi:: WriteCallback(void* contents, size_t size, size_t nmemb, std::string* userp) {
    size_t totalSize = size * nmemb;
    userp->append((char*)contents, totalSize);
    return totalSize;
}


std::string WeatherApi::constructUrl() const {
    return "https://api.weather.gov/gridpoints/" + office + "/" + gridX + "," + gridY;
}


std::string WeatherApi::getWeatherData() {
    CURL* curl;
    CURLcode res;
    std::string readBuffer;

    curl = curl_easy_init();

    std::string url = constructUrl().c_str();
    if (curl) {
        curl_easy_setopt(curl, CURLOPT_URL, constructUrl().c_str());
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


double WeatherApi::extractDewPoint(const std::string& jsonResponse) {
    double dewPoint = 0.0;
    try {
        json parsedJson = json::parse(jsonResponse);
        dewPoint = parsedJson["properties"]["dewpoint"]["values"][0]["value"];
    } catch (const json::exception& e) {
        std::cerr << "Error accessing dew point: " << e.what() << std::endl;
    }
    return dewPoint;
}


double WeatherApi::dewPointCalculator(double T, double RH) {
    return (243.04 * (log(RH / 100) + ((17.625 * T) / (243.04 + T)))) /
           (17.625 - log(RH / 100) - ((17.625 * T) / (243.04 + T)));
}
