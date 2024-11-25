pub fn calculate_dewpoint(t: f32, rh: f32) -> f32 {
    if !(0.0..=100.0).contains(&rh) {
        panic!("Relative humidity must be between 0 and 100.");
    }
    let rh_decimal = rh / 100.0;
    let log_rh = rh_decimal.ln();
    let numerator = 243.04 * (log_rh + ((17.625 * t) / (243.04 + t)));
    let denominator = 17.625 - log_rh - ((17.625 * t) / (243.04 + t));
    numerator / denominator
}

pub fn round_to_2_decimal_places(value: f32) -> f32 {
    (value * 100.0).round() / 100.0
}
