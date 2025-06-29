package command

import (
	"DBV_Script/download"
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/urfave/cli/v3"
)

var helpMessage string = `
{{- "介绍:" }}
    {{ .Description }}
{{ "使用方式:" }}
    {{ .Name }} [选项]
{{ "当前版本:" }}
    {{ .Version }}
{{ "选项:" }}
    {{ range .Flags }}-{{ range .Aliases }}{{ . }}{{ end }}, --{{ .Name }}{{ "\t\t" }}{{ .Usage }}{{ "\n    " }}{{ end }}
{{ "作者:" }}
    {{ range .Authors }}{{ . }}{{ end }}
`
var parser cli.Command = cli.Command{
	Name:        "dbv.exe",
	Description: "下载哔哩哔哩视频的命令行工具（菜鸟写的小玩具）",
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
			Usage:   "设置视频封面保存目录（默认程序运行的目录）",
		},
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"V"},
			Value:   false,
			Usage:   "是否详细输出（默认否）",
		},
		&cli.BoolFlag{
			Name:    "Verbose",
			Aliases: []string{"VV"},
			Value:   false,
			Usage:   "是否非常详细输出（默认否）",
		},
		&cli.BoolFlag{
			Name:    "savePic",
			Aliases: []string{"sp"},
			Value:   false,
			Usage:   "是否保存视频封面（默认不保存）",
		},
		&cli.BoolFlag{
			Name:    "nosaveVideo",
			Aliases: []string{"nsv"},
			Value:   false,
			Usage:   "是否保存视频（默认保存）",
		},
		&cli.BoolFlag{
			Name:    "bar",
			Aliases: []string{"b"},
			Value:   true,
			Usage:   "是否打开下载进度条（默认打开）",
		},
	},
	Action:      getArgs,
	Version:     version,
	HideVersion: true,
	Authors:     []any{"666nocoding"},
}

func getArgs(ctx context.Context, cmd *cli.Command) error {
	download.Args.SetSaveDir(cmd.String("saveDir"))
	download.Args.SetVerbose(cmd.Bool("verbose"))
	download.Args.SetVeryVerbose(cmd.Bool("Verbose"))
	download.Args.SetSavePic(cmd.Bool("savePic"))
	download.Args.SetNoSaveVideo(cmd.Bool("nosaveVideo"))
	for i := range cmd.Args().Len() {
		download.Args.PushUrlBack(cmd.Args().Get(i))
	}
	if cmd.String("file") != "" {
		if err := download.Args.LoadUrlFile(cmd.String("file")); err != nil {
			return errors.New("无法打开 url 文件，请检查文件是否存在或者权限")
		}
	}
	if download.Args.GetUrlsNum() <= 0 {
		return errors.New("必须至少提供一个 url")
	}
	return nil
}

func Parser() error {
	cli.HelpFlag = &cli.BoolFlag{
		Name:    "help",
		Aliases: []string{"h"},
		Usage:   "显示帮助信息后退出",
	}
	cli.RootCommandHelpTemplate = helpMessage
	err := parser.Run(context.Background(), os.Args)
	if err == nil {
		if download.Args.GetVerbose() {
			slog.SetLogLoggerLevel(slog.LevelInfo)
		}
		if download.Args.GetVeryVerbose() {
			slog.SetLogLoggerLevel(slog.LevelDebug)
		}
	}
	return err
}
