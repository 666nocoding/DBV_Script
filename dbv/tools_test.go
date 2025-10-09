package dbv

import "testing"

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
