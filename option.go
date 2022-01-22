package gis

import (
	"fmt"
	"log"
	"net/http"
	url2 "net/url"
	"os"
	"runtime"
	"strings"
)

type Option interface {
	apply(s *Searcher)
}

type function func(s *Searcher)

func (f function) apply(s *Searcher) {
	f(s)
}

func WithClient(c *http.Client) Option {
	return function(func(s *Searcher) {
		s.client = c
	})
}

func WithMaxRetryTimes(times int) Option {
	return function(func(s *Searcher) {
		if times < 0 {
			times = 0
		}
		s.maxRetryTimes = times
	})
}

// 必须是上传文件的请求路径
func WithURL(url string) Option {
	return function(func(s *Searcher) {
		u, err := url2.Parse(url)
		if err != nil {
			fmt.Println("URL is invalid: ", err)
			os.Exit(-1)
		}
		s.url = u
	})
}

func WithLogger(l *log.Logger) Option {
	return function(func(s *Searcher) {
		s.log = l
	})
}

func WithUserAgents(agents []string) Option {
	return function(func(s *Searcher) {
		if len(agents) == 0 {
			fmt.Println("User-Agent must have one element")
			os.Exit(-1)
		}
		s.userAgents = agents
	})
}

func WithUpload(upload string) Option {
	return function(func(s *Searcher) {
		if !s.exist(upload) {
			fmt.Println("no such path: ", upload)
			os.Exit(-1)
		}
		if runtime.GOOS == "windows" {
			upload = strings.ReplaceAll(upload, "/", string(os.PathSeparator))
		}
		s.upload = upload
	})
}

func WithDownload(download string) Option {
	return function(func(s *Searcher) {
		s.download = download
	})
}
