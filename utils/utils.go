package utils

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

var GISLog *log.Logger
var pattern *regexp.Regexp
var client http.Client
var m sync.Mutex

func init() {
	f, err := os.OpenFile("gis.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Err: ", err)
		os.Exit(1)
	}
	// 匹配js中的base64加密之后的图片内容
	pattern = regexp.MustCompile(`data:image/jpeg;base64,(.*?)';`)
	GISLog = log.New(f, "GIS", log.LstdFlags)

	rand.Seed(time.Now().UnixNano())
}

func GetRandomUserAgent() string {
	UserAgent := []string{
		`Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50`,
		`Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.102 Safari/537.36 Edg/85.0.564.51`,
		`Mozilla/5.0 (X11; U; Linux x86_64; en-us) AppleWebKit/531.2+ (KHTML, like Gecko) Version/5.0 Safari/531.2`,
		`Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36 OPR/48.0.2685.52`,
		`Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.49 Safari/537.36 OPR/48.0.2685.7`,
		`Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.83 Safari/537.36 Edg/85.0.564.44`,
		`Mozilla/5.0 (X11; U; FreeBSD i386; zh-tw; rv:31.0) Gecko/20100101 Opera/13.0`,
		`Mozilla/5.0 (X11; CrOS armv7l 9592.96.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.114 Safari/537.36`,
	}
	return UserAgent[rand.Intn(len(UserAgent))]
}

// 向网站发起POST请求，获取对应的html
func GetHtmlResource(imagePath, filename string, mirror bool) ([]byte, error) {
	fmt.Println("Uploading: ", filename)
	now := time.Now()
	var buff bytes.Buffer
	writer := multipart.NewWriter(&buff)
	writer.WriteField("image_url", "")
	writer.WriteField("image_content", "")
	writer.WriteField("filename", "")
	writer.WriteField("hl", "en")

	w, err := writer.CreateFormFile("encoded_image", filename)
	if err != nil {
		GISLog.Println("Failed to create file form: ", err)
		return nil, err
	}
	data, err := ioutil.ReadFile(imagePath)

	if err != nil {
		GISLog.Println("Failed to read the image: ", err)
		return nil, err
	}
	w.Write(data)
	writer.Close()

	var header http.Header = make(map[string][]string)
	header.Set("User-Agent", GetRandomUserAgent())

	// 默认情况下使用镜像
	url := "https://images.hk.53yu.com/searchbyimage/upload"
	header.Set("Host", "images.hk.53yu.com")

	if !mirror {
		header.Set("Host", "www.google.com")
		url = "https://www.google.com/searchbyimage/upload"
	}
	header.Set("Content-Type", writer.FormDataContentType())

	req, err := http.NewRequest("POST", url, &buff)

	req.Header = header
	if err != nil {
		GISLog.Println("Failed to create request: ", err)
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		GISLog.Println("Failed to send request: ", err)
		return nil, err
	}
	defer resp.Body.Close()

	d, err := ioutil.ReadAll(resp.Body)

	defer resp.Body.Close()
	if err != nil {
		GISLog.Println("Failed to read file")
		return nil, err
	}
	fmt.Println("\nObtain source code success, spends: ", time.Since(now))
	if len(d) == 0 {
		GISLog.Println("Response Body has no data")
		return nil, errors.New("Response Body has no data")
	}
	return d, nil
}

func GetBase64ImageData(resource []byte) ([][]byte, error) {

	data := pattern.FindAllSubmatch(resource, -1)
	var temp = make([][]byte, 0)
	for _, value := range data {
		temp = append(temp, value[1])
	}
	if len(temp) == 0 {
		GISLog.Println("No matched picture in resource")
		return nil, errors.New("No matches")
	}
	return temp, nil
}

func DecodeToImage(data []byte, filename string, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Print("\rDownloading: ", filepath.Base(filename))
	data = bytes.Replace(data, []byte("\\x3d"), []byte("="), -1)
	var binary = make([]byte, len(data))
	_, err := base64.StdEncoding.Decode(binary, data)
	if err != nil {
		GISLog.Println(filename, "Image decoding failed: ", err)
		return
	}
	err = ioutil.WriteFile(filename, binary, 0666)
	if err != nil {
		GISLog.Println(filename, "Image write failed: ", err)
		return
	}

}

func GIS(imagePath, filename, upload, download string, mirror bool, retry int, wg *sync.WaitGroup) {
	now := time.Now()
	defer wg.Done()
	// 获取源代码
	resource, err := GetHtmlResource(imagePath, filename, mirror)
	if err != nil {
		return
	}

	images, err := GetBase64ImageData(resource)
	// 那就是没有匹配上
	if err != nil {
		counter := 0
		// 如果没有请求到图片，不一定是因为不存在，
		// 而可能是请求太快，返回结果的问题
		for images == nil && counter <= retry {
			resource, err := GetHtmlResource(imagePath, filename, mirror)
			if err != nil {
				return
			}

			images, _ = GetBase64ImageData(resource)
			GISLog.Println(imagePath, " Retry: ", counter)

			counter++
		}

	}

	var localWg sync.WaitGroup
	for i, value := range images {
		dir := filepath.Dir(imagePath)
		path := strings.Replace(dir, upload, download, 1)
		path = filepath.Join(path, strings.Split(filename, ".")[0])
		err := os.MkdirAll(path, 0666)
		if err != nil {
			GISLog.Println(err)
			continue
		}
		localWg.Add(1)
		go DecodeToImage(value, filepath.Join(path, fmt.Sprintf("%d.jpeg", i)), &localWg)
	}
	localWg.Wait()
	fmt.Printf("\rProcessing %s spends: %v\n", filename, time.Since(now))
}
