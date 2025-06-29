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
		down.videoInfo.GetBvid(),
	)
}
func (down *downloader) getVideoUrlApi() string {
	return fmt.Sprintf(
		"https://api.bilibili.com/x/player/wbi/playurl?bvid=%s&cid=%s&qn=1&platform=html5&high_quality=1",
		down.videoInfo.GetBvid(), down.videoInfo.cid,
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
	down.videoInfo.SetBvid(regexp.MustCompile(`BV[0-9A-Za-z]{10}`).FindString(url))
	if len(down.videoInfo.GetBvid()) == 0 {
		return fmt.Errorf("无法从 %s 解析出 bvid", *down.url)
	}
	slog.Debug(fmt.Sprintf("解析出来的 bv 号是 %s", down.videoInfo.GetBvid()))
	return nil
}
func (down *downloader) getVideoInfo() error {
	data, err := down.sender(down.getVideoInfoApi())
	if err != nil {
		return err
	}
	dataJson := gjson.ParseBytes(data)
	if dataJson.Get("code").String() == "0" {
		down.videoInfo.SetCid(dataJson.Get("data.cid").String())
		down.videoInfo.SetTitle(dataJson.Get("data.title").String())
		down.videoInfo.SetPicUrl(dataJson.Get("data.pic").String())
		slog.Debug(fmt.Sprintf("解析出来的 cid 号是 %s", down.videoInfo.GetCid()))
		slog.Debug(fmt.Sprintf("解析出来的标题是 %s", down.videoInfo.GetTitle()))
		slog.Debug(fmt.Sprintf("解析出来的封面下载链接是 %s", down.videoInfo.GetPicUrl()))
	} else {
		return fmt.Errorf("无法从 %s 解析出 cid", down.videoInfo.GetBvid())
	}
	data, err = down.sender(down.getVideoUrlApi())
	if err != nil {
		return err
	}
	dataJson = gjson.ParseBytes(data)
	if dataJson.Get("code").String() == "0" && dataJson.Get("data.durl.0.url").Exists() {
		down.videoInfo.SetVedioUrl(dataJson.Get("data.durl.0.url").String())
		slog.Debug("解析出来的视频下载链接是" + down.videoInfo.GetVedioUrl())
		down.videoInfo.SetSize(dataJson.Get("data.durl.0.size").Int())
		slog.Debug(fmt.Sprintf("解析出来的视频大小是 %.2f MB", float64(down.videoInfo.GetSize())/1048576))
	} else {
		return fmt.Errorf("无法从 %s 解析出视频下载链接", down.videoInfo.GetBvid())
	}
	return nil
}
func (down *downloader) downloadVedio() error {
	if !Args.GetNoSaveVideo() {
		reader, err := down.senderGetRaw(down.videoInfo.GetVedioUrl())
		if err != nil {
			return err
		}
		vedio, err := os.OpenFile(fmt.Sprintf("%s%s.mp4", Args.saveDir, down.videoInfo.GetTitle()), os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			return err
		}
		bar := progressbar.DefaultBytes(
			down.videoInfo.GetSize(),
			"downloading",
		)
		_, err = io.Copy(io.MultiWriter(vedio, bar), reader)
		if err != nil {
			return err
		}
	}
	return nil
}

func (down *downloader) downloadPic() error {
	if Args.GetSavePic() {
		reader, err := down.senderGetRaw(down.videoInfo.GetPicUrl())
		if err != nil {
			return err
		}
		pic, err := os.OpenFile(fmt.Sprintf("%s%s.png", Args.saveDir, down.videoInfo.GetTitle()), os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			return err
		}
		_, err = io.Copy(pic, reader)
		if err != nil {
			return err
		}
		return nil
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
	checkErr(down.downloadPic)
	return err
}
