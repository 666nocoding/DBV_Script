package dbv

import (
	"bytes"
	"os"
	"testing"
)

func TestLoadUrlFile(t *testing.T) {
	urlfile := "../test/urls.txt"
	sd := NewSafeDeque[string]()
	expect := []string{
		"https://www.bilibili.com/video/BV1yUnFzGEwP?spm_id_from=333.1007.tianma.3-2-6.click",
		"https://www.bilibili.com/video/BV1qKnAzBESz?spm_id_from=333.1007.tianma.3-1-5.click",
		"https://www.bilibili.com/video/BV1BNnAzuEiP?spm_id_from=333.1007.tianma.3-3-7.click",
		"https://www.bilibili.com/video/BV1MzpZzHEue?spm_id_from=333.1007.tianma.4-2-12.click",
	}
	if err := LoadUrlFile(urlfile, sd); err != nil {
		t.Fatal(err)
	}
	for i := range expect {
		if sd.Len() == 0 {
			t.Fatal("expect length longer than sd.")
		}
		u := sd.PopFront()
		if expect[i] != u {
			t.Fatal("expect:", expect[i], "but:", u)
		}
	}
	if sd.Len() != 0 {
		t.Fatal("sd length longer than expect.")
	}
	wrong_file := "asdsad"
	err := LoadUrlFile(wrong_file, sd)
	if err == nil {
		t.Fatal("expect error, but error is nil.")
	}
	t.Log(err)
}
func TestWriteRawDataToFile(t *testing.T) {
	bb := bytes.Buffer{}
	bb.WriteString("sadjadjskadjkasdklakdjskldjklsajadqwewqeqweqwewqe")
	path := "../test/TestWriteRawDataToFile.txt"
	err := WriteRawDataToFile(path, bb.Bytes())
	if err != nil {
		t.Fatal(err)
	}
}
func TestWriteDataToFileFromIO(t *testing.T) {
	readpath := "../test/TestWriteRawDataToFile.txt"
	writepath := "../test/TestWriteDataToFileFromIO_write"
	file, err := os.OpenFile(readpath, os.O_RDONLY, 0666)
	if err != nil {
		t.Fatal(err)
	}
	err = WriteDataToFileFromIO(writepath, file)
	if err != nil {
		t.Fatal(err)
	}
}
func TestSenderGetAllRaw(t *testing.T) {
	url := "http://172.17.0.1:8080/1.py"
	data, err := SenderGetAllRaw(url)
	if err != nil {
		t.Fatal(err)
	}
	path := "../test/TestWriteRawDataToFile.txt"
	err = WriteRawDataToFile(path, data)
	if err != nil {
		t.Fatal(err)
	}
}
func TestSenderGetReader(t *testing.T) {
	url := "http://172.17.0.1:8080/1.py"
	reader, err := SenderGetReader(url)
	if err != nil {
		t.Fatal(err)
	}
	path := "../test/TestWriteRawDataToFile.txt"
	err = WriteDataToFileFromIO(path, reader)
	if err != nil {
		t.Fatal(err)
	}
}
func TestAddBarToWriteDataToFileFromIO(t *testing.T) {
	url := "http://172.17.0.1:8080/1.py"
	data, err := SenderGetAllRaw(url)
	if err != nil {
		t.Fatal(err)
	}
	size := int64(len(data))
	reader, err := SenderGetReader(url)
	if err != nil {
		t.Fatal(err)
	}
	path := "../test/TestWriteRawDataToFile.txt"
	title := "1.py"
	err = AddBarToWriteDataToFileFromIO(path, reader, size, title)
	if err != nil {
		t.Fatal(err)
	}
	progress.Wait()
}
