package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func downloadWndCreate(localSavePath, mangaUrl string) *mangaPullControlWnd {
	var wnd mangaPullControlWnd
	Dialog{
		AssignTo: &wnd.main,
		Title:    mangaUrl,
		MinSize:  Size{Width: 360, Height: 600},

		Layout: VBox{},
		Children: []Widget{
			ImageView{
				MinSize:  Size{Width: 120, Height: 180},
				MaxSize:  Size{Width: 350, Height: 494},
				Mode:     ImageViewModeZoom,
				AssignTo: &wnd.mangaCover,
			},
			PushButton{
				Text: "(╯°□°）╯︵ ┻━┻ metainfo repull",

				OnClicked: func() {
					go wnd.pullMetaInfo(nil, localSavePath, mangaUrl)
				},
				// Visible:  false,
				Enabled:  false,
				AssignTo: &wnd.metarepull,
			},
			PushButton{
				Text: "(┬┬﹏┬┬) images repull",
				OnClicked: func() {
					go wnd.pullImages(nil, localSavePath, mangaUrl, false)
				},
				// Visible:  false,
				Enabled:  false,
				AssignTo: &wnd.imagesrepull,
			},
			PushButton{
				Text: "zip images",
				OnClicked: func() {
					wnd.Compress(localSavePath)
				},
				Enabled: false,
				// Visible:  false,
				AssignTo: &wnd.zipImage,
			},
			Label{AssignTo: &wnd.pullStepInfo},
			ProgressBar{
				AssignTo: &wnd.pullProgressBar,
			},
			PushButton{
				Text: "(^///^) open manga's folder",
				OnClicked: func() {
					mangaLocalFolder := PATHFINAL + "\\" + wnd.main.Title()
					openExplorerFolder(mangaLocalFolder)
				},
			},
		},
	}.Create(nil)

	wnd.path = localSavePath

	return &wnd
}

// func (w *mangaPullControlWnd) pullManga(savePath, mangaURL string) {
// 	go w.pullImages(savePath, mangaURL)
// 	go w.pullMetaInfo(savePath, mangaURL)
// }

func catchCover(result string) (string, error) {
	coverstr := `<img is="lazyload-image" class="lazyload" width="350" height="" data-src="(.+)" src=`
	coverret := regexp.MustCompile(coverstr)
	coverarray := coverret.FindStringSubmatch(result)
	if len(coverarray) < 2 {
		fmt.Println("Cover array not match")
		return "", errors.New(" Cover array not match")
	}
	return coverarray[1], nil
}

func (w *mangaPullControlWnd) pullMetaInfo(ch chan string, savePath, mangaUrl string) error {
	result, err := HttpGet(mangaUrl)
	if err != nil {
		fmt.Println("pull meta info failed:", err)
		return errors.New(fmt.Sprintf(" Get cover err:%v", err))
	}
	mangaCoverUrl, err := catchCover(result)
	if err != nil {
		fmt.Println("no cover url catched:", err)
		return errors.New("cover url catched")
	}

	// waiting for title from image goroutine
	var mangaTitle string
	if ch != nil {
		mangaTitle = <-ch
	} else {
		for i := 0; i < 20; i++ {
			if len(w.title) != 0 {
				mangaTitle = w.title
				break
			}
			time.Sleep(500 * time.Microsecond)
		}
	}
	mangaLocalDir := savePath + "/" + mangaTitle
	CreatDirIfNotHad(savePath)
	CreatDirIfNotHad(mangaLocalDir)

	metaInfoMarkDown, err := os.OpenFile(mangaLocalDir+"/"+"meta.md", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Println("open meta file Failed:", err)
	} else {
		metaInfoMarkDown.WriteString("### " + mangaTitle + "\n")
		metaInfoMarkDown.WriteString(mangaUrl + "\n")
		metaInfoMarkDown.Sync()
		metaInfoMarkDown.Close()
	}

	mangaSavName := mangaLocalDir + "/" + mangaTitle + filepath.Ext(mangaCoverUrl)
	fmt.Println(mangaCoverUrl)
	if nil == pullImageFile(mangaSavName, mangaCoverUrl) {
		img, err := walk.NewImageFromFileForDPI(mangaSavName, 96)
		if err != nil {
			return errors.New(" Load mangaCover image from file imagePullFailed.")
		}
		w.mangaCover.SetImage(img)
	}
	return nil
}
func catchTitle(mangaListPageResult string) (string, error) {
	titleMatchPolicyStr := `<title>(.*)&raquo;`
	titleMatchPolicy := regexp.MustCompile(titleMatchPolicyStr)
	titlearray := titleMatchPolicy.FindStringSubmatch(mangaListPageResult)
	if len(titlearray) < 2 {
		return "", errors.New("mangaTitle no match correct")
	}
	return titlearray[1], nil
}
func catchImagesUrls(mangaListPageResult string) []string {
	var imagesUrls = make([]string, 0)
	str := `<img class="list-img lazyload" src=".+" data-src=".+"`
	ret := regexp.MustCompile(str)
	urls := ret.FindAllStringSubmatch(mangaListPageResult, -1)
	imageUrlMatchPolicyStr := `^<img class="list-img lazyload" src=".+" data-src="(.+)" onerror="`
	imageUrlMatchPolicy := regexp.MustCompile(imageUrlMatchPolicyStr)
	for _, a := range urls {
		imageMatchResult := imageUrlMatchPolicy.FindStringSubmatch(a[0])
		if 2 == len(imageMatchResult) {
			imagesUrls = append(imagesUrls, imageMatchResult[1])
		}
	}
	return imagesUrls
}

func beautyTitle(origin string) (mangaTitle string) {
	mangaTitle = strings.Replace(origin, "/", "", -1)
	mangaTitle = strings.Replace(mangaTitle, " ", "", -1)
	mangaTitle = strings.Replace(mangaTitle, "|", "", -1)
	mangaTitle = strings.Replace(mangaTitle, ":", "", -1)
	mangaTitle = strings.Replace(mangaTitle, "*", "", -1)
	mangaTitle = strings.Replace(mangaTitle, "?", "", -1)
	mangaTitle = strings.Replace(mangaTitle, "\"", "", -1)
	mangaTitle = strings.Replace(mangaTitle, "<", "", -1)
	mangaTitle = strings.Replace(mangaTitle, ">", "", -1)
	if strings.HasPrefix(mangaTitle, "[同人誌H漫畫]") {
		mangaTitle = mangaTitle[len("[同人誌H漫畫]"):]
	}
	return
}

func (w *mangaPullControlWnd) pullImages(ch chan string, savePath, mangaUrl string, autoCompress bool) int {
	mangaListPageUrl := mangaUrl + "list2/"
	fmt.Println("Pull manga images from", mangaListPageUrl)
	mangaListPageResult, err := HttpGet(mangaListPageUrl)
	if err != nil {
		fmt.Println("Get image list page err:", err)
		w.pullStepInfo.SetText(err.Error())
		return -1
	}
	mangaTitle, err := catchTitle(mangaListPageResult)
	if err != nil {
		w.pullStepInfo.SetText(err.Error())
		return -1
	}
	mangaTitle = beautyTitle(mangaTitle)
	w.title = mangaTitle
	w.main.SetTitle(mangaTitle)
	if ch != nil {
		ch <- mangaTitle
	}

	imagesSavePath := savePath + "/" + mangaTitle + "/" + DIRIMAGES
	CreatDirIfNotHad(savePath)
	CreatDirIfNotHad(savePath + "/" + mangaTitle)
	CreatDirIfNotHad(imagesSavePath)

	imagesUrls := catchImagesUrls(mangaListPageResult)
	if len(imagesUrls) == 0 {
		w.pullStepInfo.SetText("Error at fetch pics, recv 0 picsurl from regexp")
		return -1
	}
	downloadUrls := weedOutByLocal(imagesSavePath, imagesUrls)
	if downloadUrls == nil || len(downloadUrls) == 0 {
		w.pullStepInfo.SetText("all images are already on your disk.")
		return 0
	}

	w.autoDownload(imagesSavePath, downloadUrls, autoCompress)

	return 0
}

func weedOutByLocal(dirpath string, picurls []string) []string {
	var neededurls = make([]string, 0)
	for _, picurl := range picurls {
		_, fileName := filepath.Split(picurl)
		localname := dirpath + "/" + fileName
		exsitmark, err := PathExists(localname)
		if err != nil {
			fmt.Println("weedOutByLocal caught error:", err)
			continue
		}
		if exsitmark == false {
			neededurls = append(neededurls, picurl)
		}
	}
	return neededurls
}

func (w *mangaPullControlWnd) updateImagePullStatus(val int) {
	w.updatelock.Lock()
	w.imageProcessed++
	if val == -1 {
		w.imagePullFailed++
	} else {
		w.imagePullSucceed++
	}
	w.pullProgressBar.SetValue(w.imageProcessed)
	w.updatelock.Unlock()
}

func (w *mangaPullControlWnd) Compress(path string) {
	srcpath := path + "/" + w.title + "/" + DIRIMAGES + "/"
	zipname := path + "/" + w.title + "/" + w.title + ".zip"
	fmt.Println("AutoCompress imagePath:", srcpath)
	fmt.Println("AutoCompress outputZip:", zipname)
	w.pullStepInfo.SetText("Compressing...")
	Zip(srcpath, zipname)
	w.pullStepInfo.SetText("Compressing complete.")
}

func threadGen(total int) (count int) {
	if total <= 1 {
		count = 1
	} else if total <= 15 {
		count = 2
	} else if total <= 30 {
		count = 3
	} else if total <= 80 {
		count = 3
	} else if total <= 140 {
		count = 6
	} else if total <= 180 {
		count = 6
	} else if total <= 209 {
		count = 6
	} else {
		count = 8
	}
	return count
}

func (w *mangaPullControlWnd) autoDownload(imageSavPath string, imagesUrls []string, autoCompress bool) {

	w.imageTotal = len(imagesUrls)
	threadCount := threadGen(w.imageTotal)

	info := fmt.Sprintf("%v thread(s) for %v images...", threadCount, w.imageTotal)
	w.pullStepInfo.SetText(info)
	w.pullProgressBar.SetRange(0, w.imageTotal)
	arry := splitArray(imagesUrls, threadCount)

	wg := sync.WaitGroup{}
	for i, urls := range arry {
		wg.Add(1)
		go func(workid int, urls []string) {
			for _, imageUrl := range urls {
				_, imageFileName := filepath.Split(imageUrl)
				localName := imageSavPath + "/" + imageFileName
				err := pullImageFile(localName, imageUrl)
				if err != nil {
					fmt.Println("Download image file error:", err)
					w.updateImagePullStatus(-1)
				} else {
					w.updateImagePullStatus(1)
				}
			}
			wg.Done()
		}(i, urls)
	}
	wg.Wait()
	if w.imageTotal == w.imageProcessed {
		if autoCompress {
			w.Compress(w.path)
		}
		w.pullStepInfo.SetText(fmt.Sprintf("Total:%d  Pulled:%d  Failed:%d", w.imageTotal, w.imagePullSucceed, w.imagePullFailed))

		w.zipImage.SetEnabled(true)
		w.metarepull.SetEnabled(true)
		w.imagesrepull.SetEnabled(true)

		ni, _ := walk.NewNotifyIcon(w.main.Form())
		ni.SetVisible(true)
		ni.ShowInfo("Finshed", w.title)
		ni.Dispose()
	}
}

func pullImageFile(localname, url string) error {
	fmt.Println("Pulling:", "[", url, "]", " >> ", localname)
	resp, err := http.Get(url)
	if err != nil {
		return errors.New(fmt.Sprintf(" HTTP get failed:%v", err.Error()))
	}

	if resp.Status == "200 OK" {
	} else if resp.Status == "404 Not Found" {
		switch strings.ToLower(filepath.Ext(url)) {
		case ".png":
			localname = localname[:len(localname)-3] + "jpg"
			url = url[:len(url)-3] + "jpg"
		case ".jpg":
			localname = localname[:len(localname)-3] + "png"
			url = url[:len(url)-3] + "png"
		default:
			return errors.New(fmt.Sprintf("unknwon  file external, error:%v", resp.Status))
		}

		resp, err = http.Get(url)
		if err != nil {
			return errors.New(fmt.Sprintf(" HTTP get failed:%v", err.Error()))
		}
		if resp.Status != "200 OK" {
			return errors.New(fmt.Sprintf(" HTTP bad status:%v", resp.Status))
		}
	} else {
		return errors.New(fmt.Sprintf(" HTTP bad status:%v", resp.Status))
	}

	// 保存在本地绝对名字
	f, err := os.Create(localname)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := make([]byte, 4096)
	for {
		n, err2 := resp.Body.Read(buf)
		if n == 0 {
			break
		}
		if err2 != nil && err2 != io.EOF {
			err = err2
			return err
		}
		// 图片文件落地本地文件
		f.Write(buf[:n])
	}
	return nil
}

type mangaPullControlWnd struct {
	main                               *walk.Dialog
	mangaCover                         *walk.ImageView
	pullStepInfo                       *walk.Label
	pullProgressBar                    *walk.ProgressBar
	metarepull, imagesrepull, zipImage *walk.PushButton

	imageTotal, imageProcessed, imagePullSucceed, imagePullFailed int
	updatelock                                                    sync.Mutex
	path, title                                                   string
}

// type mangaHandler struct {
// 	wnd                       *mangaPullControlWnd
// 	localSavePath, mangaURL   string
// 	imagesTotal, imagesPulled int
// }

func pullMangaProject(localSavePath, mangaUrl string, autoCompress bool) {
	wnd := downloadWndCreate(localSavePath, mangaUrl)
	if wnd == nil {
		fmt.Println("Error create manga pull wnd")
		return
	}

	wnd.main.Starting().Attach(func() {
		ch := make(chan string, 1)
		go wnd.pullImages(ch, localSavePath, mangaUrl, autoCompress)
		go wnd.pullMetaInfo(ch, localSavePath, mangaUrl)
	})

	wnd.main.Run()
	return
}
