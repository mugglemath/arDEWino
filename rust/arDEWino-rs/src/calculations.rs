pub fn calculate_dewpoint(t: f64, rh: f64) -> f64 {
    if rh < 0.0 || rh > 100.0 {
        panic!("Relative humidity must be between 0 and 100.");
    }
    let rh_decimal = rh / 100.0;
    let log_rh = rh_decimal.ln();
    let numerator = 243.04 * (log_rh + ((17.625 * t) / (243.04 + t)));
    let denominator = 17.625 - log_rh - ((17.625 * t) / (243.04 + t));
    numerator / denominator
}

pub fn round_to_2_decimal_places(value: f64) -> f64 {
    (value * 100.0).round() / 100.0
}
