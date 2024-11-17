#include <Arduino.h>
#include <Wire.h>
#include "Adafruit_SHT31.h"
#include <SoftwareSerial.h>
#include <LowPower.h>
#include <avr/sleep.h>

#define SERIAL_BAUD_RATE 9600;

const uint8_t SHT31_I2C_ADDRESS = 0x44;
bool ledState = false;
const int blinkDuration = 500;
float averageTemperature = 0;
float averageHumidity = 0;
const int numReadings = 10;
float temperatureReadings[numReadings];
float humidityReadings[numReadings];
int index = 0;
unsigned long previousMillis = 0;
float humidityOffset = 3;
float temperatureOffset = 0.5;
const long interval = 10000;
const int rxPin = 2;
volatile bool wakeUpFlag = false;

Adafruit_SHT31 sht31 = Adafruit_SHT31();


void setup() {
  Serial.begin(SERIAL_BAUD_RATE);
  if (!sht31.begin(SHT31_I2C_ADDRESS)) {
    Serial.println("Couldn't find SHT31 sensor!");
    while (1) delay(1);
  }
  pinMode(LED_BUILTIN, OUTPUT);

  pinMode(rxPin, INPUT_PULLUP);
  attachInterrupt(digitalPinToInterrupt(rxPin), wakeUp, LOW);

  float initialTemperature = sht31.readTemperature() - temperatureOffset;
  float initialHumidity = sht31.readHumidity() - humidityOffset;

  for (int i = 0; i < numReadings; i++) {
    temperatureReadings[i] = initialTemperature;
    humidityReadings[i] = initialHumidity;
  }

  averageTemperature = initialTemperature;
  averageHumidity = initialHumidity;
  // Serial.print("ready");
}


void loop() {
  unsigned long currentMillis = millis();
  if (currentMillis - previousMillis >= interval) {
    previousMillis = currentMillis;

    float currentTemperature = sht31.readTemperature() - temperatureOffset;
    float currentHumidity = sht31.readHumidity() - humidityOffset;

    temperatureReadings[index] = currentTemperature;
    humidityReadings[index] = currentHumidity;

    index = (index + 1) % numReadings;
    averageTemperature = 0;
    averageHumidity = 0;

    for (int i = 0; i < numReadings; i++) {
      averageTemperature += temperatureReadings[i];
      averageHumidity += humidityReadings[i];
    }
    if (numReadings > 0) {
      averageTemperature /= numReadings;
      averageHumidity /= numReadings;
    }
  }

  if (Serial.available() > 0) {
    char incomingByte = Serial.read();
    if (incomingByte == 'd') {
      if (!isnan(averageTemperature) && !isnan(averageHumidity)) {
        Serial.flush();
        Serial.print(averageTemperature);
        Serial.print(",");
        Serial.print(averageHumidity);
        Serial.println();
      } else {
        Serial.println(-1);
      }
    } else if (incomingByte == '1' || incomingByte == '0') {
      ledState = (incomingByte == '1');
      Serial.flush();
      Serial.print("a");
    }
  }
  if (ledState) {
    blinkLED();
  }
  // set_sleep_mode(SLEEP_MODE_PWR_DOWN);
  // sleep_enable();
  // sleep_cpu();
  // sleep_disable();
  // LowPower.powerDown(SLEEP_8S, ADC_OFF, BOD_OFF);
  // delay(8000);
}


void wakeUp() {
  wakeUpFlag = true;
  Serial.println("Woke up!");
}


void blinkLED() {
  digitalWrite(LED_BUILTIN, HIGH);
  delay(blinkDuration);
  digitalWrite(LED_BUILTIN, LOW);
  delay(blinkDuration);
}