# download bilibili vedio script

下载哔哩哔哩视频脚本。用来学习go的。

该项目可能在未来很长的一段时间内都不会再更新了，因为作者觉得这很无聊，找时间再修复一下进度条不对齐的问题之后可能就不再更新了，然后作者会将其归档。

原来计划是再引入一个清晰度选择，登录啥的，但作者懒，没什么动力做，毕竟目前来说，这功能已经足够了。

## 使用方式

```shell
$ dbv -h
介绍:
    下载哔哩哔哩视频的命令行工具
使用方式:
    dbv.exe [选项] <链接>
当前版本:
    dbv-linux-x86-64-v0.3.0
选项:
    -f, --file             从指定的文件解析 BV 号，一行一个链接，行头是 # 时不解析
    -s, --saveDir          设置视频封面保存目录（默认程序运行的目录）
    -V, --verbose          是否详细输出（默认否）
    -VV, --Verbose         是否非常详细输出（默认否）
    -sp, --savePic         是否保存视频封面（默认不保存）
    -nsv, --nosaveVideo    是否保存视频（默认保存）
    -b, --bar              是否打开下载进度条（默认打开）
    -m, --maxgor           最大并发数，默认是 3
    -h, --help             显示帮助信息后退出
    
作者:
    666nocoding
```

## 构建方式

```shell
git clone --depth 1 http://docker.mydns.com:54671/Winter/DBV_Script.git
cd DBV_Script
go mod tidy
# go env -w GOOS=linux
go build -o ./dbv main.go
# go env -w GOOS=windows
# go build -o ./dbv main.go
```
