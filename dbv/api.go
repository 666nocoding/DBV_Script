package dbv

import (
	"strings"
)

func GetVideoInfoApi(bvid string) string {
	b := strings.Builder{}
	b.WriteString("https://api.bilibili.com/x/web-interface/view?bvid=")
	b.WriteString(bvid)
	return b.String()
	// "https://api.bilibili.com/x/web-interface/view?bvid=%s", bvid
}
func GetVideoUrlApi(bvid, cid string) string {
	b := strings.Builder{}
	b.WriteString("https://api.bilibili.com/x/player/wbi/playurl?bvid=")
	b.WriteString(bvid)
	b.WriteString("&cid=")
	b.WriteString(cid)
	b.WriteString("&qn=1&platform=html5&high_quality=1")
	return b.String()
	// "https://api.bilibili.com/x/player/wbi/playurl?bvid=%s&cid=%s&qn=1&platform=html5&high_quality=1", bvid, cid
}
