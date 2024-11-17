use reqwest::{self};
use serde_json::json;
use std::error::Error;

use crate::calculations::round_to_2_decimal_places;
use crate::http_requests;
use crate::models::IndoorSensorData;

pub async fn get_request(url: &str) -> Result<String, Box<dyn std::error::Error>> {
    let body = reqwest::get(url).await?.text().await?;
    Ok(body)
}

pub async fn post_request(
    url: &str,
    data: serde_json::Value,
) -> Result<String, Box<dyn std::error::Error>> {
    let client = reqwest::Client::new();
    let res = client.post(url).json(&data).send().await?;
    println!("Status: {}", res.status());
    let body = res.text().await?;
    println!("POST response body: {}", body);
    Ok(body)
}

pub async fn get_outdoor_dewpoint() -> Result<f64, Box<dyn Error>> {
    let get_url = std::env::var("GET_URL")?;
    let get_response = crate::http_requests::get_request(&get_url).await?;
    println!("GET outdoor_dewpoint response: {}", get_response.trim());
    get_response
        .parse::<f64>()
        .map_err(|_| "Invalid float format".into())
}

pub fn prepare_sensor_feed_json(
    indoor_data: &IndoorSensorData,
    device_id: String,
    indoor_dewpoint: f64,
    outdoor_dewpoint: f64,
    dewpoint_delta: f64,
    keep_windows: bool,
    humidity_alert: bool,
) -> String {
    json!({
        "device_id": device_id,
        "indoor_temperature": round_to_2_decimal_places(indoor_data.temperature),
        "indoor_humidity": round_to_2_decimal_places(indoor_data.humidity),
        "indoor_dewpoint": round_to_2_decimal_places(indoor_dewpoint),
        "outdoor_dewpoint": round_to_2_decimal_places(outdoor_dewpoint),
        "dewpoint_delta": round_to_2_decimal_places(dewpoint_delta),
        "keep_windows": if keep_windows { "Open" } else { "Closed" },
        "humidity_alert": humidity_alert,
    })
    .to_string()
}

pub async fn post_sensor_feed(json_string: &str) -> Result<(), Box<dyn Error>> {
    let sensor_feed_url = std::env::var("POST_URL_SENSOR_FEED")?;
    let _ =
        http_requests::post_request(&sensor_feed_url, serde_json::from_str(json_string)?).await?;
    Ok(())
}
