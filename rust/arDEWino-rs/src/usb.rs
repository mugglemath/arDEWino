use std::error::Error;
use std::io::{self, Read, Write};
use std::thread;
use std::time::Duration;
use std::time::Instant;

use regex::Regex;

use crate::models::IndoorSensorData;
use serialport::{DataBits, FlowControl, Parity, SerialPort, StopBits};

pub struct UsbCommunication {
    port: Box<dyn SerialPort>,
}

impl UsbCommunication {
    pub fn new(port_name: &str) -> io::Result<Self> {
        let port = serialport::new(port_name, 115200)
            .timeout(Duration::from_millis(100))
            .data_bits(DataBits::Eight)
            .parity(Parity::None)
            .stop_bits(StopBits::One)
            .flow_control(FlowControl::None)
            .open()?;

        Ok(UsbCommunication { port })
    }

    fn write_data(&mut self, data: &str) -> io::Result<()> {
        self.port.write_all(data.as_bytes())?;
        Ok(())
    }

    fn read_data(&mut self) -> io::Result<String> {
        let mut buffer = [0u8; 32];
        match self.port.read(&mut buffer) {
            Ok(n) if n > 0 => {
                let result = String::from_utf8_lossy(&buffer[..n]).to_string();
                Ok(result)
            }
            Ok(_) => Ok(String::new()),
            Err(e) => {
                if e.kind() == io::ErrorKind::TimedOut {
                    Ok(String::new())
                } else {
                    Err(e)
                }
            }
        }
    }

    fn get_arduino_response(&mut self, command: &str, sleep_duration: u64) -> io::Result<String> {
        let start_time = Instant::now();
        let max_duration = Duration::from_millis(1000);

        println!("Waiting for command '{}' ack...", command);

        loop {
            self.write_data(command)?;
            let response = self.read_data()?;
            let trimmed_response = response.trim();

            if is_valid_response(trimmed_response) {
                println!("Arduino says: {}", response);
                return Ok(response);
            }
            if start_time.elapsed() >= max_duration {
                return Err(io::Error::new(
                    io::ErrorKind::TimedOut,
                    "Arduino not responding...",
                ));
            }
            thread::sleep(Duration::from_millis(sleep_duration));
        }
    }

    pub fn get_indoor_sensor_data(
        usb_comm: &mut UsbCommunication,
    ) -> Result<IndoorSensorData, Box<dyn Error>> {
        let command = "d";
        let arduino_data = usb_comm.get_arduino_response(command, 50)?;
        let parts: Vec<&str> = arduino_data.split(',').collect();
        
        if parts.len() < 3 {
            return Err("Invalid data format from Arduino".into());
        }
    
        let device_id: u16 = parts[0].trim().parse()?;
        let temperature: f64 = parts[1].trim().parse()?;
        let humidity: f64 = parts[2].trim().parse()?;
    
        Ok(IndoorSensorData {
            device_id,
            temperature,
            humidity,
        })
    }

    pub fn toggle_warning_light(
        usb_comm: &mut UsbCommunication,
        keep_windows: bool,
    ) -> Result<(), Box<dyn Error>> {
        if keep_windows {
            usb_comm.get_arduino_response("0", 50)?;
            println!("Warning light OFF");
        } else {
            usb_comm.get_arduino_response("1", 50)?;
            println!("Warning light ON");
        }
        Ok(())
    }
}

fn is_valid_response(response: &str) -> bool {
    !response.is_empty() && (is_valid_data_format(response) || response == "a")
}

pub fn is_valid_data_format(input: &str) -> bool {
    let pattern = Regex::new(r"^\d{1,5},\d{2}\.\d{2},\d{2}\.\d{2}$").unwrap();
    pattern.is_match(input)
}
