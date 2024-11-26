use std::error::Error;
use tokio::time::sleep;
use tokio::time::Duration;

use crate::http_requests::get_request;
use crate::models::IndoorSensorData;
use crate::usb::is_valid_data_format;

pub async fn fetch_indoor_data(
    arduino_data_endpoint: &str,
) -> Result<IndoorSensorData, Box<dyn Error>> {
    const MAX_ATTEMPTS: usize = 3;

    for attempt in 0..MAX_ATTEMPTS {
        let arduino_data = get_request(arduino_data_endpoint).await?;
        println!("GET Arduino data response = {}", arduino_data.trim());

        if !is_valid_data_format(arduino_data.trim()) {
            println!("Data format is invalid. Attempt {} of {}", attempt + 1, MAX_ATTEMPTS);
            sleep(Duration::from_secs(1)).await;
            continue;
        }

        let parts: Vec<&str> = arduino_data.trim().split(',').collect();
        let device_id: u64 = parts[0].trim().parse().map_err(|_| "Invalid device_id format")?;
        let temperature: f32 = parts[1].trim().parse().map_err(|_| "Invalid temperature format")?;
        let humidity: f32 = parts[2].trim().parse().map_err(|_| "Invalid humidity format")?;
        let led_state: bool = match parts[3].trim() {
            "1" => true,
            "0" => false,
            _ => return Err("Invalid LED state value".into()),
        };

        return Ok(IndoorSensorData { device_id, temperature, humidity, led_state });
    }

    eprintln!("Failed to retrieve valid indoor data after {} attempts.", MAX_ATTEMPTS);
    std::process::exit(1);
}

pub async fn toggle_warning_light(open_windows: bool) -> Result<(), Box<dyn Error>> {
    let arduino_ip = std::env::var("ARDUINO_IP")?;
    let command_value = if open_windows { "0" } else { "1" };
    let arduino_led_endpoint = format!("{}/led?state={}", arduino_ip, command_value);
    let _ =
        crate::http_requests::post_request(&arduino_led_endpoint, serde_json::json!({})).await?;
    Ok(())
}
