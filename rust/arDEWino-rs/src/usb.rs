use std::error::Error;
use std::time::{Duration, Instant};

use regex::Regex;
use tokio::time::sleep;
use tokio_serial::{SerialStream, SerialPortBuilderExt};
use tokio::io::{AsyncWriteExt, AsyncReadExt};

use crate::models::IndoorSensorData;

pub struct UsbCommunication {
    port: SerialStream,
}

impl UsbCommunication {
    pub async fn new(port_name: &str) -> Result<Self, Box<dyn Error>> {
        let port = tokio_serial::new(port_name, 115200)
            .open_native_async()?;
        Ok(UsbCommunication { port })
    }

    async fn write_data(&mut self, data: &str) -> Result<(), Box<dyn Error>> {
        self.port.write_all(data.as_bytes()).await?;
        Ok(())
    }

    async fn read_data(&mut self) -> Result<String, Box<dyn Error>> {
        let mut buffer = [0u8; 32];
        let n = self.port.read(&mut buffer).await?;
        
        if n > 0 {
            let result = String::from_utf8_lossy(&buffer[..n]).to_string();
            Ok(result)
        } else {
            Ok(String::new())
        }
    }

    pub async fn get_arduino_response(&mut self, command: &str, sleep_duration: u64) -> Result<String, Box<dyn Error>> {
        let start_time = Instant::now();
        let max_duration = Duration::from_millis(1000);

        println!("Waiting for command '{}' ack...", command);

        loop {
            self.write_data(command).await?;
            let response = self.read_data().await?;
            let trimmed_response = response.trim();

            if is_valid_response(trimmed_response) {
                println!("Arduino says: {}", response);
                return Ok(response);
            }
            if start_time.elapsed() >= max_duration {
                return Err("Arduino not responding...".into());
            }
            sleep(Duration::from_millis(sleep_duration)).await;
        }
    }

    pub async fn get_indoor_sensor_data(&mut self) -> Result<IndoorSensorData, Box<dyn Error>> {
        let command = "d";
        let arduino_data = self.get_arduino_response(command, 50).await?;
        let parts: Vec<&str> = arduino_data.split(',').collect();
        
        let device_id: u64 = parts[0].trim().parse()?;
        let temperature: f32 = parts[1].trim().parse()?;
        let humidity: f32 = parts[2].trim().parse()?;
        let led_state: bool = match parts[3].trim() {
            "1" => true,
            "0" => false,
            _ => return Err("Invalid LED state value".into()),
        };
    
        Ok(IndoorSensorData {
            device_id,
            temperature,
            humidity,
            led_state
        })
    }

    pub async fn toggle_warning_light(&mut self, open_windows: bool) -> Result<(), Box<dyn Error>> {
        if open_windows {
            self.get_arduino_response("0", 50).await?;
            println!("Warning light OFF");
        } else {
            self.get_arduino_response("1", 50).await?;
            println!("Warning light ON");
        }
        Ok(())
    }
}

fn is_valid_response(response: &str) -> bool {
    !response.is_empty() && (is_valid_data_format(response) || response == "a")
}

pub fn is_valid_data_format(input: &str) -> bool {
    let pattern = Regex::new(r"^\d{1,20},\d{2}\.\d{2},\d{2}\.\d{2},[01]$").unwrap();
    pattern.is_match(input)
}
