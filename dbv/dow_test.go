package dbv

import "testing"

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
