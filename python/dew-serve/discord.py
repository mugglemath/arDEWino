import requests


def send_discord_message(message, webhook_url):
    """Send a message to a Discord channel via webhook."""
    data = {"content": message}
    response = requests.post(webhook_url, json=data)
    if response.status_code != 204:
        print(f"Failed to send message to Discord: {response.status_code}, {response.text}")
