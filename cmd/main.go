package main

import (
	"flag"
	"log"
	"os"

	gis "github.com/junhaideng/go-gis"
)

var upload string
var download string
var url string
var logger string
var retry int

func init() {
	flag.StringVar(&upload, "u", "upload", "图片上传路径")
	flag.StringVar(&download, "d", "download", "下载的图片保存路径")
	// 目前可用镜像网址 "https://shitu.paodekuaiweixinqun.com/searchbyimage/upload"
	flag.StringVar(&url, "url", "https://www.google.com/searchbyimage/upload", "完整的图片搜索路径，应和请求体 Google 原站请求一致")
	flag.StringVar(&logger, "l", "gis.log", "日志保存文件")
	flag.IntVar(&retry, "r", 10, "重试次数，往往一次不会成功")
}

func newSearcher() *gis.Searcher {
	// 配置日志
	f, _ := os.Create(logger)
	l := log.New(f, "", log.LstdFlags)
	l.SetOutput(f)

	// 自定义配置，不配置使用默认值
	options := []gis.Option{
		gis.WithUpload(upload),
		gis.WithDownload(download),
		gis.WithLogger(l),
		gis.WithURL(url),
		gis.WithMaxRetryTimes(retry),
	}

	searcher := gis.NewSearcher(options...)
	return searcher
}

func main() {
	flag.Parse()
	var searcher = newSearcher()
	// 运行搜索程序
	searcher.Run()
}
