package dbv

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

type videoInfo struct {
	bvid     string
	cid      string
	title    string
	picUrl   string
	vedioUrl string
	size     int64
}

func setTitle(title string) string {
	if strings.ContainsAny(title, `\/:*?"<>|`) {
		slog.Info("标题出现特殊字符，现将标题的特殊字符进行剔除，否则无法保存")
		return strings.
			NewReplacer(`/`, "", `\`, "", `:`, "", `*`, "", `?`, "", `"`, "", `<`, "", `>`, "", `|`, "").
			Replace(title)
	}
	return title
}
func (d *downloader) clear() {
	d.bvid = ""
	d.cid = ""
	d.title = ""
	d.picUrl = ""
	d.vedioUrl = ""
	d.size = 0
}

type downloader struct {
	client     *resty.Client
	httpHeader map[string][]string
	videoInfo
	progress *mpb.Progress
}

func newDownloader(p *mpb.Progress) *downloader {
	return &downloader{
		client: resty.New(),
		httpHeader: map[string][]string{
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
		},
		progress: p,
	}
}
func (d *downloader) sender(url string) ([]byte, error) {
	request := d.client.R()
	request.Header = d.httpHeader
	respose, err := request.Get(url)
	if err != nil {
		return nil, err
	}
	return respose.Body(), nil
}
func (d *downloader) senderGetRaw(url string) (io.ReadCloser, error) {
	request := d.client.R()
	request.Header = d.httpHeader
	respose, err := request.
		SetDoNotParseResponse(true).
		Get(url)
	if err != nil {
		return nil, err
	}
	return respose.RawBody(), nil
}
func GetVideoInfoApi(bvid string) string {
	return fmt.Sprintf("https://api.bilibili.com/x/web-interface/view?bvid=%s", bvid)
}
func GetVideoUrlApi(bvid, cid string) string {
	return fmt.Sprintf(
		"https://api.bilibili.com/x/player/wbi/playurl?bvid=%s&cid=%s&qn=1&platform=html5&high_quality=1", bvid, cid)
}
func (d *downloader) getBvid(url string) error {
	if strings.Contains(url, "b23.tv") {
		if url[0:8] != "https://" {
			url = fmt.Sprintf("https://%s", url)
		} else if url[0:7] != "http://" {
			url = fmt.Sprintf("https://%s", url)
		}
		data, err := d.sender(url)
		if err != nil {
			return err
		}
		if !strings.Contains(string(data), "www.bilibili.com/video/") {
			return fmt.Errorf("无法从 %s 解析出 bvid", url)
		}
		url = string(data)
	}
	d.bvid = regexp.MustCompile(`BV[0-9A-Za-z]{10}`).FindString(url)
	if len(d.bvid) == 0 {
		return fmt.Errorf("无法从 %s 解析出 bvid", url)
	}
	slog.Debug(fmt.Sprintf("解析出来的 bv 号是 %s", d.bvid))
	return nil
}
func (d *downloader) getVideoInfo() error {
	data, err := d.sender(GetVideoInfoApi(d.bvid))
	if err != nil {
		return err
	}
	dataJson := gjson.ParseBytes(data)
	if dataJson.Get("code").String() == "0" {
		d.cid = dataJson.Get("data.cid").String()
		d.title = setTitle(dataJson.Get("data.title").String())
		d.picUrl = dataJson.Get("data.pic").String()
		slog.Debug(fmt.Sprintf("解析出来的 cid 号是 %s", d.cid))
		slog.Debug(fmt.Sprintf("解析出来的标题是 %s", d.title))
		slog.Debug(fmt.Sprintf("解析出来的封面下载链接是 %s", d.picUrl))
	} else {
		return fmt.Errorf("无法从 %s 解析出 cid", d.bvid)
	}
	data, err = d.sender(GetVideoUrlApi(d.bvid, d.cid))
	if err != nil {
		return err
	}
	dataJson = gjson.ParseBytes(data)
	if s.veryVerbose {
		jsonFile, err := os.OpenFile(fmt.Sprintf("%s%s.json", s.saveDir, d.title), os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			return err
		}
		_, err = jsonFile.Write(data)
		if err != nil {
			return err
		}
	}
	if dataJson.Get("code").String() == "0" && dataJson.Get("data.durl.0.url").Exists() {
		d.vedioUrl = dataJson.Get("data.durl.0.url").String()
		slog.Debug("解析出来的视频下载链接是" + d.vedioUrl)
		d.size = dataJson.Get("data.durl.0.size").Int()
		slog.Debug(fmt.Sprintf("解析出来的视频大小是 %.2f MB", float64(d.size)/1048576))
	} else {
		return fmt.Errorf("无法从 %s 解析出视频下载链接", d.bvid)
	}
	return nil
}
func (d *downloader) downloadVedio() error {
	if s.noSaveVideo {
		return nil
	}
	reader, err := d.senderGetRaw(d.vedioUrl)
	if err != nil {
		return err
	}
	vedio, err := os.OpenFile(fmt.Sprintf("%s%s.mp4", s.saveDir, d.title), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer vedio.Close()
	// bar := progressbar.DefaultBytes(
	// 	d.size,
	// 	"downloading",
	// )
	// _, err = io.Copy(io.MultiWriter(vedio, bar), reader)
	// if err != nil {
	// 	return err
	// }
	bar := d.progress.AddBar(
		d.size,
		mpb.PrependDecorators(
			decor.Name(fmt.Sprintf("%s ", d.title)),
		),
		mpb.AppendDecorators(
			decor.CountersKibiByte("% .2f / % .2f"),
			decor.Percentage(),
		),
	)

	proxy := bar.ProxyReader(reader)
	defer proxy.Close()

	_, err = io.Copy(vedio, proxy)
	if err != nil {
		return err
	}
	bar.SetTotal(d.size, true) // 标记完成
	slog.Info(fmt.Sprintf("%s 视频下载完成", d.title))
	return nil
}
func (d *downloader) downloadPic() error {
	if !s.savePic {
		return nil
	}
	reader, err := d.senderGetRaw(d.picUrl)
	if err != nil {
		return err
	}
	pic, err := os.OpenFile(fmt.Sprintf("%s%s.png", s.saveDir, d.title), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer pic.Close()
	_, err = io.Copy(pic, reader)
	if err != nil {
		return err
	}
	slog.Info(fmt.Sprintf("%s 封面下载完成", d.title))
	return nil
}
func (d *downloader) download(url string) error {
	defer d.clear()
	if err := d.getBvid(url); err != nil {
		return err
	}
	if err := d.getVideoInfo(); err != nil {
		return err
	}
	if err := d.downloadVedio(); err != nil {
		return err
	}
	if err := d.downloadPic(); err != nil {
		return err
	}
	return nil
}
func (d *downloader) run(wg *sync.WaitGroup) {
	for s.urls.Len() != 0 {
		u := s.urls.PopFront()
		if err := d.download(u); err != nil {
			slog.Warn(err.Error())
			s.fail.PushBack(u)
		}
	}
	wg.Done()
}
func Run() {
	if s.urls.Len() == 0 {
		return
	}
	p := mpb.New(mpb.WithWidth(64))
	wg := sync.WaitGroup{}
	wg.Add(s.maxgor)
	for range s.maxgor {
		go newDownloader(p).run(&wg)
	}
	wg.Wait()
	p.Wait()
	slog.Info("全部下载完成")
	if s.fail.Len() != 0 {
		slog.Warn("以下是无法下载的链接")
		for s.fail.Len() != 0 {
			fmt.Println(s.fail.PopFront())
		}
	}
}
