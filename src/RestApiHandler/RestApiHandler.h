#ifndef SRC_RESTAPIHANDLER_RESTAPIHANDLER_H_
#define SRC_RESTAPIHANDLER_RESTAPIHANDLER_H_

#include <curl/curl.h>
#include <string>

class RestApiHandler {
 public:
    RestApiHandler(const std::string& userAgent, const std::string& postUrlSensorFeed, const std::string& postUrlWindowAlert, const std::string& postUrlHumidityAlert);
    RestApiHandler();

    static size_t WriteCallback(void* contents, size_t size, size_t nmemb, std::string* userp);
    std::string sendGetRequest(const std::string& url);
    std::string sendPostRequest(const std::string& url, const std::string& jsonData);
    double dewPointCalculator(double T, double RH);
    void sendSensorFeed(double indoorTemperature, double indoorHumidity, double indoorDewpoint);
    void sendWindowAlert(double indoorDewpoint, double outdoorDewpoint);
    void sendHumidityAlert(double indoorHumidity);

 private:
    std::string userAgent;
    std::string postUrlSensorFeed;
    std::string postUrlWindowAlert;
    std::string postUrlHumidityAlert;
};

#endif  // SRC_RESTAPIHANDLER_RESTAPIHANDLER_H_
