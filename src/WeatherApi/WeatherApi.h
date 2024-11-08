#ifndef SRC_WEATHERAPI_WEATHERAPI_H_
#define SRC_WEATHERAPI_WEATHERAPI_H_

#include <curl/curl.h>
#include <string>
#include <cmath>

class WeatherApi {
 public:
    WeatherApi(const std::string& office, const std::string& gridX, const std::string& gridY, const std::string& userAgent);
    WeatherApi();

    static size_t WriteCallback(void* contents, size_t size, size_t nmemb, std::string* userp);
    std::string getWeatherData();
    double extractDewPoint(const std::string& jsonResponse);
    static double dewPointCalculator(double T, double RH);

 private:
    std::string office;
    std::string gridX;
    std::string gridY;
    std::string userAgent;

    std::string constructUrl() const;
};

#endif  // SRC_WEATHERAPI_WEATHERAPI_H_
