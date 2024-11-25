use std::error::Error;
use std::time::Instant;

use clap::{Arg, Command};
use dotenv::dotenv;

mod calculations;
mod http_requests;
mod models;
mod usb;
mod wifi;

use crate::models::IndoorSensorData;
use calculations::calculate_dewpoint;
use usb::UsbCommunication;

// TODO: refactor for concurrency
// TODO: consider removing USB support
#[tokio::main]
async fn main() -> Result<(), Box<dyn Error>> {
    let start_time_program = Instant::now();
    dotenv().ok();

    let matches = Command::new("Sensor Program")
        .about("Runs in USB or Wi-Fi mode")
        .arg(
            Arg::new("mode")
                .help("Sets the mode of operation (usb or wifi)")
                .required(true)
                .index(1),
        )
        .get_matches();
    let mode = matches.get_one::<String>("mode").unwrap();
    println!("Running in {} mode", mode);

    // establish serial communication if mode == "usb"
    let mut usb_comm = if mode == "usb" {
        let port = std::env::var("ARDUINO_PORT")?;
        Some(UsbCommunication::new(&port)?)
    } else {
        None
    };

    // get indoor sensor data depending on "usb" or "wifi" mode
    let indoor_data: IndoorSensorData = match mode.as_str() {
        "usb" => usb::UsbCommunication::get_indoor_sensor_data(usb_comm.as_mut().unwrap())?,
        "wifi" => {
            let arduino_ip = std::env::var("ARDUINO_IP")?;
            let arduino_data_endpoint = format!("{}/data", arduino_ip);
            wifi::fetch_indoor_data(arduino_data_endpoint.as_str()).await?
        }
        _ => {
            eprintln!("Invalid mode: {}", mode);
            std::process::exit(1);
        }
    };

    let outdoor_dewpoint = http_requests::get_outdoor_dewpoint().await?;
    let indoor_dewpoint = calculate_dewpoint(indoor_data.temperature, indoor_data.humidity);
    let dewpoint_delta = indoor_dewpoint - outdoor_dewpoint;
    let keep_windows = dewpoint_delta > -1.0;
    let humidity_alert = indoor_data.humidity > 60.0;
    let json_data_sensor_feed = http_requests::prepare_sensor_feed_json(
        &indoor_data,
        indoor_dewpoint,
        outdoor_dewpoint,
        dewpoint_delta,
        keep_windows,
        humidity_alert,
    );

    // print to stdout for log files
    println!("Indoor Temperature: {}", indoor_data.temperature);
    println!("Indoor Humidity: {}", indoor_data.humidity);
    println!("Outdoor Dewpoint: {}", outdoor_dewpoint);
    println!("Indoor Dewpoint: {}", indoor_dewpoint);
    println!("Dewpoint Delta: {}", dewpoint_delta);
    println!("Keep Windows Open: {}", keep_windows);
    println!("Humidity Alert: {}", humidity_alert);
    println!("Sensor Feed JSON Data: {}", json_data_sensor_feed);

    // post to sensor feed
    http_requests::post_sensor_feed(&json_data_sensor_feed).await?;

    // toggle Arduino warning light
    if keep_windows {
        if mode == "usb" {
            if let Some(ref mut comm) = usb_comm {
                usb::UsbCommunication::toggle_warning_light(comm, keep_windows)?;
            } else {
                eprintln!("USB communication not initialized.");
                std::process::exit(1);
            }
        } else if mode == "wifi" {
            wifi::toggle_warning_light(keep_windows).await?;
        }
    }

    let elapsed_time = start_time_program.elapsed();
    println!("Total program runtime: {:.2?}", elapsed_time);
    Ok(())
}
