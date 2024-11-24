#!/bin/bash

# this variable determines how often arDEWino-rs runs in seconds
# defaults to 60 seconds
INTERVAL=${INTERVAL:-60}

while true; do
    ./target/release/arDEWino-rs -- wifi
    sleep $INTERVAL
done