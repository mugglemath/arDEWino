CREATE DATABASE IF NOT EXISTS dew;
USE dew;

CREATE TABLE data (
    device_id UInt64,
    indoor_temperature Float32,
    indoor_humidity Float32,
    indoor_dewpoint Float32,
    outdoor_dewpoint Float32,
    dewpoint_delta Float32,
    open_windows UInt8,
    humidity_alert UInt8,
    time DateTime DEFAULT now()
) ENGINE = MergeTree()
ORDER BY (time, device_id);