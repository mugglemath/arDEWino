use calculations::calculate_dewpoint;
use dotenv::dotenv;
use std::error::Error;
use std::time::Instant;

mod usb;
use usb::UsbCommunication;
mod calculations;
mod http_requests;
mod models;

// TODO: refactor for concurrent/parallel execution after hardware upgrade
#[tokio::main]
async fn main() -> Result<(), Box<dyn Error>> {
    let start_time_program = Instant::now();
    dotenv().ok();

    // establish serial communication
    let port = std::env::var("ARDUINO_PORT")?;
    let mut usb_comm = UsbCommunication::new(&port)?;

    // TODO: change later to grab uid from ESP32 board
    let device_id = "Arduino Nano USB".to_string();

    // initialize sensor feed variables
    let indoor_data = usb::UsbCommunication::get_indoor_sensor_data(&mut usb_comm)?;
    let outdoor_dewpoint = http_requests::get_outdoor_dewpoint().await?;
    let indoor_dewpoint = calculate_dewpoint(indoor_data.temperature, indoor_data.humidity);
    let dewpoint_delta = indoor_dewpoint - outdoor_dewpoint;
    let keep_windows = dewpoint_delta > -1.0;
    let humidity_alert = indoor_data.humidity > 60.0;
    let json_data_sensor_feed = http_requests::prepare_sensor_feed_json(
        &indoor_data,
        device_id,
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

    // handle alerts
    http_requests::handle_alerts(humidity_alert, keep_windows, &json_data_sensor_feed).await?;

    // toggle Arduino warning light
    usb::UsbCommunication::toggle_warning_light(&mut usb_comm, keep_windows)?;

    let elapsed_time = start_time_program.elapsed();
    println!("Total program runtime: {:.2?}", elapsed_time);
    Ok(())
}
