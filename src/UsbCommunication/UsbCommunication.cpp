#include <iostream>
#include <string>
#include <thread>
#include <chrono>
#include "UsbCommunication.h"

UsbCommunication::UsbCommunication(const char* portname, speed_t baudRate)
    : portname(portname), baudRate(baudRate), fd(-1) {}
UsbCommunication::~UsbCommunication() {
    closePort();
}


bool UsbCommunication::openPort() {
    fd = open(portname, O_RDWR | O_NOCTTY | O_NDELAY);
    if (fd == -1) {
        std::cerr << "Error opening port: " << strerror(errno) << std::endl;
        return false;
    }
    return true;
}


void UsbCommunication::closePort() {
    if (fd != -1) {
        close(fd);
        fd = -1;
    }
}


bool UsbCommunication::configurePort() {
    struct termios options;
    if (tcgetattr(fd, &options) != 0) {
        std::cerr << "Error getting port attributes: " << strerror(errno) << std::endl;
        return false;
    }

    cfsetispeed(&options, baudRate);
    cfsetospeed(&options, baudRate);

    options.c_cflag |= (CLOCAL | CREAD);
    options.c_cflag &= ~PARENB;
    options.c_cflag &= ~CSTOPB;
    options.c_cflag &= ~CSIZE;
    options.c_cflag |= CS8;

    if (tcsetattr(fd, TCSANOW, &options) != 0) {
        std::cerr << "Error setting port attributes: " << strerror(errno) << std::endl;
        return false;
    }

    return true;
}


void UsbCommunication::writeData(const char* data) {
    write(fd, data, strlen(data));
}


std::string UsbCommunication::readData() {
    char buffer[12];
    int n = read(fd, buffer, sizeof(buffer) - 1);

    if (n > 0) {
        buffer[n] = '\0';
        return std::string(buffer);
    } else {
        return "";
    }
}

std::string UsbCommunication::getArduinoResponse(UsbCommunication* usbComm, const char* command, size_t expectedLength, int sleepDuration) {
    std::string response;
    std::cout << "Waiting for command '" << command << "' ack..." << std::endl;
    while (true) {
        usbComm->writeData(command);
        response = usbComm->readData();

        if (!response.empty() && response.size() == expectedLength) {
            std::cout << "Arduino says: " << response << std::endl;
            return response;
        } else {
            std::this_thread::sleep_for(std::chrono::milliseconds(sleepDuration));
        }
    }
    return response;
}
