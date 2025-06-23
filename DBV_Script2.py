"""
DBV mean download bilibili video.
"""

from json import loads as json_loads
from requests import get as req_get
from sys import argv
from time import time as get_time
from os import rename
from re import search
import logging

from requests.exceptions import (
    Timeout as Req_Timeout,
    ChunkedEncodingError,
    ConnectionError as Req_Connection_Error,
)


class Parser_Exception(Exception):
    """A parser error occurred."""


class Option_Need_Value(Parser_Exception):
    """The option need a value."""


class Download_Exception(Exception):
    """A download error occurred."""


class BV_Invalid(Download_Exception):
    """The BV is invalid."""


class AVID_CID_Not_Found(Download_Exception):
    """The avid and cid is not found."""


class Download_URL_Not_Found(Download_Exception):
    """The download url is not found."""


class BVStructure:

    def __init__(self):
        self.max_number: int = 100
        self.current_pointer: int = 0
        self.url_list: list[str]
        self.is_empty: bool = False
        self.file: any = None

    def set_url_list(self, url_list: list[str]):
        self.url_list = url_list

    def set_url_list_from_file(self, file_path: str):
        self.file = open(file_path, "r", encoding="utf-8")

    def set_max_min(self, max_number: int):
        self.max_number = max_number

    def empty(self) -> bool:
        if self.current_pointer >= len(self.url_list):
            if self.file:
                self.url_list = self.file.readlines(self.max_number)
                self.eliminate()
                if len(self.url_list) == 0:
                    self.file.close()
                    self.is_empty = True
                else:
                    self.current_pointer = 0
            else:
                self.is_empty = True
        return self.is_empty

    def head(self) -> str:
        return self.url_list[self.current_pointer]

    def pop(self):
        self.current_pointer += 1

    def back(self):
        self.current_pointer -= 1

    def eliminate(self):
        i, length = 0, len(self.url_list)
        while i < length:
            self.url_list[i] = self.url_list[i].replace("\n", "")
            self.url_list[i] = self.url_list[i].replace('"', "")
            i += 1
        if "" in self.url_list:
            self.url_list.remove("")


class Download:

    def __init__(self):
        self.http_header = {
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
            "Referer": "https://www.bilibili.com/",
            "Upgrade-Insecure-Requests": "1",
            "Dnt": "1",
            "Cache-Control": "max-age=0",
            "priority": "u=0, i",
            "Connection": "close",
        }
        self.save_path = "."
        self.start_name = int(get_time())

    def download(self, url: str):
        bvid = self.get_bvid(url)
        avid, cid, title = self.get_avid_and_cid_url(bvid)
        title.replace("/", "_")
        download_url = self.get_download_url(avid, cid)
        self.download_video(download_url, title)

    def get_video_message_api(self, bvid: str) -> str:
        return f"https://api.bilibili.com/x/web-interface/view?bvid={bvid}"

    def get_video_download_url_api(self, avid: str, cid: str) -> str:
        return f"https://api.bilibili.com/x/player/playurl?avid={avid}&cid={cid}&qn=80&type=mp4&platform=html5&high_quality=1"

    def get_data(self, url: str) -> bytes:
        response = req_get(url, headers=self.http_header, timeout=5)
        response.close()
        return response.content

    def get_bvid(self, url: str) -> str:
        bvid: str
        if "BV" == url[0:2]:
            bvid = url
        elif "b23.tv" in url:
            url = url if "https://" in url else f"https://{url}"
            bvid_url_data = self.get_data(url).decode("utf-8")
            if "https://www.bilibili.com/video/" not in bvid_url_data:
                raise BV_Invalid(f"Download.get_bvid: 无法从 {url} 解析出 bv 号")
            else:
                url = bvid_url_data
        if "www.bilibili.com/video/" in url or "m.bilibili.com/video/" in url:
            match = search(r"BV[0-9a-zA-Z]{10}", url)
            bvid = match.group(0)
        else:
            raise BV_Invalid(f"Download.get_bvid: 无法从 {url} 解析出 bv 号")
        logging.debug(f"Download.get_bvid: 解析出来的 bv 号是 {bvid}")
        return bvid

    def get_avid_and_cid_url(self, bvid: str) -> tuple:
        url = self.get_video_message_api(bvid)
        logging.debug(f"Download.get_avid_and_cid_url: 请求地址是 {url}")
        video_message = json_loads(self.get_data(url).decode("utf-8"))
        logging.debug("Download.get_avid_and_cid_url: 成功获取到报文")
        if video_message["code"] == 0:
            return (
                str(video_message["data"]["aid"]),
                str(video_message["data"]["cid"]),
                video_message["data"]["title"],
            )
        else:
            raise AVID_CID_Not_Found(f"{video_message}")

    def get_download_url(self, avid: str, cid: str) -> str:
        url = self.get_video_download_url_api(avid=avid, cid=cid)
        logging.debug(f"Download.get_download_url: 请求地址是 {url}")
        video_message = json_loads(self.get_data(url).decode("utf-8"))
        logging.debug("Download.get_download_url: 成功获取到报文")
        if video_message["code"] == 0:
            return video_message["data"]["durl"][0]["url"]
        else:
            raise Download_URL_Not_Found(f"Download.get_download_url: {video_message}")

    def download_video(self, url: str, title: str):
        logging.debug(f"Download.download_video: 请求地址是 {url}")
        response = req_get(url=url, headers=self.http_header, stream=True, timeout=5)
        write_time, video_size = 0, int(response.headers["content-length"])
        logging.debug(f"Download.download_video: 视频大小 {video_size/1048576} MB")
        self.start_name += 1
        with open(rf"{self.save_path}/{self.start_name}.mp4", "wb") as file:
            for chunk in response.iter_content(chunk_size=1048576):
                write_time += 1
                logging.debug(f"Download.download_video: 写入硬盘 {file.write(chunk)} B")
        response.close()
        rename(
            rf"{self.save_path}/{self.start_name}.mp4", rf"{self.save_path}/{title}.mp4"
        )
        logging.debug("Download.download_video: 下载完成.")
        logging.debug(f"Download.download_video: 写入硬盘 {write_time} 次")
        logging.debug(
            f"Download.download_video: 视频保存到 {self.save_path}/{title}.mp4."
        )

    def set_save_path(self, save_path: str):
        self.save_path = save_path

    def start_name_back(self):
        self.start_name -= 1


class Parameter_Parser:

    def __init__(self):
        self.options: set[str] = {"-f", "-o"}
        self.only_options: set[str] = {"-v", "-vv"}

    def parser(self, args: list[str], bvs: BVStructure, dow: Download):
        i, length, bvlist = 0, len(args), []
        while i < length:
            if args[i] in self.options:
                if i + 1 < len(args):
                    match args[i]:
                        case "-f":
                            bvs.set_url_list_from_file(args[i + 1])
                        case "-o":
                            dow.set_save_path(args[i + 1])
                    i += 1
                else:
                    raise Option_Need_Value("Parameter_Parser.parser: 该选项需要一个参数")
            elif args[i] in self.only_options:
                match args[i]:
                    case "-v":
                        logging.getLogger().setLevel(level=logging.INFO)
                    case "-vv":
                        logging.getLogger().setLevel(level=logging.DEBUG)
            else:
                bvlist.append(args[i])
            i += 1
        bvs.set_url_list(bvlist)


class Deal_Exception:

    def __init__(self, bvs: BVStructure, dow: Download):
        self.allow_timeout_times: int = 3
        self.cannot_download_bv_list: list[str] = []
        self.bvs, self.dow = bvs, dow

    def set_bvs_dow(self, bvs: BVStructure, dow: Download):
        self.bvs, self.dow = bvs, dow

    def deal_bv_invalid(self, error: any, error_message):
        logging.warning(error)
        logging.warning(f"{error_message}")
        self.cannot_download_bv_list.append(self.bvs.head())

    def deal_again_download(self, error: any):
        if self.allow_timeout_times <= 0:
            logging.error(error)
            logging.error("连接超时或者中断超过 3 次, 跳过该 BV 号")
            self.cannot_download_bv_list.append(self.bvs.head())
            self.allow_timeout_times = 3
        else:
            self.allow_timeout_times -= 1
            self.dow.start_name_back()
            logging.warning(error)
            logging.warning(
                "连接超时或者中断, 重新请求一次. 如果超过 3 次, 则跳过该 BV 号"
            )

    def output_invalid_bv(self):
        if len(self.cannot_download_bv_list) > 0:
            for bv in self.cannot_download_bv_list:
                logging.warning(f"该链接或 BV 无法下载: {bv}")


def main():
    if len(argv) > 1:
        format = "[%(asctime)s] [%(levelname)s] [%(message)s]"
        logging.basicConfig(level=logging.WARN, format=format)
        pp, bvs, dow = Parameter_Parser(), BVStructure(), Download()
        de = Deal_Exception(bvs, dow)
        pp.parser(argv[1:], bvs, dow)
        print("Download start...")
        while not bvs.empty():
            try:
                logging.info(f"正在下载 {bvs.head()}")
                logging.debug(bvs.head())
                # dow.download(bvs.head())
                de.allow_timeout_times = 3
                logging.info(f"{bvs.head()} 下载完成")
            except BV_Invalid as error:
                de.deal_bv_invalid(error, "该 BV 号无效")
            except AVID_CID_Not_Found as error:
                de.deal_bv_invalid(
                    error, "无法获取 avid 和 cid, 该视频可能已下架或者不存在"
                )
            except Download_URL_Not_Found as error:
                de.deal_bv_invalid(error, "无法获取下载链接, 该视频可能是充电视频")
            except ChunkedEncodingError as error:
                de.deal_bv_invalid(error, "下载失败, 未知原因")
            except Req_Timeout or ConnectionError as error:
                de.deal_again_download(error, bvs, dow)
            except Parser_Exception as error:
                logging.error(error)
                break
            except Req_Connection_Error as error:
                logging.error(error)
                logging.error("无法连接互联网")
                break
            except FileNotFoundError as error:
                logging.error(error)
                logging.error("文件不存在或者输出目录不存在")
                break
            finally:
                logging.info(f"download {bvs.head()} end")
                bvs.pop()
        print("All download end")
        de.output_invalid_bv()
    else:
        print("DBV_Script2 version: 2.1")
        print("Usage: python DBV_Script2.py [bvid] [options] [args]")
        print("Options:")
        print("\t-f [file_path] 从文件中读取 BV 号")
        print("\t-o [save_path] 设置视频保存路径")
        print("\t-v             简单输出")
        print("\t-vv            详细输出")


if __name__ == "__main__":
    main()
