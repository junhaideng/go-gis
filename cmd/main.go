package main

import (
	gis "github.com/junhaideng/go-gis"
	"log"
	"os"
)

func newSearcher() *gis.Searcher {
	// 配置日志
	f, _ := os.Create("gis.log")
	l := log.New(f, "", log.LstdFlags)
	l.SetOutput(f)
	
  // 自定义配置，不配置使用默认值
	options := []gis.Option{
		gis.WithUpload("upload"),
		gis.WithDownload("download"),
		gis.WithLogger(l),
		gis.WithURL("https://shitu.paodekuaiweixinqun.com/searchbyimage/upload"),
		gis.WithMaxRetryTimes(10),
	}

	searcher := gis.NewSearcher(options...)
	return searcher
}

func main() {
	var searcher = newSearcher()
	// 运行搜索程序
	searcher.Run()
}
