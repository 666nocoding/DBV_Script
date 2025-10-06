package dbv

import (
	"bufio"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/gammazero/deque"
	"github.com/go-resty/resty/v2"
)

// Remove: \/:*?"<>|
func RemoveIllegalCharacters(title string) string {
	if strings.ContainsAny(title, `\/:*?"<>|`) {
		slog.Info("标题出现特殊字符，现将标题的特殊字符进行剔除，否则无法保存")
		return strings.
			NewReplacer(`/`, "", `\`, "", `:`, "", `*`, "", `?`, "", `"`, "", `<`, "", `>`, "", `|`, "").
			Replace(title)
	}
	return title
}
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

var client *resty.Client

func createClient() {
	if client == nil {
		client = resty.New()
		client.SetHeader("Referer", "https://www.bilibili.com/")
		client.SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36 Edg/132.0.0.0")
		client.SetHeader("Accept-Encoding", "gzip, deflate, br, zstd")
	}
}
func SenderGetAllRaw(url string) ([]byte, error) {
	createClient()
	request := client.R()
	respose, err := request.Get(url)
	if err != nil {
		return nil, err
	}
	return respose.Body(), nil
}
func SenderGetReader(url string) (io.ReadCloser, error) {
	createClient()
	request := client.R()
	respose, err := request.
		SetDoNotParseResponse(true).
		Get(url)
	if err != nil {
		return nil, err
	}
	return respose.RawBody(), nil
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
