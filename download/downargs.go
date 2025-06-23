package download

import (
	"bufio"
	"os"

	"github.com/gammazero/deque"
)

var Args DownArgs

type DownArgs struct {
	saveDir string
	verbose bool
	urls    *deque.Deque[string]
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
func (args *DownArgs) SetVerbose(verbose bool) {
	args.verbose = verbose
}
func (args *DownArgs) GetVerbose() bool {
	return args.verbose
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
