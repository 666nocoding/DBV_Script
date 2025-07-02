""" 
DBV mean download bilibili video.
you have to pip install zlib brotli pyzstd before run this script.
"""

from json import loads as json_loads
from requests import get as req_get
from gzip import decompress as gzip_decompress
from zlib import decompress as zlib_decompress
from brotli import decompress as br_decompress
from pyzstd import decompress as zstd_decompress
from urllib.request import Request as URequest, urlopen
from sys import argv
from time import time as get_time
from os import rename
from re import search

CST_HTTP_HEADER = {
    "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36 Edg/132.0.0.0",
    "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
    "Accept-Encoding": "gzip, deflate, br, zstd",
    "Sec-Fetch-Dest": "document",
    "Accept-Language": "zh-CN,zh;q=0.9,en-GB;q=0.8,en;q=0.7,en-US;q=0.6",
    "Sec-Ch-Ua": '"Not A(Brand";v="8", "Chromium";v="132", "Microsoft Edge";v="132"',
    "Sec-Ch-Ua-Mobile": "?0",
    "Sec-Ch-Ua-Platform": "Windows",
    "Sec-Fetch-Mode": "navigate",
    "Sec-Fetch-Site": "none",
    "Sec-Fetch-User": "?1",
    "Upgrade-Insecure-Requests": "1",
    "Dnt": "1",
    "Cache-Control": "max-age=0",
    "priority": "u=0, i",
}


class DBVException(Exception):
    source: str
    error_message: any

    def __init__(self, data: any, funtion_name: str):
        self.source = funtion_name
        self.error_message = data

    def __str__(self):
        return f"{self.source}: {self.error_message}"


def get_bvid(url: str) -> str:
    bvid: str
    if "b23.tv" in url:
        url = url if "https://" in url else f"https://{url}"
        bvid_url_data = get_data(url).decode("utf-8")
        if "https://www.bilibili.com/video/" not in bvid_url_data:
            raise DBVException(f"无法从 {url} 解析出 bv 号", "get_bvid")
        else:
            url = bvid_url_data
    if "www.bilibili.com/video/" in url:
        match = search(r"BV[0-9a-zA-Z]{10}", url)
        bvid = match.group(0)
    else:
        raise DBVException(f"无法从 {url} 解析出 bv 号", "get_bvid")
    print(f"get_bvid: 解析出来的 bv 号是 {bvid}")
    return bvid


def get_video_message_api(bvid: str) -> str:
    return f"https://api.bilibili.com/x/web-interface/view?bvid={bvid}"


def get_video_download_url_api(avid: str, cid: str) -> str:
    return f"https://api.bilibili.com/x/player/playurl?avid={avid}&cid={cid}&qn=80&type=mp4&platform=html5&high_quality=1"


def decode_response(response) -> bytes:
    data: bytes
    match response.info().get("Content-Encoding"):
        case "gzip":
            data = gzip_decompress(response.read())
        case "deflate":
            data = zlib_decompress(response.read())
        case "br":
            data = br_decompress(response.read())
        case "zstd":
            data = zstd_decompress(response.read())
        case _:
            data = response.read()
    return data


def get_data(url: str) -> bytes:
    request = URequest(url=url, headers=CST_HTTP_HEADER, method="GET")
    response = urlopen(request, timeout=5)
    return decode_response(response)


def get_avid_and_cid_url(bvid: str) -> tuple:
    url = get_video_message_api(bvid)
    print(f"get_avid_and_cid: 请求地址是 {url}")
    video_message = json_loads(get_data(url).decode("utf-8"))
    print("get_avid_and_cid: 成功获取到报文")
    if video_message["code"] == 0:
        return (
            str(video_message["data"]["aid"]),
            str(video_message["data"]["cid"]),
            video_message["data"]["title"],
        )
    else:
        raise DBVException(video_message, "get_avid_and_cid_url")


def get_download_url(avid: str, cid: str) -> str:
    url = get_video_download_url_api(avid=avid, cid=cid)
    print(f"get_download_url: 请求地址是 {url}")
    video_message = json_loads(get_data(url).decode("utf-8"))
    print("get_download_url: 成功获取到报文")
    if video_message["code"] == 0:
        return video_message["data"]["durl"][0]["url"]
    else:
        raise DBVException(video_message, "get_download_url")


def download_video(url: str, title: str, save_path):
    print(f"download_video: 请求地址是 {url}")
    request = req_get(url=url, headers=CST_HTTP_HEADER, stream=True, timeout=5)
    write_time, video_size = 0, 0
    timestamp = int(get_time())
    with open(rf"{save_path}/{timestamp}.mp4", "wb") as file:
        for chunk in request.iter_content(chunk_size=1048576):
            video_size += file.write(chunk)
            write_time += 1
            print(f"download_video: 写入硬盘 {video_size} B")
    rename(rf"{save_path}/{timestamp}.mp4", rf"{save_path}/{title}.mp4")
    print("download_video: 下载完成")
    print(f"download_video: 写入硬盘 {write_time} 次")
    print(f"download_video: 视频大小 {video_size/1048576} MB")
    print(f"download_video: 视频标题是 {title}")
    print(f"download_video: 视频保存到 {save_path}/{title}.mp4")


def mian():
    try:
        if len(argv) > 1:
            bvid = argv[1] if ("BV" == argv[1][0:2]) else get_bvid(argv[1])
            save_path = argv[2] if len(argv) > 2 else "."
            avid, cid, title = get_avid_and_cid_url(bvid)
            download_url = get_download_url(avid, cid)
            download_video(download_url, title, save_path)
        else:
            print("Usage: python DBV_Script.py [bvid] [save_path]")
    except DBVException as error:
        print(error)
    except Exception as error:
        print(error)


if __name__ == "__main__":
    mian()
