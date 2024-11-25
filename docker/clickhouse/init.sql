CREATE DATABASE IF NOT EXISTS my_database;
USE my_database;

CREATE TABLE indoor_environment (
    device_id UInt16,
    indoor_temperature Float64,
    indoor_humidity Float64,
    indoor_dewpoint Float64,
    outdoor_dewpoint Float64,
    dewpoint_delta Float64,
    keep_windows String,
    humidity_alert UInt8,
    isoTimestamp DateTime DEFAULT now()
) ENGINE = MergeTree()
ORDER BY (isoTimestamp, device_id);