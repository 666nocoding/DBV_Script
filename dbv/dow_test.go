package dbv

import (
	"testing"
)

func TestGetBvid(t *testing.T) {
	d := downloader{}
	urls := []string{
		"https://www.bilibili.com/video/BV1nEpxz1EuF?spm_id_from=333.1007.tianma.1-1-1.click",
		"https://www.bilibili.com/video/BV1nEpxz1EuF",
		"http://www.bilibili.com/video/BV1nEpxz1EuF",
		"BV1nEpxz1EuF",
		"bilibili.com/video/BV1nEpxz1EuF",
		"www.bilibili.com/video/BV1nEpxz1EuF",
		"https://m.bilibili.com/video/BV1nEpxz1EuF",
		"BV1nEpxz1EuF123",
	}
	expect := "BV1nEpxz1EuF"
	for _, s := range urls {
		d.getBvid(s)
		if d.Err != nil {
			t.Fatal("input:", s, "expect:", expect, "but:", d.vi.bvid)
			t.Fatal(d.Err)
		}
	}
	tv_urls := []string{
		"https://b23.tv/G5cYz1g",
		"http://b23.tv/G5cYz1g",
		"b23.tv/G5cYz1g",
	}
	expect = "BV1t5yDYrEfH"
	for _, s := range tv_urls {
		d.getBvid(s)
		if d.Err != nil {
			t.Fatal("input:", s, "expect:", expect, "but:", d.vi.bvid)
			t.Fatal(d.Err)
		}
	}
	wrong_urls := []string{
		"asds",
		"bv1234567890",
	}
	for _, s := range wrong_urls {
		d.getBvid(s)
		if d.Err == nil {
			t.Fatal("input:", s, "expect: should be error", "but:", d.vi.bvid)
		}
		t.Log(d.Err)
		d.Err = nil
	}
}
func TestGetVideoInfoNoVideoUrl(t *testing.T) {
	d := downloader{
		se: NewSetting(),
	}
	d.se.veryVerbose = true
	d.se.saveDir = "../test/"
	d.vi = &videoInfo{}
	d.vi.bvid = "BV1aWJyz5Erm"
	d.getVideoInfoNoVideoUrl()
	if d.Err != nil {
		t.Fatal(d.Err)
	}
	d.vi.bvid = "BV1aWJyz5Er1"
	d.getVideoInfoNoVideoUrl()
	if d.Err == nil {
		t.Fatal("expect error, but error is nil.")
	}
	t.Log(d.Err)
}
func TestGetVideoUrl(t *testing.T) {
	d := downloader{
		se: NewSetting(),
	}
	d.se.veryVerbose = true
	d.se.saveDir = "../test/"
	d.vi = &videoInfo{}
	d.vi.bvid = "BV1aWJyz5Erm"
	d.vi.cid = "32573817186"
	d.vi.title = "搜狗输入法案中篡改用户浏览器主页事件梳理"
	d.getVideoUrl()
	if d.Err != nil {
		t.Fatal(d.Err)
	}
	t.Log(d.vi.vedioUrl)
	d.vi.bvid = "BV1aWJyz5Er1"
	d.vi.cid = "32573817186"
	d.vi.title = "搜狗输入法案中篡改用户浏览器主页事件梳理_wrong"
	d.getVideoUrl()
	if d.Err == nil {
		t.Fatal("expect error, but error is nil.")
	}
	t.Log(d.Err)
}
func TestDownloadVedio(t *testing.T) {
	d := downloader{
		se: NewSetting(),
	}
	d.se.veryVerbose = true
	d.se.saveDir = "../test/"
	d.vi = &videoInfo{}
	d.vi.vedioUrl = "https://cn-scdy-ct-01-17.bilivideo.com/upgcxcode/86/71/32573817186/32573817186-1-16.mp4?e=ig8euxZM2rNcNbRVhwdVhwdlhWdVhwdVhoNvNC8BqJIzNbfq9rVEuxTEnE8L5F6VnEsSTx0vkX8fqJeYTj_lta53NCM=&trid=00002e637d7014794e79add7ede2e775a98h&og=hw&deadline=1759998929&nbs=1&oi=236938662&gen=playurlv3&os=bcache&mid=0&uipk=5&platform=html5&upsig=ebb2653072f72eb4f44fb224f942d36d&uparams=e,trid,og,deadline,nbs,oi,gen,os,mid,uipk,platform&cdnid=88617&bvc=vod&nettype=0&bw=161059&lrs=27&agrr=1&buvid=&build=0&dl=0&f=h_0_0&orderid=0,1"
	d.vi.title = "搜狗输入法案中篡改用户浏览器主页事件梳理"
	d.downloadVedio()
	if d.Err != nil {
		t.Fatal(d.Err)
	}
}
func TestRun(t *testing.T) {
	d := Newdownloader(NewSetting())
	d.se.saveDir = "../test/"
	d.getBvid("BV1aWJyz5Erm").getVideoInfoNoVideoUrl().getVideoUrl().downloadVedio()
	if d.Err != nil {
		t.Fatal(d.Err)
	}
	d.getBvid("BV1aWJyz5Er1").getVideoInfoNoVideoUrl().getVideoUrl().downloadVedio()
	if d.Err == nil {
		t.Fatal("expect error, but error is nil.")
	}
	t.Log(d.Err)
}
