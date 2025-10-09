package dbv

import (
	"fmt"
	"log/slog"
	"path"
	"regexp"
	"strings"
	"sync"

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

type downloader struct {
	progress *mpb.Progress
	Err      error
	vi       *videoInfo
}

func newDownloader(p *mpb.Progress) *downloader {
	return &downloader{
		progress: p,
	}
}

// 从 url 获取 bvid, 如果不是为了解析
// b23.tv 开头的链接, 不然不会写这么长.
func (d *downloader) getBvid(url string) *downloader {
	if d.Err != nil {
		return d
	}
	if strings.Contains(url, "b23.tv") {
		if url[0:8] != "https://" {
			url = fmt.Sprintf("https://%s", url)
		} else if url[0:7] != "http://" {
			url = fmt.Sprintf("https://%s", url)
		}
		data, err := SenderGetAllRaw(url)
		if err != nil {
			return d
		}
		url = string(data)
	}
	bvid := regexp.MustCompile(`BV[0-9A-Za-z]{10}`).FindString(url)
	if len(bvid) == 0 {
		d.Err = fmt.Errorf("无法从 %s 解析出 bvid", url)
		return d
	}
	d.vi = &videoInfo{
		bvid: bvid,
	}
	slog.Debug(fmt.Sprintf("解析出来的 bv 号是 %s", bvid))
	return d
}

// 获取视频信息, 但是没有视频链接.
func (d *downloader) getVideoInfoNoVideoUrl() *downloader {
	if d.Err != nil {
		return d
	}
	data, err := SenderGetAllRaw(GetVideoInfoApi(d.vi.bvid))
	if s.veryVerbose {
		if err := WriteRawDataToFile(path.Join(s.saveDir, d.vi.title+"_videoinfo.json"), data); err != nil {
			slog.Warn(err.Error())
		}
	}
	if err != nil {
		d.Err = err
		return d
	}
	dataJson := gjson.ParseBytes(data)
	if dataJson.Get("code").String() == "0" {
		d.vi.cid = dataJson.Get("data.cid").String()
		d.vi.title = RemoveIllegalCharacters(dataJson.Get("data.title").String())
		d.vi.picUrl = dataJson.Get("data.pic").String()
		slog.Debug(fmt.Sprintf("解析出来的 cid 号是 %s", d.vi.cid))
		slog.Debug(fmt.Sprintf("解析出来的标题是 %s", d.vi.title))
		slog.Debug(fmt.Sprintf("解析出来的封面下载链接是 %s", d.vi.picUrl))
	} else {
		d.Err = fmt.Errorf("无法从 %s 解析出 cid", d.vi.bvid)
	}
	return d
}

// 获取视频链接, 只能获取 720p.
func (d *downloader) getVideoUrl() *downloader {
	if d.Err != nil {
		return d
	}
	data, err := SenderGetAllRaw(GetVideoUrlApi(d.vi.bvid, d.vi.cid))
	if s.veryVerbose {
		if err := WriteRawDataToFile(path.Join(s.saveDir, d.vi.title+"_videourl.json"), data); err != nil {
			slog.Warn(err.Error())
		}
	}
	if err != nil {
		d.Err = err
		return d
	}
	dataJson := gjson.ParseBytes(data)
	if dataJson.Get("code").String() == "0" && dataJson.Get("data.durl.0.url").Exists() {
		d.vi.vedioUrl = dataJson.Get("data.durl.0.url").String()
		slog.Debug("解析出来的视频下载链接是" + d.vi.vedioUrl)
		d.vi.size = dataJson.Get("data.durl.0.size").Int()
		slog.Debug(fmt.Sprintf("解析出来的视频大小是 %.2f MB", float64(d.vi.size)/1048576))
	} else {
		d.Err = fmt.Errorf("无法从 %s 解析出视频下载链接", d.vi.bvid)
	}
	return d
}

// 下载封面.
func (d *downloader) downloadPic() *downloader {
	if !s.savePic || d.Err != nil {
		return d
	}
	_, err := WriteFileFromUrl(path.Join(s.saveDir, d.vi.title+".png"), d.vi.picUrl)
	if err != nil {
		d.Err = err
		return d
	}
	slog.Info(fmt.Sprintf("%s 封面下载完成", d.vi.title))
	return d
}

// 下载视频.
func (d *downloader) downloadVedio() *downloader {
	if s.noSaveVideo || d.Err != nil {
		return d
	}
	reader, err := WriteFileFromUrl(path.Join(s.saveDir, d.vi.title+".mp4"), d.vi.vedioUrl)
	if err != nil {
		d.Err = err
		return d
	}
	// bar := progressbar.DefaultBytes(
	// 	d.size,
	// 	"downloading",
	// )
	// _, err = io.Copy(io.MultiWriter(vedio, bar), reader)
	// if err != nil {
	// 	return err
	// }
	bar := d.progress.AddBar(
		d.vi.size,
		mpb.PrependDecorators(
			decor.Name(fmt.Sprintf("%s ", d.vi.title)),
		),
		mpb.AppendDecorators(
			decor.CountersKibiByte("% .2f / % .2f"),
			decor.Percentage(),
		),
	)
	proxy := bar.ProxyReader(reader)
	defer proxy.Close()
	bar.SetTotal(d.vi.size, true) // 标记完成
	slog.Info(fmt.Sprintf("%s 视频下载完成", d.vi.title))
	return d
}
func (d *downloader) run(wg *sync.WaitGroup) {
	for s.urls.Len() != 0 {
		u := s.urls.PopFront()
		d.getBvid(u).getVideoInfoNoVideoUrl().getVideoUrl().downloadPic().downloadVedio()
		if d.Err != nil {
			slog.Warn(d.Err.Error())
			s.fail.PushBack(u)
			d.Err = nil
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
