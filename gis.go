package gis

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
)

var pattern = regexp.MustCompile(`data:image/jpeg;base64,(.*?)';`)
var wg sync.WaitGroup

const DEFUALT_DOWNLOAD_PATH = "download"
const DEFUALT_UPLOAD_PATH = "upload"

var DEFUALT_USER_AGENTS = []string{
	`Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50`,
	`Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.102 Safari/537.36 Edg/85.0.564.51`,
	`Mozilla/5.0 (X11; U; Linux x86_64; en-us) AppleWebKit/531.2+ (KHTML, like Gecko) Version/5.0 Safari/531.2`,
	`Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36 OPR/48.0.2685.52`,
	`Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.49 Safari/537.36 OPR/48.0.2685.7`,
	`Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.83 Safari/537.36 Edg/85.0.564.44`,
	`Mozilla/5.0 (X11; U; FreeBSD i386; zh-tw; rv:31.0) Gecko/20100101 Opera/13.0`,
	`Mozilla/5.0 (X11; CrOS armv7l 9592.96.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.114 Safari/537.36`,
}

func init() {
	// 设置随机数种子
	rand.Seed(time.Now().UnixNano())
}

type Searcher struct {
	maxRetryTimes int         // 最大尝试次数
	mirror        bool        // 是否使用镜像
	log           *log.Logger // 日志
	client        http.Client // 请求客户端
	userAgents    []string    // 请求头部中的用户代理
	upload        string      // 上传图片路径
	download      string      // 下载图片所在路径
}

func NewSearcher(client http.Client) *Searcher {
	var l = &log.Logger{}
	l.SetOutput(os.Stdout)
	return &Searcher{
		10, true, l, client, DEFUALT_USER_AGENTS, DEFUALT_UPLOAD_PATH, DEFUALT_DOWNLOAD_PATH,
	}
}

// 发送请求
func (s *Searcher) SendRequest(req *http.Request) ([]byte, error) {
	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// 读取响应
	d, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return d, nil
}

// 创建请求
func (s *Searcher) buildRequest(image string) (*http.Request, error) {
	var buff bytes.Buffer

	writer := multipart.NewWriter(&buff)
	// 基本表单项
	writer.WriteField("image_url", "")
	writer.WriteField("image_content", "")
	writer.WriteField("filename", "")
	writer.WriteField("hl", "en")

	// 图片文件
	w, err := writer.CreateFormFile("encoded_image", "")
	if err != nil {
		return nil, err
	}
	// 打开图片文件
	f, err := os.Open(image)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// 将图片数据写入w中
	if _, err := io.Copy(w, f); err != nil{
		return nil, err
	}

	var url = "https://images.hk.53yu.com/searchbyimage/upload"

	if !s.mirror {
		url = "https://www.google.com/searchbyimage/upload"
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, &buff)
	if err != nil {
		return nil, err
	}

	// 根据条件设置头部
	if s.mirror {
		req.Header.Set("Host", "images.hk.53yu.com")
	} else {
		req.Header.Set("Host", "www.google.com")
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", s.userAgents[rand.Intn(len(s.userAgents))])

	return req, nil
}

// 从网页中获取到相关图片的base64编码数据
func (s *Searcher) getBase64ImageData(html []byte) ([][]byte, error) {
	data := pattern.FindAllSubmatch(html, -1)
	var temp = make([][]byte, 0)
	for _, value := range data {
		temp = append(temp, value[1])
	}
	if len(temp) == 0 {
		return nil, errors.New("no matches")
	}
	return temp, nil
}

// 将base64编码的图片数据解码并写入文件中
func (s *Searcher) decodeBase64(data []byte, filename string, wg *sync.WaitGroup) error {
	defer wg.Done()
	data = bytes.Replace(data, []byte("\\x3d"), []byte("="), -1)
	var binary = make([]byte, len(data))
	// 进行解码
	_, err := base64.StdEncoding.Decode(binary, data)
	if err != nil {
		s.log.Println("decode base64 data error: ", err)
		return err
	}
	// 写入文件中
	err = ioutil.WriteFile(filename, binary, 0666)
	if err != nil {
		s.log.Println("write file error: ", err)
		return err
	}
	fmt.Printf("写入图片文件 %s 成功\n", filename)
	return nil
}

func (s *Searcher) SetMirror(flag bool) {
	s.mirror = flag
}

func (s *Searcher) SetMaxRetryTimes(times int) {
	if times <= 0 {
		s.maxRetryTimes = 0
	}
	s.maxRetryTimes = times
}

func (s *Searcher) SetDownloadPath(download string) {
	s.download = download
}

func (s *Searcher) SetUploadPath(upload string)  {
	if !s.exist(upload){
		fmt.Println("no such path: ", upload)
		os.Exit(-1)
	}
	if runtime.GOOS == "windows"{
		upload = strings.ReplaceAll(upload, "/", string(os.PathSeparator))
	}
	s.upload = upload
}

func (s *Searcher) SetUserAgents(agents []string) {
	if len(agents) == 0{
		fmt.Println("User-Agent must have one element")
		os.Exit(-1)
	}
	s.userAgents = agents
}

func (s *Searcher) SetLogger(log *log.Logger){
	s.log = log
}

func (s *Searcher) walkFunc(path string, info os.FileInfo, err error) error {
	if !info.IsDir() {
		wg.Add(1)
		fmt.Println("upload file: ", info.Name())
		go func() {
			defer wg.Done()
			var imagesData [][]byte
			counter := 0

			for counter < s.maxRetryTimes{
				time.Sleep(time.Microsecond*time.Duration(rand.Int31n(20)))
				counter ++
				fmt.Printf("第 %d 次尝试上传图片 %s \n", counter, info.Name())
				req, er := s.buildRequest(path)
				if er != nil {
					s.log.Println("build request error")
					return
				}
				html, er := s.SendRequest(req)
				if er != nil {
					s.log.Println(er)
					continue
				}
				imagesData, er = s.getBase64ImageData(html)
				if er != nil {
					continue
				}
				if len(imagesData) != 0{
					break
				}
			}
			if counter >= s.maxRetryTimes{
				s.log.Println("max retry times to upload file ", path)
				return
			}

			fmt.Printf("图片 %s 上传成功\n", info.Name())

			for i, v := range imagesData {
				filename := filepath.Base(info.Name())
				// 文件所在目录
				dir := filepath.Join(filepath.Dir(path), strings.TrimSuffix(filename, filepath.Ext(filename)))
				// 下载图片所在目录
				dir = strings.Replace(dir, s.upload, s.download, 1)
				if !s.exist(dir){
					if er := os.MkdirAll(dir, 0666); er != nil{
						s.log.Fatalf("create path %s error: %s", dir, er.Error())
						return
					}
				}
				wg.Add(1)
				go s.decodeBase64(v, filepath.Join(dir, fmt.Sprintf("%d.jpeg", i+1)), &wg)
			}
		}()
	}

	return nil
}

func (s *Searcher) Run() {
	start := time.Now()
	filepath.Walk(s.upload, s.walkFunc)
	wg.Wait()
	fmt.Printf("Total time: %d s\n", time.Since(start)/time.Second)
}


func (s Searcher) exist(path string)bool{
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist){
		return false
	}
	return true
}