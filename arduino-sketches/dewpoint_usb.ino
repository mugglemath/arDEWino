#include <Arduino.h>
#include <Wire.h>
#include "Adafruit_SHT31.h"
#include <SoftwareSerial.h>

#define SERIAL_BAUD_RATE 9600
const uint8_t SHT31_I2C_ADDRESS = 0x44;
bool ledState = false;
SoftwareSerial mySerial(10, 11);

Adafruit_SHT31 sht31 = Adafruit_SHT31();


void setup() {
  // put your setup code here, to run once:

  Serial.begin(SERIAL_BAUD_RATE);

  if (!sht31.begin(SHT31_I2C_ADDRESS)) {
    Serial.println("Couldn't find SHT31 sensor!");
    while (1) delay(1);
  }

  mySerial.begin(9600);

  // built-in LED pin as output
  pinMode(LED_BUILTIN, OUTPUT);
}


void loop() {
  // put your main code here, to run repeatedly:

  float temperature = sht31.readTemperature();
  float humidity = sht31.readHumidity();

  if (!isnan(temperature) && !isnan(humidity)) {
    Serial.print(temperature);
    Serial.print(",");
    Serial.print(humidity);
    Serial.println();
  } else {
    Serial.println(-1);
  }
  
  delay(500);

  if (Serial.available() > 0) {
    char incomingByte = Serial.read();

    if (incomingByte == '1') {
      ledState = true;
    } else if (incomingByte == '0') {
      ledState = false;
    }
  }
  
  if (ledState) {
    digitalWrite(LED_BUILTIN, HIGH);
    delay(500);
    digitalWrite(LED_BUILTIN, LOW);
    delay(500);
  }
  // Serial.println(ledState);
}
