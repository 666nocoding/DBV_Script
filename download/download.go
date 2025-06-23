package download

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"regexp"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/schollz/progressbar/v3"
	"github.com/tidwall/gjson"
)

type downloader struct {
	url        *string
	client     *resty.Client
	httpHeader *map[string][]string
	videoInfo  *VideoInfo
}

func (down *downloader) sender(url string) ([]byte, error) {
	request := down.client.R()
	request.Header = httpHeader
	respose, err := request.Get(url)
	if err != nil {
		return nil, err
	}
	return respose.Body(), nil
}
func (down *downloader) senderGetRaw(url string) (io.ReadCloser, error) {
	request := down.client.R()
	request.Header = httpHeader
	respose, err := request.
		SetDoNotParseResponse(true).
		Get(url)
	if err != nil {
		return nil, err
	}
	return respose.RawBody(), nil
}
func (down *downloader) getVideoInfoApi() string {
	return fmt.Sprintf("https://api.bilibili.com/x/web-interface/view?bvid=%s",
		down.videoInfo.bvid,
	)
}
func (down *downloader) getVideoUrlApi() string {
	return fmt.Sprintf(
		"https://api.bilibili.com/x/player/wbi/playurl?bvid=%s&cid=%s&qn=1&platform=html5&high_quality=1",
		down.videoInfo.bvid, down.videoInfo.cid,
	)
}
func (down *downloader) getBvid() error {
	var url string = *down.url
	if strings.Contains(url, "b23.tv") {
		if url[0:8] != "https://" {
			url = fmt.Sprintf("https://%s", url)
		}
		data, err := down.sender(url)
		if err != nil {
			return err
		}
		if !strings.Contains(string(data), "www.bilibili.com/video/") {
			return fmt.Errorf("无法从 %s 解析出 bvid", *down.url)
		}
		url = string(data)
	}
	down.videoInfo.bvid = regexp.MustCompile(`BV[0-9A-Za-z]{10}`).FindString(url)
	if len(down.videoInfo.bvid) == 0 {
		return fmt.Errorf("无法从 %s 解析出 bvid", *down.url)
	}
	slog.Debug(fmt.Sprintf("解析出来的 bv 号是 %s", down.videoInfo.bvid))
	return nil
}
func (down *downloader) getVideoInfo() error {
	data, err := down.sender(down.getVideoInfoApi())
	if err != nil {
		return err
	}
	dataJson := gjson.ParseBytes(data)
	if dataJson.Get("code").String() == "0" {
		down.videoInfo.cid = dataJson.Get("data.cid").String()
		down.videoInfo.title = dataJson.Get("data.title").String()
		down.videoInfo.picUrl = dataJson.Get("data.pic").String()
		slog.Debug(fmt.Sprintf("解析出来的 cid 号是 %s", down.videoInfo.cid))
		slog.Debug(fmt.Sprintf("解析出来的标题是 %s", down.videoInfo.title))
		slog.Debug("注意：由于文件名不能出现特殊字符，所以文件名可能与标题不完全一致")
		down.videoInfo.title = strings.ReplaceAll(down.videoInfo.title, `\`, "")
		slog.Debug(fmt.Sprintf("解析出来的封面下载链接是 %s", down.videoInfo.picUrl))
	} else {
		return fmt.Errorf("无法从 %s 解析出 cid", down.videoInfo.bvid)
	}
	data, err = down.sender(down.getVideoUrlApi())
	if err != nil {
		return err
	}
	dataJson = gjson.ParseBytes(data)
	if dataJson.Get("code").String() == "0" && dataJson.Get("data.durl.0.url").Exists() {
		down.videoInfo.vedioUrl = dataJson.Get("data.durl.0.url").String()
		slog.Debug("解析出来的视频下载链接是" + down.videoInfo.vedioUrl)
		down.videoInfo.size = dataJson.Get("data.durl.0.size").Int()
		slog.Debug(fmt.Sprintf("解析出来的视频大小是 %.2f MB", float64(down.videoInfo.size)/1048576))
	} else {
		return fmt.Errorf("无法从 %s 解析出视频下载链接", down.videoInfo.bvid)
	}
	return nil
}
func (down *downloader) downloadVedio() error {
	reader, err := down.senderGetRaw(down.videoInfo.vedioUrl)
	if err != nil {
		return err
	}
	vedio, err := os.OpenFile(fmt.Sprintf("%s%s.mp4", Args.saveDir, down.videoInfo.title), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	bar := progressbar.DefaultBytes(
		down.videoInfo.size,
		"downloading",
	)
	_, err = io.Copy(io.MultiWriter(vedio, bar), reader)
	if err != nil {
		return err
	}
	return nil
}

var httpHeader map[string][]string = map[string][]string{
	"Accept-Encoding":           {"gzip, deflate, br, zstd"},
	"User-Agent":                {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36 Edg/132.0.0.0"},
	"Accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
	"Sec-Fetch-Dest":            {"document"},
	"Accept-Language":           {"zh-CN,zh;q=0.9,en-GB;q=0.8,en;q=0.7,en-US;q=0.6"},
	"Sec-Ch-Ua":                 {`"Not A(Brand";v="8", "Chromium";v="132", "Microsoft Edge";v="132"`},
	"Sec-Ch-Ua-Mobile":          {"?0"},
	"Sec-Ch-Ua-Platform":        {"Windows"},
	"Sec-Fetch-Mode":            {"navigate"},
	"Sec-Fetch-Site":            {"none"},
	"Sec-Fetch-User":            {"?1"},
	"Referer":                   {"https://www.bilibili.com/"},
	"Upgrade-Insecure-Requests": {"1"},
	"Dnt":                       {"1"},
	"Cache-Control":             {"max-age=0"},
	"priority":                  {"u=0, i"},
	"Connection":                {"close"},
}

func Download(url string) error {
	var down downloader = downloader{
		url:        &url,
		client:     resty.New(),
		httpHeader: &httpHeader,
		videoInfo:  new(VideoInfo),
	}
	var err error = nil
	checkErr := func(f func() error) {
		if err != nil {
			return
		}
		err = f()
	}
	checkErr(down.getBvid)
	checkErr(down.getVideoInfo)
	checkErr(down.downloadVedio)
	return err
}
