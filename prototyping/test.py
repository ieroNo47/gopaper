# Play around with the Instapaper API
# https://www.instapaper.com/api/full

import os
import sys
import requests

from dotenv import load_dotenv
from requests.auth import HTTPBasicAuth


def auth_check(s, api):
    url = f"{api}/authenticate"
    r = s.get(url)
    try:
        r.raise_for_status()
        print(f"OK: {r.status_code}")
    except requests.exceptions.HTTPError as he:
        print(f"FAILED: {he}")
        sys.exit(1)


def main():
    load_dotenv()
    username = os.environ.get("IP_USER")
    password = os.environ.get("IP_PASSWORD")
    api = os.environ.get("IP_API")
    s = requests.Session()
    s.auth = HTTPBasicAuth(username, password)
    auth_check(s, api)

    url = f"{api}/1/oauth/access_token"
    tr = s.post(url, params={
        "x_auth_username": username,
        "x_auth_password": password,
        "x_auth_mode": "client_auth"
    })
    print(tr.text)




if __name__ == "__main__":
    main()