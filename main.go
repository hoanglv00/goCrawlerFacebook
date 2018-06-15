package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
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
var crawler = flag.String("s", "", "Crawler videos or photos")
var numOfWorkersPtr = flag.Int("c", 2, "The number of concurrent rename workers. default = 2")
var m sync.Mutex
var TOKEN string

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

	data.ThreadNumber = *numOfWorkersPtr

	//Get system user folder
	usr, _ := user.Current()
	baseDir := fmt.Sprintf("%v/Pictures/goFBPages", usr.HomeDir)

	//baseDir := "D:/goFBPages"

	//Get User info
	resUser := runFBGraphAPI("/" + inputPage)
	userRet := data.FBUser{}
	photos.ParseMapToStruct(resUser, &userRet)
	userFolderName := fmt.Sprintf("[%s]%s", userRet.Username, userRet.Name)

	if *crawler == "" {
		dir := fmt.Sprintf("%v/%v", baseDir, userFolderName)
		resUserJson, err := json.Marshal(userRet)
		if err != nil {
			log.Fatalln("marshal error, err=", err)
		}
		os.MkdirAll(dir, 0755)
		err = ioutil.WriteFile(dir+"/PageInfor.json", resUserJson, 0644) // Ghi dữ liệu vào file JSON
		if err != nil {
			log.Fatalln("write file error, err=", err)
		}

	} else if *crawler == "videos" {
		//Get all videos
		resVideos := videos.RunFBGraphAPIVideos("/" + inputPage + "/videos?limit=100")
		videosRet := data.FBVideos{}
		photos.ParseMapToStruct(resVideos, &videosRet)

		videos.FindAllVideos(videosRet, baseDir, userRet.Name, userRet.ID)

	} else if *crawler == "photos" {
		//Get all albums
		resAlbums := photos.RunFBGraphAPIAlbums("/" + inputPage + "/albums")
		albumRet := data.FBAlbums{}
		photos.ParseMapToStruct(resAlbums, &albumRet)

		//use limit to avoid error: Please reduce the amount of data you're asking for, then retry your request
		//Curently 30 is a magic number of FB Graph API call, 50 will still occur failed.  >_<
		maxCount := 30

		for _, v := range albumRet.Data {
			fmt.Println("Starting download ["+v.Name+"]-"+v.From.Name, " total count:", v.Count)

			if v.Count > maxCount {
				currentOffset := 0
				for {
					if currentOffset > v.Count {
						break
					}
					photos.FindPhotoByAlbum(userFolderName, v.Name, v.ID, baseDir, maxCount, currentOffset)
					currentOffset = currentOffset + maxCount
				}
			} else {
				photos.FindPhotoByAlbum(userFolderName, v.Name, v.ID, baseDir, v.Count, 0)
			}

		}
	} else {
		log.Fatalln("You need to input -s=videos_or_photos.")
	}
}
