package dbv

import (
	"fmt"
	"log/slog"
	"path"
	"regexp"
	"strings"
	"sync"

	"github.com/tidwall/gjson"
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
	Err error
	se  *settings
	vi  *videoInfo
}

func Newdownloader(se *settings) *downloader {
	return &downloader{
		se: se,
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
	saveTitle := d.vi.bvid
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
		saveTitle = d.vi.title
	} else {
		d.Err = fmt.Errorf("无法从 %s 解析出 cid", d.vi.bvid)
	}
	if d.se.veryVerbose {
		if err := WriteRawDataToFile(path.Join(d.se.saveDir, saveTitle+"_videoinfo.json"), data); err != nil {
			slog.Warn(err.Error())
		}
	}
	return d
}

// 获取视频链接, 只能获取 720p.
func (d *downloader) getVideoUrl() *downloader {
	if d.Err != nil {
		return d
	}
	data, err := SenderGetAllRaw(GetVideoUrlApi(d.vi.bvid, d.vi.cid))
	if d.se.veryVerbose {
		if err := WriteRawDataToFile(path.Join(d.se.saveDir, d.vi.title+"_videourl.json"), data); err != nil {
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
	if !d.se.savePic || d.Err != nil {
		return d
	}
	err := WriteFileFromUrl(path.Join(d.se.saveDir, d.vi.title+".png"), d.vi.picUrl)
	if err != nil {
		d.Err = err
		return d
	}
	slog.Info(fmt.Sprintf("%s 封面下载完成", d.vi.title))
	return d
}

// 下载视频.
func (d *downloader) downloadVedio() *downloader {
	if d.se.noSaveVideo || d.Err != nil {
		return d
	}
	if !d.se.nobar {
		reader, err := SenderGetReader(d.vi.vedioUrl)
		if err != nil {
			d.Err = err
			return d
		}
		defer reader.Close()
		err = AddBarToWriteDataToFileFromIO(path.Join(d.se.saveDir, d.vi.title+".mp4"), reader, d.vi.size, d.vi.title)
		if err != nil {
			d.Err = err
			return d
		}
	} else {
		err := WriteFileFromUrl(path.Join(d.se.saveDir, d.vi.title+".mp4"), d.vi.vedioUrl)
		if err != nil {
			d.Err = err
			return d
		}
	}
	slog.Info(fmt.Sprintf("%s 视频下载完成", d.vi.title))
	return d
}
func (d *downloader) run() {
	for d.se.urls.Len() != 0 {
		u := d.se.urls.PopFront()
		d.getBvid(u).getVideoInfoNoVideoUrl().getVideoUrl().downloadPic().downloadVedio()
		if d.Err != nil {
			slog.Warn(d.Err.Error())
			d.se.fail.PushBack(u)
			d.Err = nil
		}
	}
}
func Run() {
	if globalSettings.urls.Len() == 0 {
		return
	}
	wg := sync.WaitGroup{}
	wg.Add(globalSettings.maxgor)
	for range globalSettings.maxgor {
		go func() {
			Newdownloader(globalSettings).run()
			wg.Done()
		}()
	}
	wg.Wait()
	WaitBarFinish()
	slog.Info("全部下载完成")
	if globalSettings.fail.Len() != 0 {
		slog.Warn("以下是无法下载的链接")
		for globalSettings.fail.Len() != 0 {
			fmt.Println(globalSettings.fail.PopFront())
		}
	}
}
