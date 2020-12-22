package gis

import (
	"log"
	"net/http"
	"os"
	"testing"
)

func newSearcher() *Searcher{
	var client = http.Client{}
	searcher := NewSearcher(client)
	searcher.SetUploadPath("../upload")
	searcher.SetDownloadPath("download")
	searcher.SetMaxRetryTimes(66)
	f, _ := os.Create("gis.log")
	l := log.New(f, "", log.LstdFlags)
	l.SetOutput(f)

	searcher.SetLogger(l)
	return searcher
}

var searcher  = newSearcher()

func TestSendRequest(t *testing.T) {
	req, err := searcher.buildRequest("golang.png")
	if err != nil{
		t.Fatal("error: ", err)
	}
	html, err := searcher.SendRequest(req)
	if err != nil {
		t.Log(err)
		t.Fatal("send request error")
	}

	_, err = searcher.getBase64ImageData(html)
	if err != nil{
		t.Fatal(err)
	}
}


func TestSearcher_Run(t *testing.T) {
	searcher.Run()
}

