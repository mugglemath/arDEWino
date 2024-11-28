CREATE TABLE IF NOT EXISTS data (
    device_id BIGINT,
    indoor_temperature REAL,
    indoor_humidity REAL,
    indoor_dewpoint REAL,
    outdoor_dewpoint REAL,
    dewpoint_delta REAL,
    open_windows BOOLEAN DEFAULT FALSE,
    humidity_alert BOOLEAN DEFAULT FALSE,
    time TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS data_time_idx ON data (time DESC);