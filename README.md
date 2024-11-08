# arduino_dewpoint

Compares outdoor dewpoint with indoor dewpoint using National Weather Service API and an Arduino + SHT31 humidity sensor.

Blinks yellow if outdoor dewpoint is more than 1 degree Celsius higher than indoor dewpoint.

# TODO

* switch hardware to communicate over BLE using battery
* implement power consumption monitoring
* optimize power consumption
* log power usage over time
* log sensor readings over time
* graph logs
