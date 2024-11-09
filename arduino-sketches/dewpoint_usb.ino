#include <Arduino.h>
#include <Wire.h>
#include "Adafruit_SHT31.h"
#include <SoftwareSerial.h>

#define SERIAL_BAUD_RATE 9600

const uint8_t SHT31_I2C_ADDRESS = 0x44;
bool ledState = false;
const int blinkDuration = 500;

Adafruit_SHT31 sht31 = Adafruit_SHT31();


void setup() {
  Serial.begin(SERIAL_BAUD_RATE);
  if (!sht31.begin(SHT31_I2C_ADDRESS)) {
    Serial.println("Couldn't find SHT31 sensor!");
    while (1) delay(1);
  }
  pinMode(LED_BUILTIN, OUTPUT);
  Serial.print("ready");
}


void loop() {
  if (Serial.available() > 0) {
    char incomingByte = Serial.read();
    if (incomingByte == 'd') {
      float temperature = sht31.readTemperature();
      float humidity = sht31.readHumidity();
      if (!isnan(temperature) && !isnan(humidity)) {
        Serial.flush();
        Serial.print(temperature);
        Serial.print(",");
        Serial.print(humidity);
        // Serial.println();
      } else {
        Serial.println(-1);
      }
    }
    else if (incomingByte == '1' || incomingByte == '0') {
        ledState = (incomingByte == '1');
        Serial.flush();
        Serial.print("a");
    }
  }
  if (ledState) {
    blinkLED();
  }
}


void blinkLED() {
  digitalWrite(LED_BUILTIN, HIGH);
  delay(blinkDuration);
  digitalWrite(LED_BUILTIN, LOW);
  delay(blinkDuration);
}
