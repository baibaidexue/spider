package main

import (
	"errors"
	"fmt"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
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
					wnd.pullMetaInfo(localSavePath, mangaUrl)
				},
			},
			PushButton{
				Text: "(┬┬﹏┬┬) images repull",
				OnClicked: func() {
					wnd.pullImages(localSavePath, mangaUrl)
				},
			},
			PushButton{
				Text: "zip images",
				OnClicked: func() {
					wnd.Compress(localSavePath)
				},
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

func (w *mangaPullControlWnd) pullManga(savePath, mangaURL string) {
	go w.pullImages(savePath, mangaURL)
	go w.pullMetaInfo(savePath, mangaURL)
}

func (w *mangaPullControlWnd) pullMetaInfo(savePath, mangaUrl string) error {
	result, err := HttpGet(mangaUrl)
	if err != nil {
		fmt.Println("pull meta info failed:", err)
		return errors.New(fmt.Sprintf("Spycover err:", err))
	}

	coverstr := `<img is="lazyload-image" class="lazyload" width="350" height="" data-src="(.+)" src=`
	coverret := regexp.MustCompile(coverstr)
	coverarray := coverret.FindStringSubmatch(result)
	if len(coverarray) < 2 {
		fmt.Println("Cover array not match")
		return errors.New("Cover array not match")
	}
	mangaCoverUrl := coverarray[1]

	// waiting for title set in pull images
	var mamgaTitle string
	for i := 0; i < 20; i++ {
		mamgaTitle = w.title
		if len(mamgaTitle) != 0 {
			break
		}
		time.Sleep(time.Second)
	}
	if mamgaTitle == mangaUrl {
		fmt.Println("pull metainfo get title error")
		return errors.New("mamgaTitle get error")
	}

	mangaLocalDir := savePath + "/" + mamgaTitle
	CreatDirIfNotHad(savePath)
	CreatDirIfNotHad(mangaLocalDir)

	metaInfoMarkDown, err := os.OpenFile(mangaLocalDir+"/"+"meta.md", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Println("write mangaUrl file imagePullFailed", err)
	} else {
		metaInfoMarkDown.WriteString("### " + mamgaTitle + "\n")
		metaInfoMarkDown.WriteString(mangaUrl + "\n")
		metaInfoMarkDown.Sync()
		metaInfoMarkDown.Close()
	}

	mangaSavName := mangaLocalDir + "/" + mamgaTitle + filepath.Ext(mangaCoverUrl)
	fmt.Println(mangaCoverUrl)
	if nil == pullImageFile(mangaSavName, mangaCoverUrl) {
		img, err := walk.NewImageFromFileForDPI(mangaSavName, 96)
		if err != nil {
			return errors.New("Load mangaCover image from file imagePullFailed.")
		}
		w.mangaCover.SetImage(img)
	}
	return nil
}

func (w *mangaPullControlWnd) pullImages(savePath, mangaUrl string) int {
	w.pullStepInfo.SetText("Fetching manga image lists...")

	mangaListPageUrl := fmt.Sprintf("%vlist2/", mangaUrl)
	fmt.Println("Pull manga images from", mangaListPageUrl)
	mangaListPageResult, err := HttpGet(mangaListPageUrl)
	if err != nil {
		fmt.Println("Get image list page err:", err)
		w.pullStepInfo.SetText(err.Error())
		return -1
	}

	titleMatchPolicyStr := `<title>(.*)&raquo;`
	titleMatchPolicy := regexp.MustCompile(titleMatchPolicyStr)
	titlearray := titleMatchPolicy.FindStringSubmatch(mangaListPageResult)
	if len(titlearray) < 2 {
		w.pullStepInfo.SetText("mangaTitle array error")
		return -1
	}
	mangaTitle := titlearray[1]
	mangaTitle = strings.Replace(mangaTitle, " ", "", -1)
	mangaTitle = strings.Replace(mangaTitle, "/", "", -1)
	if strings.HasPrefix(mangaTitle, "[同人誌H漫畫]") {
		mangaTitle = mangaTitle[len("[同人誌H漫畫]"):]
	}
	w.main.SetTitle(mangaTitle)
	w.title = mangaTitle

	imagesSavePath := savePath + "/" + mangaTitle + "/" + DIRIMAGES
	CreatDirIfNotHad(savePath)
	CreatDirIfNotHad(savePath + "/" + mangaTitle)
	CreatDirIfNotHad(imagesSavePath)

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

	if len(imagesUrls) == 0 {
		w.pullStepInfo.SetText("Error at fetch pics, recv 0 picsurl from regexp")
		return -1
	}

	// check if exsit already files
	notexsitsrcs := delbylocal(imagesSavePath, imagesUrls)
	if notexsitsrcs == nil {
		w.pullStepInfo.SetText("no need to pull for all images are already on your disk.")
		return 0
	}

	total := len(notexsitsrcs)
	if total == 0 {
		info := fmt.Sprintf("All images were pulled to local already.")
		w.pullStepInfo.SetText(info)
		return 0
	}

	w.imageTotal = total
	info := fmt.Sprintf("%v thread(s) for %v images...", threadGen(total), total)
	w.pullProgressBar.SetRange(0, total)
	w.pullStepInfo.SetText(info)

	w.autoDownload(imagesSavePath, notexsitsrcs)

	return 0
}

func delbylocal(dirpath string, picurls []string) []string {
	var neededurls = make([]string, 0)
	for _, picurl := range picurls {
		_, fileName := filepath.Split(picurl)
		localname := dirpath + "/" + fileName
		exsitmark, err := PathExists(localname)
		if err != nil {
			fmt.Println("delbylocal caught error:", err)
			continue
		}
		if exsitmark == false {
			neededurls = append(neededurls, picurl)
		}
	}
	return neededurls
}

func (w *mangaPullControlWnd) updateImagePullStatus(val int) {
	w.imageProcessed++
	if val == -1 {
		w.imagePullFailed++
	} else {
		w.imagePullSucceed++
	}

	w.pullProgressBar.SetValue(w.imageProcessed)
	if w.imageTotal == w.imageProcessed {
		// compress
		w.pullStepInfo.SetText("Compressing...")
		w.Compress(w.path)
		w.pullStepInfo.SetText(fmt.Sprintf("Total:%d  Pulled:%d  Failed:%d", w.imageTotal, w.imagePullSucceed, w.imagePullFailed))

		ni, _ := walk.NewNotifyIcon(w.main.Form())
		ni.SetVisible(true)
		ni.ShowInfo("Finshed", w.title)
		ni.Dispose()
	}
}

func (w *mangaPullControlWnd) Compress(path string) {
	srcpath := path + "/" + w.title + "/" + DIRIMAGES + "/"
	zipname := path + "/" + w.title + "/" + w.title + ".zip"
	fmt.Println("AutoZip imagePath:", srcpath)
	fmt.Println("AutoZip outputZip:", zipname)
	Zip(srcpath, zipname)
}

func threadGen(total int) int {
	var gates int
	if total <= 15 {
		gates = 2
	} else if total <= 30 {
		gates = 3
	} else if total <= 80 {
		gates = 3
	} else if total <= 140 {
		gates = 6
	} else if total <= 180 {
		gates = 6
	} else if total <= 209 {
		gates = 6
	} else {
		gates = 8
	}
	return gates
}

func (w *mangaPullControlWnd) autoDownload(imageSavPath string, imagesUrls []string) {
	// 切片计算
	gates := threadGen(len(imagesUrls))
	arry := splitArray(imagesUrls, gates)

	for i, url := range arry {
		go func(workid int, url []string) {
			for _, imageUrl := range url {
				_, imageFileName := filepath.Split(imageUrl)
				localname := imageSavPath + "/" + imageFileName
				fmt.Println("Pulling:", "[", imageUrl, "]", " >> ", localname)
				err := pullImageFile(localname, imageUrl)
				if err != nil {
					fmt.Println("Pull error", err)
					w.updateImagePullStatus(-1)
				} else {
					w.updateImagePullStatus(1)
				}
			}
		}(i, url)
	}
}

//写入文件
func pullImageFile(localname, url string) error {
	//读取url的信息
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("pull image file http get error:", err)
		return err
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
			fmt.Println("unknwon  file external, error:", resp.Status)
			return errors.New(resp.Status)
		}

		fmt.Println("Repull as >>> ", localname)
		resp, err = http.Get(url)
		if err != nil {
			return err
		}
		if resp.Status != "200 OK" {
			return errors.New(resp.Status)
		}
	} else {
		return errors.New(resp.Status)
	}

	//保存在本地绝对名字
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
	main            *walk.Dialog
	mangaCover      *walk.ImageView
	pullStepInfo    *walk.Label
	pullProgressBar *walk.ProgressBar

	imageTotal, imageProcessed, imagePullSucceed, imagePullFailed int
	updatelock                                                    sync.Mutex
	path, title                                                   string
}

type mangaHandler struct {
	wnd                       *mangaPullControlWnd
	localSavePath, mangaURL   string
	imagesTotal, imagesPulled int
}

func pullMangaProject(localSavePath, mangaUrl string) {
	handle := &mangaHandler{
		localSavePath: localSavePath,
		mangaURL:      mangaUrl,
	}

	handle.wnd = downloadWndCreate(localSavePath, mangaUrl)
	if handle.wnd == nil {
		fmt.Println("Error create manga pull wnd")
		return
	}

	// handle.wnd.main.SetFocus()
	handle.wnd.main.Starting().Attach(func() {
		handle.wnd.pullManga(localSavePath, mangaUrl)
	})

	go handle.wnd.main.Run()

	return
}
