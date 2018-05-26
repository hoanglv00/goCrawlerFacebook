package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"sync"

	data "github.com/hoanglv00/goCrawlerFacebook/data"
	photos "github.com/hoanglv00/goCrawlerFacebook/photos"
	videos "github.com/hoanglv00/goCrawlerFacebook/videos"
	fb "github.com/huandu/facebook"
)

var pageName = flag.String("n", "", "Facebook page name such as: scottiepippen")
var numOfWorkersPtr = flag.Int("c", 2, "the number of concurrent rename workers. default = 2")
var m sync.Mutex
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

func runFBGraphAPI(query string) (queryResult interface{}) {
	res, err := fb.Get(query, fb.Params{
		"access_token": TOKEN,
	})

	if err != nil {
		log.Fatalln("FB connect error, err=", err.Error())
	}
	return res
}

func main() {
	flag.Parse()
	var inputPage string
	if TOKEN == "" {
		log.Fatalln("Set your FB token as environment variables 'export FBTOKEN=XXXXXXX'")
	}

	if *pageName == "" {
		log.Fatalln("You need to input -n=Name_or_Id.")
	}
	inputPage = *pageName

	//Get system user folder
	usr, _ := user.Current()
	baseDir := fmt.Sprintf("%v/Pictures/goFBPages", usr.HomeDir)

	//Get User info
	resUser := runFBGraphAPI("/" + inputPage)
	userRet := data.FBUser{}
	photos.ParseMapToStruct(resUser, &userRet)

	//Get all videos
	resVideos := videos.RunFBGraphAPIVideos("/" + inputPage + "/videos.limit(100)")
	videosRet := data.FBVideos{}
	photos.ParseMapToStruct(resVideos, &videosRet)

	//Get all albums
	resAlbums := photos.RunFBGraphAPIAlbums("/" + inputPage + "/albums")
	albumRet := data.FBAlbums{}
	photos.ParseMapToStruct(resAlbums, &albumRet)

	//use limit to avoid error: Please reduce the amount of data you're asking for, then retry your request
	//Curently 30 is a magic number of FB Graph API call, 50 will still occur failed.  >_<
	// maxCount := 30

	// userFolderName := fmt.Sprintf("[%s]%s", userRet.Username, userRet.Name)
	// for _, v := range albumRet.Data {
	// 	fmt.Println("Starting download ["+v.Name+"]-"+v.From.Name, " total count:", v.Count)

	// 	if v.Count > maxCount {
	// 		currentOffset := 0
	// 		for {
	// 			if currentOffset > v.Count {
	// 				break
	// 			}
	// 			photos.FindPhotoByAlbum(userFolderName, v.Name, v.ID, baseDir, maxCount, currentOffset)
	// 			currentOffset = currentOffset + maxCount
	// 		}
	// 	} else {
	// 		photos.FindPhotoByAlbum(userFolderName, v.Name, v.ID, baseDir, v.Count, 0)
	// 	}

	// }
	// for _, v := range videosRet.Data {
	// 	fmt.Println(v.ID)
	// }
	videos.FindAllVideos(videosRet, baseDir, userRet.Name, userRet.ID)
}
