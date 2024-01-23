import requests, json
from bs4 import BeautifulSoup
import urllib.parse
from dotenv import load_dotenv
import os

load_dotenv()

# Needs to be manually updated
client_id = os.getenv("CLIENT_ID")

heads = {"Cookie": os.getenv("COOKIE")}

oauth_tkn = heads["Cookie"].split("oauth_token=")[1].split(";")[0]

soup = BeautifulSoup(
    requests.get(
        "https://soundcloud.com/riotvirtual/now-we-got-problems", headers=heads
    ).content,
    "html.parser",
)
scripts = soup.find_all("script")
js = None
for script in scripts:
    try:
        js = json.loads(
            str(script)
            .replace("<script>window.__sc_hydration = ", "")
            .replace(";</script>", "")
        )
        break
    except:
        continue
song_best_quality = js[-1]["data"]["media"]["transcodings"][0]
base_url = song_best_quality["url"]
track_auth = js[-1]["data"]["track_authorization"]

final_url = (
    base_url
    + "?client_id="
    + urllib.parse.quote(client_id)
    + "&track_authorization="
    + urllib.parse.quote(track_auth)
)
print(final_url)
m3u_url = requests.get(
    final_url,
    headers={"Authorization": "OAuth " + oauth_tkn, "Cookie": heads["Cookie"]},
).json()
print(m3u_url)
m3u_url = m3u_url["url"]
raw = requests.get(m3u_url).text
# print(raw)
with open("music.wav", "wb") as f:
    init_url = [line for line in raw.splitlines() if "#EXT-X-MAP:URI" in line][0]
    init_url = init_url.replace('#EXT-X-MAP:URI="', "").replace('"', "")
    for url in [init_url] + [line for line in raw.splitlines() if not "#" in line]:
        f.write(
            requests.get(
                url,
                headers={
                    "Authorization": "OAuth " + oauth_tkn,
                    "Cookie": heads["Cookie"],
                },
            ).content
        )
