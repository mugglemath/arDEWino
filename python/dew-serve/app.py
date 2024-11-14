import os
from flask import Flask, request, jsonify
from weather_api import nws_api_response, parse_outdoor_dewpoint
from discord import send_discord_message
from dotenv import load_dotenv
from datetime import datetime

load_dotenv()

office = os.getenv('OFFICE')
grid_x = os.getenv('GRID_X')
grid_y = os.getenv('GRID_Y')
user_agent = os.getenv('USER_AGENT')
discord_sensor_feed_webhook_url = os.getenv('DISCORD_SENSOR_FEED_WEBHOOK_URL')
discord_window_alert_webhook_url = os.getenv('DISCORD_WINDOW_ALERT_WEBHOOK_URL')
discord_humidity_alert_webhook_url = os.getenv('DISCORD_HUMIDITY_ALERT_WEBHOOK_URL')

outdoor_dewpoint = ""
iso_timestamp = ""

app = Flask(__name__)


@app.route('/weather/outdoor-dewpoint', methods=['GET'])
def handle_outdoor_dewpoint():
    global outdoor_dewpoint
    if request.method == 'GET':
        response = nws_api_response(office, grid_x, grid_y, user_agent)
        parsed = parse_outdoor_dewpoint(response)
        outdoor_dewpoint = parsed
        return jsonify(parsed), 200
    else:
        response_data = {"message": "Dew Server says ??"}
        return jsonify(response_data), 200


@app.route('/discord/sensor-feed', methods=['POST'])
def handle_sensor_data():
    global iso_timestamp
    if request.method == 'POST':
        data = request.json

        indoor_temperature = data.get('indoorTemperature')
        indoor_humidity = data.get('indoorHumidity')
        indoor_dewpoint = data.get('indoorDewpoint')
        iso_timestamp = datetime.now().isoformat()

        message = f'{iso_timestamp}\n' + \
                  f'Indoor Temperature = {indoor_temperature} C\n' + \
                  f'Indoor Humidity = {indoor_humidity} %\n' + \
                  f'Indoor Dewpoint = {indoor_dewpoint:.2f} C\n' + \
                  f'Outdoor Dewpoint = {outdoor_dewpoint:.2f} C\n'

        send_discord_message(message, discord_sensor_feed_webhook_url)
        print("Received data from C++ app:", data)
        return jsonify({"status": "success", "message": "POST request received"}), 200
    else:
        response_data = {"message": "Dew Server says ??"}
        return jsonify(response_data), 200


@app.route('/discord/window-alert', methods=['POST'])
def handle_window_alert():
    global iso_timestamp
    if request.method == 'POST':
        data = request.json

        indoor_dewpoint = data.get('indoorDewpoint')
        outdoor_dewpoint = data.get('outdoorDewpoint')
        dewpoint_delta = data.get('dewpointDelta')
        iso_timestamp = datetime.now().isoformat()

        message = f'{iso_timestamp}\n' + \
                  f'@everyone\n' + \
                  f'Indoor Dewpoint = {indoor_dewpoint} C\n' + \
                  f'Outdoor Dewpoint = {outdoor_dewpoint} %\n' + \
                  f'Dewpoint Delta = {dewpoint_delta} C\n' 

        send_discord_message(message, discord_window_alert_webhook_url)
        print("Received data from C++ app:", data)
        return jsonify({"status": "success", "message": "POST request received"}), 200
    else:
        response_data = {"message": "Dew Server says ??"}
        return jsonify(response_data), 200


@app.route('/discord/humidity-alert', methods=['POST'])
def handle_humidity_alert():
    if request.method == 'POST':
        data = request.json

        indoor_humidity = data.get('indoorHumidity')
        iso_timestamp = datetime.now().isoformat()

        message = f'{iso_timestamp}\n' + \
                  f'@everyone\n' + \
                  f'Indoor Humidity = {indoor_humidity} %\n'

        send_discord_message(message, discord_humidity_alert_webhook_url)
        print("Received data from C++ app:", data)
        return jsonify({"status": "success", "message": "POST request received"}), 200
    else:
        response_data = {"message": "Dew Server says ??"}
        return jsonify(response_data), 200


if __name__ == '__main__':
    app.run(port=5000)
