package main

import (
	"DBV_Script/command"
	"DBV_Script/download"
	"log/slog"
)

func main() {
	download.Args.Init()
	if err := command.Parser(); err != nil {
		slog.Error(err.Error())
	}
	for !download.Args.IsUrlsEmpty() {
		url := download.Args.GetUrlsFront()
		if err := download.Download(url); err != nil {
			slog.Warn(err.Error())
		}
		download.Args.PopUrlsFront()
	}
	slog.Info("全部下载完成")
}
