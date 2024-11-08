#ifndef SRC_USBCOMMUNICATION_USBCOMMUNICATION_H_
#define SRC_USBCOMMUNICATION_USBCOMMUNICATION_H_

#include <fcntl.h>
#include <termios.h>
#include <unistd.h>
#include <iostream>
#include <cstring>
#include <string>

class UsbCommunication {
 public:
    UsbCommunication(const char* portname, speed_t baudRate);
    ~UsbCommunication();

    bool openPort();
    void closePort();
    bool configurePort();
    void writeData(const char* data);
    std::string readData();
    std::string getArduinoResponse(UsbCommunication* usbComm, const char* command, size_t expectedLength, int sleepDuration);

 private:
    const char* portname;
    speed_t baudRate;
    int fd;
};

#endif  // SRC_USBCOMMUNICATION_USBCOMMUNICATION_H_
