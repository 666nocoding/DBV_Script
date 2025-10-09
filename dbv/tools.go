package dbv

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/gammazero/deque"
	"github.com/go-resty/resty/v2"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

// 剔除非法字符: \/:*?"<>|
func RemoveIllegalCharacters(title string) string {
	if strings.ContainsAny(title, `\/:*?"<>|`) {
		slog.Info("标题出现特殊字符，现将标题的特殊字符进行剔除，否则无法保存")
		return strings.
			NewReplacer(`/`, "", `\`, "", `:`, "", `*`, "", `?`, "", `"`, "", `<`, "", `>`, "", `|`, "").
			Replace(title)
	}
	return title
}

// 从文件加载 url, 每行开头是 # 不加载.
func LoadUrlFile(urlfile string, sd *safeDeque[string]) error {
	file, err := os.Open(urlfile)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line[:1] != "#" {
			sd.PushBack(line)
		}
	}
	return nil
}
func WriteRawDataToFile(path string, data []byte) (err error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	defer file.Close()
	_, err = file.Write(data)
	return
}

var progress *mpb.Progress
var onceCreateProgress sync.Once

func createProgress() {
	onceCreateProgress.Do(func() {
		if progress == nil {
			option := mpb.WithWidth(64)
			progress = mpb.New(option)
		}
	})
}
func AddBarToWriteDataToFileFromIO(path string, reader io.Reader, size int64, title string) error {
	createProgress()
	bar := progress.AddBar(
		size,
		mpb.PrependDecorators(
			decor.Name(fmt.Sprintf("%s ", title)),
		),
		mpb.AppendDecorators(
			decor.CountersKibiByte("% .2f / % .2f"),
			decor.Percentage(),
		),
	)
	proxyreader := bar.ProxyReader(reader)
	defer bar.SetTotal(size, true)
	defer proxyreader.Close()
	return WriteDataToFileFromIO(path, proxyreader)
}
func WaitBarFinish() {
	onceCreateProgress.Do(createProgress)
	progress.Wait()
}
func WriteDataToFileFromIO(path string, reader io.Reader) (err error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	defer file.Close()
	_, err = io.Copy(file, reader)
	return
}
func WriteFileFromUrl(path, url string) error {
	reader, err := SenderGetReader(url)
	if err != nil {
		return err
	}
	err = WriteDataToFileFromIO(path, reader)
	if err != nil {
		return err
	}
	return nil
}

var client *resty.Client
var onceCreateClient sync.Once

func createClient() {
	onceCreateClient.Do(func() {
		if client == nil {
			client = resty.New()
			client.SetHeader("Referer", "https://www.bilibili.com/")
			client.SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36 Edg/132.0.0.0")
			client.SetHeader("Accept-Encoding", "gzip, deflate, br, zstd")
		}
	})
}
func setClientDebug(b bool) {
	createClient()
	client.SetDebug(b)
}
func SenderGet(url string, parseResponse bool) (*resty.Response, error) {
	createClient()
	request := client.R()
	request.SetDoNotParseResponse(!parseResponse)
	return request.Get(url)
}
func SenderGetAllRaw(url string) ([]byte, error) {
	resp, err := SenderGet(url, true)
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}

// 返回的 ReadCloser 一定要关闭, 不然会内存泄漏.
func SenderGetReader(url string) (io.ReadCloser, error) {
	resp, err := SenderGet(url, false)
	if err != nil {
		return nil, err
	}
	return resp.RawBody(), nil
}

type safeDeque[T any] struct {
	d  deque.Deque[T]
	mu sync.Mutex
}

func NewSafeDeque[T any]() *safeDeque[T] {
	return &safeDeque[T]{
		d: deque.Deque[T]{},
	}
}
func (sd *safeDeque[T]) PushBack(elem T) {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	sd.d.PushBack(elem)
}
func (sd *safeDeque[T]) PopBack() T {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	return sd.d.PopBack()
}
func (sd *safeDeque[T]) PushFront(elem T) {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	sd.d.PushFront(elem)
}
func (sd *safeDeque[T]) PopFront() T {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	return sd.d.PopFront()
}
func (sd *safeDeque[T]) Len() int {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	return sd.d.Len()
}
