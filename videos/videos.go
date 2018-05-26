package videos

import (
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sync"
	"time"

	data "github.com/hoanglv00/goCrawlerFacebook/data"
	fb "github.com/huandu/facebook"
)

var TOKEN string
var FakeHeaders = map[string]string{
	"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
	"Accept-Charset":  "UTF-8,*;q=0.5",
	"Accept-Encoding": "gzip,deflate,sdch",
	"Accept-Language": "en-US,en;q=0.8",
	"User-Agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3325.146 Safari/537.36",
}

func init() {
	TOKEN = os.Getenv("FBTOKEN")
}

func findAllVideos(videosRet data.FBVideos, baseDir string, ownerName string, id string) {
	dir := fmt.Sprintf("%v/%v", baseDir, ownerName)
	os.MkdirAll(dir, 0755)
	linkChan := make(chan data.VideoData)
	wg := new(sync.WaitGroup)
	for i := 0; i < 1; i++ {
		wg.Add(1)
		go downloadVideoFromLink(dir, linkChan, wg)
	}
	for _, v := range videosRet.Data {
		dlChan := data.VideoData{}
		dlChan.VideoID = v.ID
		dlChan.VideoURL = v.Permalink_url
		linkChan <- dlChan
	}
}

func Get(url, refer string) string {
	headers := map[string]string{}
	if refer != "" {
		headers["Referer"] = refer
	}
	res := Request("GET", url, nil, headers)
	defer res.Body.Close()
	var reader io.ReadCloser
	if res.Header.Get("Content-Encoding") == "gzip" {
		reader, _ = gzip.NewReader(res.Body)
	} else {
		reader = res.Body
	}
	body, _ := ioutil.ReadAll(reader)
	return string(body)
}

func MatchOneOf(text string, patterns ...string) []string {
	var (
		re    *regexp.Regexp
		value []string
	)
	for _, pattern := range patterns {
		// (?flags): set flags within current group; non-capturing
		// s: let . match \n (default false)
		// https://github.com/google/re2/wiki/Syntax
		re = regexp.MustCompile(pattern)
		value = re.FindStringSubmatch(text)
		if len(value) > 0 {
			return value
		}
	}
	return nil
}

func Request(
	method, url string, body io.Reader, headers map[string]string,
) *http.Response {
	transport := &http.Transport{
		DisableCompression:  true,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	client := &http.Client{
		Transport: transport,
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Print(url)
		panic(err)
	}
	for k, v := range FakeHeaders {
		req.Header.Set(k, v)
	}
	req.Header.Set("Referer", url)
	for k, v := range headers {
		req.Header.Set(k, v)
		fmt.Println(req)
	}
	retryTimes := 3
	var (
		res          *http.Response
		requestError error
	)
	for i := 0; i < retryTimes; i++ {
		res, requestError = client.Do(req)
		if requestError == nil {
			break
		}
		if requestError != nil && i+1 == retryTimes {
			log.Print(url)
			panic(requestError)
		}
		time.Sleep(1 * time.Second)
	}
	return res
}

func FileSize(filePath string) (int64, bool) {
	file, err := os.Stat(filePath)
	if err != nil && os.IsNotExist(err) {
		return 0, false
	}
	return file.Size(), true
}

func downloadVideoFromLink(baseDir string, linkChan chan data.VideoData, wg *sync.WaitGroup) {
	defer wg.Done()

	for target := range linkChan {
		downloadLink := "https://www.facebook.com" + target.VideoURL
		html := Get(downloadLink, downloadLink)
		//title := MatchOneOf(html, `<title id="pageTitle">(.+)</title>`)[1]

		//var tag string
		var u string
		for _, quality := range []string{"sd"} {
			//tag := quality
			u = MatchOneOf(
				html, fmt.Sprintf(`%s_src_no_ratelimit:"(.+?)"`, quality))[1]
		}
		downloadLink = u

		var filePath = fmt.Sprintf("%v/videos/%v.mp4", baseDir, target.VideoID)
		tempFilePath := filePath //+ ".download"
		tempFileSize, _ := FileSize(tempFilePath)
		headers := map[string]string{
			"Referer": downloadLink,
		}
		var file *os.File
		if tempFileSize > 0 {
			// range start from zero
			headers["Range"] = fmt.Sprintf("bytes=%d-", tempFileSize)
			file, _ = os.OpenFile(tempFilePath, os.O_APPEND|os.O_WRONLY, 0644)
		} else {
			file, _ = os.Create(tempFilePath)
		}

		// defer func() {
		// 	file.Close()
		// 	// must close the file before rename or it will cause `The process cannot access the file because it is being used by another process.` error.
		// 	err := os.Rename(tempFilePath, filePath)
		// 	if err != nil {
		// 		log.Fatal(err)
		// 	}
		// }()

		res := Request("GET", downloadLink, nil, headers)
		if res.StatusCode >= 400 {
			log.Fatal(fmt.Sprintf("HTTP error: %d", res.StatusCode))
		}
		fmt.Println(res.Body)

		defer res.Body.Close()
		defer file.Close()
		_, err := io.Copy(file, res.Body)
		if err != nil {
			log.Println("download video err=", err)
		}
	}
}

func runFBGraphAPIVideos(query string) (queryResult interface{}) {
	res, err := fb.Get(query, fb.Params{
		"access_token": TOKEN,
		"fields":       "permalink_url,updated_time,description,id",
	})

	if err != nil {
		log.Fatalln("FB connect error, err=", err.Error())
	}
	return res
}
