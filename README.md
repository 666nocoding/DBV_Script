# download bilibili vedio script

下载哔哩哔哩视频脚本。用来学习go的。

## 使用方式

```shell
$ dbv -h
介绍:
    下载哔哩哔哩视频的命令行工具（菜鸟写的小玩具）
使用方式:
    dbv.exe [选项] <链接>
当前版本:
    dbv-windows-x86-64-v0.2.1
选项:
    -f, --file             从指定的文件解析 BV 号，一行一个链接，行头是 # 时不解析
    -s, --saveDir          设置视频封面保存目录（默认程序运行的目录）
    -V, --verbose          是否详细输出（默认否）
    -VV, --Verbose         是否非常详细输出（默认否）
    -sp, --savePic         是否保存视频封面（默认不保存）
    -nsv, --nosaveVideo    是否保存视频（默认保存）
    -b, --bar              是否打开下载进度条（默认打开）
    -h, --help             显示帮助信息后退出
```

## 构建方式

```shell
git clone --depth 1 http://docker.mydns.com:54671/Winter/DBV_Script.git
cd DBV_Script.git
go mod tidy
# go env -w GOOS=linux
go build -o ./dbv-linux-x86-64-v0.3.0.o main.go
# go env -w GOOS=windows
# go build -o ./dbv-windows-x86-64-v0.3.0.exe main.go
```
