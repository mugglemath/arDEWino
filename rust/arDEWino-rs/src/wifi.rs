use std::error::Error;
use tokio::time::sleep;
use tokio::time::Duration;

use crate::http_requests::get_request;
use crate::models::IndoorSensorData;
use crate::usb::is_valid_float_format;

pub async fn fetch_indoor_data(
    arduino_data_endpoint: &str,
) -> Result<IndoorSensorData, Box<dyn Error>> {
    let mut attempts = 0;
    let max_attempts = 3;

    loop {
        let arduino_data = get_request(arduino_data_endpoint).await?;
        println!("GET Arduino data response = {}", arduino_data.trim());

        if is_valid_float_format(arduino_data.trim()) {
            let parts: Vec<&str> = arduino_data.split(',').collect();

            if parts.len() < 2 {
                attempts += 1;
                println!(
                    "Data format is invalid. Attempt {} of {}",
                    attempts, max_attempts
                );

                if attempts >= max_attempts {
                    eprintln!(
                        "Failed to retrieve valid indoor data after {} attempts.",
                        max_attempts
                    );
                    std::process::exit(1);
                }

                sleep(Duration::from_secs(1)).await;
                continue;
            }

            let temperature: f64 = parts[0].trim().parse()?;
            let humidity: f64 = parts[1].trim().parse()?;

            return Ok(IndoorSensorData {
                temperature,
                humidity,
            });
        } else {
            attempts += 1;
            println!(
                "Data format is invalid. Attempt {} of {}",
                attempts, max_attempts
            );

            if attempts >= max_attempts {
                eprintln!(
                    "Failed to retrieve valid indoor data after {} attempts.",
                    max_attempts
                );
                std::process::exit(1);
            }

            sleep(Duration::from_secs(1)).await;
        }
    }
}

pub async fn toggle_warning_light(keep_windows: bool) -> Result<(), Box<dyn Error>> {
    let arduino_ip = std::env::var("ARDUINO_IP")?;
    let command_value = if keep_windows { "0" } else { "1" };
    let arduino_led_endpoint = format!("{}/led?state={}", arduino_ip, command_value);
    let _ =
        crate::http_requests::post_request(&arduino_led_endpoint, serde_json::json!({})).await?;
    Ok(())
}
