## GIS (Google Image Searcher)
> 使用Go实现，除此之外还有[Python版本](https://git.io/Jvrbv)的，但是该实现更快

### 如何使用
```golang
func newSearcher() *Searcher{
	var client = http.Client{}
	searcher := NewSearcher(client)
	// 设置上传目录
	searcher.SetUploadPath("../upload")
	// 设置图片保存目录
	searcher.SetDownloadPath("download")
	// 设置最大重试次数
	searcher.SetMaxRetryTimes(66)
	// 配置日志
	f, _ := os.Create("gis.log")
	l := log.New(f, "", log.LstdFlags)
	l.SetOutput(f)
	searcher.SetLogger(l)

	return searcher
}

var searcher  = newSearcher()

// 运行搜索程序
searcher.Run()

```
