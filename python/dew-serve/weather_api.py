import requests
from math import log


def nws_api_response(office: str, grid_x: str, grid_y: str, user_agent: str) -> str:
    try:
        r = requests.get(f'https://api.weather.gov/gridpoints/{office}/{grid_x},\
                     {grid_y}', auth=(user_agent, ''))
        if r.status_code == 200:
            return r.json()
        else:
            print(f'error fetching weather data: {r.status_code}')
    except Exception as e:
        print(f'error making weather api call: {e}')
        return None


def parse_outdoor_dewpoint(response: str) -> str:
    return response['properties']['dewpoint']['values'][0]['value']


def dewpoint_calculator(T, RH):
    return (243.04 * (log(RH / 100) + ((17.625 * T) / (243.04 + T)))) \
          / (17.625 - log(RH / 100) - ((17.625 * T) / (243.04 + T)))
