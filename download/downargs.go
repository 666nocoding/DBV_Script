package download

import (
	"bufio"
	"log/slog"
	"os"
	"strings"

	"github.com/gammazero/deque"
)

var Args DownArgs

type DownArgs struct {
	saveDir     string
	savePic     bool
	noSaveVideo bool
	verbose     bool
	veryVerbose bool
	urls        *deque.Deque[string]
}

func (args *DownArgs) Init() {
	args.urls = new(deque.Deque[string])
	args.saveDir = "."
	args.verbose = false
}
func (args *DownArgs) LoadUrlFile(urlfile string) error {
	file, err := os.Open(urlfile)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line[:1] != "#" {
			args.urls.PushBack(line)
		}
	}
	return nil
}
func (args *DownArgs) SetSaveDir(saveDir string) {
	args.saveDir = saveDir
}
func (args *DownArgs) GetSaveDir() string {
	return args.saveDir
}
func (args *DownArgs) SetSavePic(savePic bool) {
	args.savePic = savePic
}
func (args *DownArgs) GetSavePic() bool {
	return args.savePic
}
func (args *DownArgs) SetNoSaveVideo(noSaveVideo bool) {
	args.noSaveVideo = noSaveVideo
}
func (args *DownArgs) GetNoSaveVideo() bool {
	return args.noSaveVideo
}
func (args *DownArgs) SetVerbose(verbose bool) {
	args.verbose = verbose
}
func (args *DownArgs) GetVerbose() bool {
	return args.verbose
}
func (args *DownArgs) SetVeryVerbose(veryVerbose bool) {
	args.veryVerbose = veryVerbose
}
func (args *DownArgs) GetVeryVerbose() bool {
	return args.veryVerbose
}
func (args *DownArgs) PushUrlBack(url string) {
	args.urls.PushBack(url)
}
func (args *DownArgs) GetUrlsFront() string {
	if args.urls.Len() > 0 {
		return args.urls.Front()
	}
	return ""
}
func (args *DownArgs) PopUrlsFront() {
	if args.urls.Len() > 0 {
		args.urls.PopFront()
	}
}
func (args *DownArgs) GetUrlsNum() int {
	return args.urls.Len()
}
func (args *DownArgs) IsUrlsEmpty() bool {
	return args.urls.Len() == 0
}

type VideoInfo struct {
	bvid     string
	cid      string
	title    string
	picUrl   string
	vedioUrl string
	size     int64
}

func (videoInfo *VideoInfo) SetTitle(title string) {
	illegalChars := `\/:*?"<>|`
	if strings.ContainsAny(title, illegalChars) {
		slog.Info("标题出现特殊字符，现将标题的特殊字符进行剔除，否则无法保存")
		videoInfo.title = strings.NewReplacer(`/`, "", `\`, "", `:`, "", `*`, "", `?`, "", `"`, "", `<`, "", `>`, "", `|`, "").
			Replace(title)
	} else {
		videoInfo.title = title
	}
}
func (videoInfo *VideoInfo) GetTitle() string {
	return videoInfo.title
}
func (videoInfo *VideoInfo) SetBvid(bvid string) {
	videoInfo.bvid = bvid
}
func (videoInfo *VideoInfo) GetBvid() string {
	return videoInfo.bvid
}
func (videoInfo *VideoInfo) SetCid(cid string) {
	videoInfo.cid = cid
}
func (videoInfo *VideoInfo) GetCid() string {
	return videoInfo.cid
}
func (videoInfo *VideoInfo) SetPicUrl(picUrl string) {
	videoInfo.picUrl = picUrl
}
func (videoInfo *VideoInfo) GetPicUrl() string {
	return videoInfo.picUrl
}
func (videoInfo *VideoInfo) SetVedioUrl(vedioUrl string) {
	videoInfo.vedioUrl = vedioUrl
}
func (videoInfo *VideoInfo) GetVedioUrl() string {
	return videoInfo.vedioUrl
}
func (videoInfo *VideoInfo) SetSize(size int64) {
	videoInfo.size = size
}
func (videoInfo *VideoInfo) GetSize() int64 {
	return videoInfo.size
}
