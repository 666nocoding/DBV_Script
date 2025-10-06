package dbv

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/urfave/cli/v3"
)

const version string = "dbv-v0.3.1"

const helpMessage string = `
{{- "介绍:" }}
    {{ .Description }}
{{ "使用方式:" }}
    {{ .Name }} [选项] <链接>
{{ "当前版本:" }}
    {{ .Version }}
{{ "选项:" }}
    {{ range .Flags }}-{{ range .Aliases }}{{ . }}{{ end }}, --{{ .Name }}{{ "\t\t" }}{{ .Usage }}{{ "\n    " }}{{ end }}
{{ "作者:" }}
    {{ range .Authors }}{{ . }}{{ end }}
`

var parser cli.Command = cli.Command{
	Name:        "dbv.exe",
	Description: "下载哔哩哔哩视频的命令行工具",
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
			Usage:   "设置视频和封面保存目录（默认程序运行的目录）",
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
		&cli.IntFlag{
			Name:    "maxgor",
			Aliases: []string{"m"},
			Value:   3,
			Usage:   "最大并发数，默认是 3",
		},
	},
	Action:      getArgs,
	Version:     version,
	HideVersion: true,
	Authors:     []any{"666nocoding"},
}

type settings struct {
	saveDir     string
	savePic     bool
	noSaveVideo bool
	verbose     bool
	veryVerbose bool
	maxgor      int
	urls        *safeDeque[string]
	fail        *safeDeque[string]
}

var s settings = settings{
	urls: NewSafeDeque[string](),
	fail: NewSafeDeque[string](),
}

func getArgs(ctx context.Context, cmd *cli.Command) error {
	s.saveDir = cmd.String("saveDir")
	s.verbose = cmd.Bool("verbose")
	s.veryVerbose = cmd.Bool("Verbose")
	s.savePic = cmd.Bool("savePic")
	s.noSaveVideo = cmd.Bool("nosaveVideo")
	s.maxgor = cmd.Int("maxgor")
	for i := range cmd.Args().Len() {
		s.urls.PushBack(cmd.Args().Get(i))
	}
	if cmd.String("file") != "" {
		if err := LoadUrlFile(cmd.String("file"), s.urls); err != nil {
			return errors.New("无法打开 url 文件，请检查文件是否存在或者权限")
		}
	}
	if _, err := os.Stat(s.saveDir); os.IsNotExist(err) {
		return errors.New("无法进入保存目录，请保存目录是否存在或者权限")
	}
	if s.urls.Len() <= 0 {
		return errors.New("必须至少提供一个 url")
	}
	if s.maxgor <= 0 {
		return errors.New("协程必须大于 0")
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
	ctx := context.Background()
	err := parser.Run(ctx, os.Args)
	if err == nil {
		if s.veryVerbose {
			slog.SetLogLoggerLevel(slog.LevelDebug)
		} else if s.verbose {
			slog.SetLogLoggerLevel(slog.LevelInfo)
		} else {
			slog.SetLogLoggerLevel(slog.LevelWarn)
		}
	}
	return err
}
