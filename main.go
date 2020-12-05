package main

import (
	"flag"
	"fmt"
	"github.com/junhaideng/go-gis/utils"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	upload   string
	download string
	mirror   bool
	sleep    time.Duration
	retry    int
	wg       sync.WaitGroup
)

func init() {
	flag.StringVar(&upload, "upload", "upload", "which directory you want to upload images")
	flag.StringVar(&download, "download", "download", "which directory you want to save the images")
	flag.BoolVar(&mirror, "mirror", true, "use the mirror website or not")
	flag.DurationVar(&sleep, "sleep", 2*time.Second, "sleep to avoid high rate request")
	flag.IntVar(&retry, "retry", 66, "retry times to search the failed file")
	flag.Usage = usage

	rand.Seed(time.Now().UnixNano())
}

func usage() {
	fmt.Fprintln(os.Stderr, "GIS is developed by Edgar")
	fmt.Fprintln(os.Stderr, "which you can use for downloading similar images from Google(or its mirror website) using your own images")
	fmt.Fprintln(os.Stderr, "OPTIONS:")
	flag.PrintDefaults()
}

func main() {
	now := time.Now()
	flag.Parse()
	
	_, err := os.Stat(upload)
	if os.IsNotExist(err){
		fmt.Printf("上传目录 %s 不存在\n", upload)
		return
	}

	filepath.Walk(upload, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			wg.Add(1)
			time.Sleep(sleep)
			go utils.GIS(path, info.Name(), upload, download, mirror, retry, &wg)
		}
		return nil
	})
	wg.Wait()
	fmt.Println("Total time: ", time.Since(now))
}
