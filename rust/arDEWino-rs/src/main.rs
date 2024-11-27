use std::error::Error;
use std::time::Instant;

use clap::{Command, Arg};
use dotenv::dotenv;
use std::pin::Pin;
use std::future::Future;

mod calculations;
mod http_requests;
mod models;
mod usb;
mod wifi;

use calculations::calculate_dewpoint;

type IndoorDataFuture = Pin<Box<dyn Future<Output = Result<models::IndoorSensorData, Box<dyn Error>>> + Send>>;
type ToggleLightFuture = Pin<Box<dyn Future<Output = Result<(), Box<dyn Error>>> + Send>>;

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
        Some(usb::UsbCommunication::new(&port).await?)
    } else {
        None
    };

    // fetch outdoor dewpoint and indoor sensor data concurrently
    let outdoor_dewpoint_future = http_requests::get_outdoor_dewpoint();

    let indoor_data_future: IndoorDataFuture = match mode.as_str() {
        "usb" => {
            let mut comm = usb_comm.take().expect("USB communication not initialized.");
            Box::pin(async move {
                usb::UsbCommunication::get_indoor_sensor_data(&mut comm).await
            })
        },
        "wifi" => {
            let arduino_ip = std::env::var("ARDUINO_IP")?;
            let arduino_data_endpoint = format!("{}/data", arduino_ip);
            Box::pin(wifi::fetch_indoor_data(arduino_data_endpoint))
        },
        _ => {
            eprintln!("Invalid mode: {}", mode);
            std::process::exit(1);
        }
    };

    // await both futures concurrently
    let (indoor_data, outdoor_dewpoint) = tokio::join!(indoor_data_future, outdoor_dewpoint_future);

    // handle any errors from the futures
    let indoor_data = indoor_data?;
    let outdoor_dewpoint = outdoor_dewpoint?;

    let led_state = indoor_data.led_state;
    let indoor_dewpoint = calculate_dewpoint(indoor_data.temperature, indoor_data.humidity);
    let dewpoint_delta = indoor_dewpoint - outdoor_dewpoint;
    let open_windows = dewpoint_delta > -1.0;
    let humidity_alert = indoor_data.humidity > 60.0;

    let json_data_sensor_feed = http_requests::prepare_sensor_feed_json(
        &indoor_data,
        indoor_dewpoint,
        outdoor_dewpoint,
        dewpoint_delta,
        open_windows,
        humidity_alert,
    );

    // print to stdout for log files
    println!("Indoor Temperature: {}", indoor_data.temperature);
    println!("Indoor Humidity: {}", indoor_data.humidity);
    println!("Outdoor Dewpoint: {}", outdoor_dewpoint);
    println!("Indoor Dewpoint: {}", indoor_dewpoint);
    println!("Dewpoint Delta: {}", dewpoint_delta);
    println!("Open Windows: {}", open_windows);
    println!("Humidity Alert: {}", humidity_alert);
    println!("Sensor Feed JSON Data: {}", json_data_sensor_feed);

    // post to sensor feed and toggle warning light concurrently
    let post_feed_future = http_requests::post_sensor_feed(&json_data_sensor_feed);

    // toggle light future
    let toggle_light_future: ToggleLightFuture = if open_windows == led_state {
        if mode == "usb" {
            let mut comm = usb_comm.take().expect("USB communication not initialized.");
            Box::pin(async move {
                usb::UsbCommunication::toggle_warning_light(&mut comm, open_windows).await
            })
        } else if mode == "wifi" {
            Box::pin(wifi::toggle_warning_light(open_windows))
        } else {
            Box::pin(async { Ok(()) })
        }
    } else {
        Box::pin(async { Ok(()) })
    };

    // await both futures concurrently and handle their results
    let (post_result, toggle_result) = tokio::join!(post_feed_future, toggle_light_future);

    post_result?;
    toggle_result?;

   let elapsed_time = start_time_program.elapsed();
   println!("Total program runtime: {:.2?}", elapsed_time);

   Ok(())
}
