import requests, json
from bs4 import BeautifulSoup

client_id = "C1pU2xkeCVWOwE2HL6qDLUarH2e61RWw"

headers = {
    "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36"
}


def get_track_authorization():
    soup = BeautifulSoup(
        requests.get("https://soundcloud.com/riotvirtual/now-we-got-problems").content,
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
    return (
        js[-1]["data"]["track_authorization"],
        js[-1]["data"]["media"]["transcodings"][0]["url"],
    )


def get_m3u_url(req_url):
    url = req_url + f"?client_id={client_id}&track_authorization={track_authorization}"
    print(url)
    m3u_resp = requests.get(url, headers=headers)
    print(m3u_resp.text)
    return m3u_resp.json()["url"]


def download_file(m3u_url):
    print(m3u_url)
    raw = requests.get(m3u_url).text
    with open("music.mp3", "wb") as f:
        for url in [line for line in raw.splitlines() if not "#" in line]:
            f.write(requests.get(url).content)


track_authorization, req_url = get_track_authorization()
m3u_url = get_m3u_url(req_url)

download_file(m3u_url)
