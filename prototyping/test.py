# Play around with the Instapaper API
# https://www.instapaper.com/api/full

import os
import sys
import requests

from dotenv import load_dotenv
from requests.auth import HTTPBasicAuth
from requests_oauthlib import OAuth1, OAuth1Session
from urllib.parse import parse_qs


def basic_auth_check():
    username = os.environ.get("IP_USER")
    password = os.environ.get("IP_PASSWORD")
    api = os.environ.get("IP_API")
    url = f"{api}/authenticate"
    r = requests.get(url, auth=HTTPBasicAuth(username, password))
    try:
        r.raise_for_status()
        print(f"BASIC AUTH: OK: {r.status_code}")
    except requests.exceptions.HTTPError as he:
        print(f"FAILED: {he}")
        sys.exit(1)


def oauth_check():
    consumer_key = os.environ.get("IP_OAUTH_CONSUMER_ID")
    consumer_secret = os.environ.get("IP_OAUTH_CONSUMER_SECRET")
    username = os.environ.get("IP_USER")
    password = os.environ.get("IP_PASSWORD")
    api = os.environ.get("IP_API")
    oauth_consumer = OAuth1(consumer_key, client_secret=consumer_secret)
    url = f"{api}/1.1/oauth/access_token"
    params = {
        "x_auth_username": username,
        "x_auth_password": password,
        "x_auth_mode": "client_auth"
    }
    r = requests.post(url, auth=oauth_consumer, params=params)
    try:
        r.raise_for_status()
        print(f"ACCESS TOKEN: OK: {r.status_code}")
    except requests.exceptions.HTTPError as he:
        print(f"FAILED: {he}")
        sys.exit(1)
    token_data = parse_qs(r.text)
    oauth_token_secret = token_data.get("oauth_token_secret")[0]
    oauth_token = token_data.get("oauth_token")[0]
    oauth_owner = OAuth1(
        consumer_key,
        client_secret=consumer_secret,
        resource_owner_key=oauth_token,
        resource_owner_secret=oauth_token_secret
    )
    url = f"{api}/1.1/account/verify_credentials"
    r = requests.post(url, auth=oauth_owner)
    try:
        r.raise_for_status()
        print(f"OAUTH: OK: {r.status_code}")
        print(r.json())
    except requests.exceptions.HTTPError as he:
        print(f"FAILED: {he}")
        sys.exit(1)


def main():
    load_dotenv()
    basic_auth_check()
    oauth_check()


if __name__ == "__main__":
    main()