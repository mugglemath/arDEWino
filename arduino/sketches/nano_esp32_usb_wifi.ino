#include <WiFi.h>
#include <WebServer.h>
#include "Adafruit_SHT31.h"
#include <UUID.h>
#include <Preferences.h>

const char* SSID = "<your-router-ssid>";
const char* PASSWORD = "your-router-password";
const uint8_t SHT31_I2C_ADDRESS = 0x44;
const int BLINK_DURATION = 500;
const int NUM_READINGS = 48;
const long INTERVAL = 10000;

String uuidStr;
uint16_t hash = 0;
bool ledState = false;
float temperatureReadings[NUM_READINGS];
float humidityReadings[NUM_READINGS];
int readingIndex = 0;
unsigned long previousMillis = 0;
float humidityOffset = 0;
float temperatureOffset = 0;

Preferences preferences;
Adafruit_SHT31 sht31 = Adafruit_SHT31();
WebServer server(80);

void initWiFi();
void handleSerialInput();
void handleDataRequest();
void handleLedRequest();
float averageTemperature();
float averageHumidity();
void readSensorData();
void blinkLED();

void setup() {
  Serial.begin(115200);
  initWiFi();

  if (!sht31.begin(SHT31_I2C_ADDRESS)) {
    while (1) delay(1);
  }

  pinMode(LED_BUILTIN, OUTPUT);

  float initialTemperature = sht31.readTemperature();
  float initialHumidity = sht31.readHumidity();

  for (int i = 0; i < NUM_READINGS; i++) {
    temperatureReadings[i] = initialTemperature;
    humidityReadings[i] = initialHumidity;
  }

  // store UUID in flash memory
  preferences.begin("uuid", false);
  // check if UUID already exists
  if (!preferences.isKey("uuid")) {
    UUID uuidGenerate = UUID();
    uuidGenerate.generate();
    uuidStr = uuidGenerate.toCharArray();
    preferences.putString("uuid", uuidStr);
  } else {
    uuidStr = preferences.getString("uuid");
  }
  preferences.end();

  int len = uuidStr.length();
  for (int i = 0; i < len; i++) {
    hash = (hash * 31) + uuidStr[i];
  }

  Serial.print("UUID: ");
  Serial.println(uuidStr);
  Serial.print("Hash: ");
  Serial.println(hash);

  server.on("/data", HTTP_GET, handleDataRequestTwo);
  server.on("/led", HTTP_POST, handleLedRequest);
  server.begin();
}

void loop() {
  server.handleClient();
  delay(1);
  readSensorData();
  handleSerialInput();

  if (ledState) {
    blinkLED();
  }
}

void initWiFi() {
  WiFi.mode(WIFI_STA);
  WiFi.begin(SSID, PASSWORD);

  Serial.print("Connecting to WiFi ..");

  while (WiFi.status() != WL_CONNECTED) {
    Serial.print('.');
    delay(1000);
  }

  Serial.println("Connected!");
  Serial.print("IP Address: ");
  Serial.println(WiFi.localIP());
}

void handleSerialInput() {
  if (Serial.available() > 0) {
    char incomingByte = Serial.read();

    if (incomingByte == 'd') {
      float avgTemp = averageTemperature();
      float avgHumid = averageHumidity();

      if (!isnan(avgTemp) && !isnan(avgHumid)) {
        Serial.flush();
        Serial.print(avgTemp);
        Serial.print(",");
        Serial.print(avgHumid);
        Serial.println();
      } else {
        Serial.println(-1);
      }
    } else if (incomingByte == '1' || incomingByte == '0') {
      ledState = (incomingByte == '1');
      digitalWrite(LED_BUILTIN, ledState ? HIGH : LOW);
      Serial.flush();
      Serial.print("a");
    }
  }
}

void handleDataRequest() {
    String message = String(hash) + "," + String(averageTemperature()) + "," + String(averageHumidity()) + "\n";
    server.send(200, "text/plain", message);
}

void handleLedRequest() {
  if (server.hasArg("state")) {
    String state = server.arg("state");
    ledState = (state == "1");
    digitalWrite(LED_BUILTIN, ledState ? HIGH : LOW);
    server.send(200, "text/plain", "LED state changed to " + state);
  } else {
    server.send(400, "text/plain", "Bad Request: Missing 'state' parameter");
  }
}

float averageTemperature() {
  float totalTemperature = 0;

  for (int i = 0; i < NUM_READINGS; i++) {
    totalTemperature += temperatureReadings[i];
  }

  return (NUM_READINGS > 0) ? totalTemperature / NUM_READINGS : 0;
}

float averageHumidity() {
  float totalHumidity = 0;

  for (int i = 0; i < NUM_READINGS; i++) {
    totalHumidity += humidityReadings[i];
  }

  return (NUM_READINGS > 0) ? totalHumidity / NUM_READINGS : 0;
}

void readSensorData() {
  unsigned long currentMillis = millis();

  if (currentMillis - previousMillis >= INTERVAL) {
    previousMillis = currentMillis;

    float currentTemperature = sht31.readTemperature() - temperatureOffset;
    float currentHumidity = sht31.readHumidity() - humidityOffset;

    temperatureReadings[readingIndex] = currentTemperature;
    humidityReadings[readingIndex] = currentHumidity;

    readingIndex = (readingIndex + 1) % NUM_READINGS;
  }
}

void blinkLED() {
  digitalWrite(LED_BUILTIN, HIGH);
  delay(BLINK_DURATION);
  digitalWrite(LED_BUILTIN, LOW);
}