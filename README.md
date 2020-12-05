## GIS (Google Image Searcher)
> 使用Go实现，除此之外还有[Python版本](https://git.io/Jvrbv)的，但是该实现更快

### 如何使用
- [x] : 指定你的上传目录夹
- [x] : 执行 `go run main.go`(使用`-h`参数可以查看其他设置和帮助)

```
OPTIONS:
  -download string
        which directory you want to save the images (default "download")
  -mirror
        use the mirror website or not (default true)
  -retry int
        retry times to search the failed file (default 66)
  -sleep duration
        sleep to avoid high rate request (default 2s)
  -upload string
        which directory you want to upload images (default "upload")
```