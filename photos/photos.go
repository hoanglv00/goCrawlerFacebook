package photos

import (
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	data "github.com/hoanglv00/goCrawlerFacebook/data"
	fb "github.com/huandu/facebook"
)

var TOKEN string

func init() {
	TOKEN = os.Getenv("FBTOKEN")
}

//Use to change json result to struct
func ParseMapToStruct(inData interface{}, decodeStruct interface{}) {
	jret, _ := json.Marshal(inData)
	err := json.Unmarshal(jret, &decodeStruct)
	if err != nil {
		log.Fatalln(err)
	}
}

//Use to download photos
func DownloadWorker(destDir string, linkChan chan data.DLData, wg *sync.WaitGroup) {
	defer wg.Done()

	for target := range linkChan {
		var imageType string
		if strings.Contains(target.ImageSource, ".png") {
			imageType = ".png"
		} else {
			imageType = ".jpg"
		}

		resp, err := http.Get(target.ImageSource)
		if err != nil {
			log.Println("Http.Get\nerror: " + err.Error() + "\ntarget: " + target.ImageSource)
			continue
		}
		defer resp.Body.Close()

		m, _, err := image.Decode(resp.Body)
		if err != nil {
			log.Println("image.Decode\nerror: " + err.Error() + "\ntarget: " + target.ImageSource)
			continue
		}

		// Ignore small images
		bounds := m.Bounds()
		if bounds.Size().X > 300 && bounds.Size().Y > 300 {
			out, err := os.Create(destDir + "/" + target.ImageID + imageType)
			if err != nil {
				log.Println("os.Create\nerror: %s", err)
				continue
			}
			defer out.Close()
			if imageType == ".png" {
				png.Encode(out, m)
			} else {
				jpeg.Encode(out, m, nil)
			}
		}
	}
}

func FindPhotoByAlbum(ownerName string, albumName string, albumId string, baseDir string, photoCount int, photoOffset int) {
	photoRet := data.FBPhotos{}
	var queryString string
	if photoOffset > 0 {
		queryString = fmt.Sprintf("/%s/photos?limit=%d&offset=%d", albumId, photoCount, photoOffset)

	} else {
		queryString = fmt.Sprintf("/%s/photos?limit=%d", albumId, photoCount)
	}

	resPhoto := RunFBGraphAPIPhotos(queryString)
	ParseMapToStruct(resPhoto, &photoRet)
	dir := fmt.Sprintf("%v/%v/%v - %v", baseDir, ownerName, albumId, albumName)
	os.MkdirAll(dir, 0755)

	linkChan := make(chan data.DLData)
	wg := new(sync.WaitGroup)
	for i := 0; i < 1; i++ {
		wg.Add(1)
		go DownloadWorker(dir, linkChan, wg)
	}
	//Send data to DownloadWorker
	for _, v := range photoRet.Data {
		dlChan := data.DLData{}
		dlChan.ImageID = v.ID
		dlChan.ImageURL = v.Link
		dlChan.ImageSource = v.Images[0].Source
		linkChan <- dlChan
	}
}

//Get from, count and name albums by facebook api
func RunFBGraphAPIAlbums(query string) (queryResult interface{}) {
	res, err := fb.Get(query, fb.Params{
		"access_token": TOKEN,
		"fields":       "from,count,name",
	})

	if err != nil {
		log.Fatalln("FB connect error, err=", err.Error())
	}
	return res
}

//Get link, images photos by facebook api
func RunFBGraphAPIPhotos(query string) (queryResult interface{}) {
	res, err := fb.Get(query, fb.Params{
		"access_token": TOKEN,
		"fields":       "link,images",
	})

	if err != nil {
		log.Fatalln("FB connect error, err=", err.Error())
	}
	return res
}
