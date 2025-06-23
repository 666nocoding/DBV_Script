package command

import (
	"DBV_Script/download"
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/urfave/cli/v3"
)

var parser cli.Command = cli.Command{
	Name:  "dbv.exe",
	Usage: "下载哔哩哔哩视频的命令行工具（菜鸟写的小玩具）",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "file",
			Aliases: []string{"f"},
			Usage:   "从指定的文件解析 BV 号，一行一个链接，行头是 # 时不解析",
		},
		&cli.StringFlag{
			Name:    "saveDir",
			Aliases: []string{"s"},
			Value:   "./",
			Usage:   "数据保存目录（默认程序运行的目录）",
		},
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"V"},
			Value:   false,
			Usage:   "是否详细输出（默认否）",
		},
		&cli.BoolFlag{
			Name:    "savePic",
			Aliases: []string{"sp"},
			Value:   true,
			Usage:   "保存视频封面（默认不保存）",
		},
		&cli.BoolFlag{
			Name:    "bar",
			Aliases: []string{"b"},
			Value:   true,
			Usage:   "打开下载进度条（默认打开）",
		},
	},
	Action:  getArgs,
	Version: "v0.2.2, written by 666nocoding",
}

func getArgs(ctx context.Context, cmd *cli.Command) error {
	download.Args.SetSaveDir(cmd.String("saveDir"))
	download.Args.SetVerbose(cmd.Bool("verbose"))
	for i := range cmd.Args().Len() {
		download.Args.PushUrlBack(cmd.Args().Get(i))
	}
	if cmd.String("file") != "" {
		if err := download.Args.LoadUrlFile(cmd.String("file")); err != nil {
			return errors.New("cant not open url file, please check whether the file exists")
		}
	}
	if download.Args.GetUrlsNum() <= 0 {
		return errors.New("at least one url")
	}
	return nil
}

func Parser() error {
	err := parser.Run(context.Background(), os.Args)
	if download.Args.GetVerbose() {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
	return err
}
