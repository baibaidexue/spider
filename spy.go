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

func workWndCreate(path, url string)*workerwnd{
	var wnd workerwnd
	Dialog{
		AssignTo:      &wnd.main,
		Title:         url,
		MinSize: Size{Width: 360, Height: 450},
		Layout:  VBox{},
		Children: []Widget{
			ImageView{
				MinSize: Size{Width: 120, Height: 160},
				MaxSize: Size{Width: 350, Height: 494},
				Mode:	ImageViewModeZoom,
				AssignTo: &wnd.cover,
			},
			Label{AssignTo: &wnd.rate},
			ProgressBar{
				AssignTo: &wnd.pb,
			},
		},
	}.Create(nil)

	return &wnd
}

func (w *workerwnd)spyCover(path, url string )error {
	result,err :=HttpGet(url)
	//判断是否出错，并打印信息
	if err!=nil{
		fmt.Println("Spycover err:",err)
		return errors.New(fmt.Sprintf("Spycover err:",err))
	}

	coverstr :=`<img is="lazyload-image" class="lazyload" width="350" height="" data-src="(.+)" src=`
	coverret :=regexp.MustCompile(coverstr)
	coverarray :=coverret.FindStringSubmatch(result)
	if len(coverarray) < 2{
		fmt.Println("Cover array not match")
		return errors.New("Cover array not match")
	}
	coverurl := coverarray[1]
	var projectName string

	for i:=0; i < 10; i++{
		projectName = w.main.Title()
		if projectName != url{
			break
		}
		time.Sleep(time.Second)
	}

	if projectName == url{
		fmt.Println("projectName get error")
		return errors.New("projectName get error")
	}

	projectDir := PATHFINAL + "/" + projectName
	CreatDirIfNotHad(PATHFINAL)
	CreatDirIfNotHad(projectDir)

	localname := projectDir + "/" + "meta.md"
	urlfilefd, err := os.OpenFile(localname,os.O_CREATE | os.O_TRUNC | os.O_WRONLY, 0600)
	if err != nil{
		fmt.Println("write url file failed", err)
	}else{
		urlfilefd.WriteString("### " + projectName + "\n")
		urlfilefd.WriteString(url + "\n")
		urlfilefd.Sync()
		urlfilefd.Close()
	}

	fullpath := projectDir + "/" + projectName + filepath.Ext(coverurl)
	fmt.Println(coverurl)
	if nil == SavePicFile(fullpath,coverurl){
		fmt.Println("fullpath", fullpath)
		img,err := walk.NewImageFromFileForDPI(fullpath, 96)
		if err != nil{
			fmt.Println("Load cover image from file failed.")
			return errors.New("Load cover image from file failed.")
		}

		w.cover.SetImage(img)
	}
	return nil
}

func (w *workerwnd)spyPics(path,url string)int{

	fullurl := fmt.Sprintf("%vlist2/", url)
	fmt.Println("Spy pics", fullurl)
	w.rate.SetText("Fetching urls...")

	//读取这个页面的所有信息
	httpWebresult,err :=HttpGet(fullurl)
	//判断是否出错，并打印信息
	if err!=nil{
		fmt.Println("SpyPage err:",err)
		w.rate.SetText(err.Error())
		return -1
	}

	titlestr :=`<title>(.*)&raquo;`
	titleret :=regexp.MustCompile(titlestr)
	titlearray :=titleret.FindStringSubmatch(httpWebresult)
	if len(titlearray) < 2{
		w.rate.SetText("title array error")
		return -1
	}

	title := titlearray[1]
	// creating download located directory
	title = strings.Replace(title, " ", "", -1)
	title = strings.Replace(title, "/", "", -1)
	if strings.HasPrefix(title, "[同人誌H漫畫]"){
		title = title[len("[同人誌H漫畫]"):]
	}
	w.main.SetTitle(title)

	w.title = title
	Picsavpath := PATHFINAL + "/" + title + "/" + DIRIMAGES
	CreatDirIfNotHad(PATHFINAL)
	CreatDirIfNotHad(PATHFINAL + "/" + title)
	CreatDirIfNotHad(Picsavpath)

	var datasrcs = make([]string, 0)
	str :=`<img class="list-img lazyload" src=".+" data-src=".+"`
	ret :=regexp.MustCompile(str)
	urls :=ret.FindAllStringSubmatch(httpWebresult,-1)
	srcstr :=`^<img class="list-img lazyload" src=".+" data-src="(.+)" onerror="`
	srcret :=regexp.MustCompile(srcstr)
	for _,a := range urls{
		datasrc := srcret.FindStringSubmatch(a[0])
		if 2 == len(datasrc){
			datasrcs = append(datasrcs, datasrc[1])
		}
	}

	if len(datasrcs) == 0{
		w.rate.SetText("Error at fetch pics, recv 0 picsurl from regexp")
		return -1
	}

	// check if exsit already files
	notexsitsrcs := delbylocal(Picsavpath, datasrcs)
	if notexsitsrcs == nil{
		w.rate.SetText("no need to download for all pics are already on your disk.")
		return 0
	}

	total := len(notexsitsrcs)
	w.total = total
	info := fmt.Sprintf("%v thread(s) was started to download %v imgs...", threadGen(total), total)
	w.pb.SetRange(0,total)
	w.rate.SetText(info)

	w.autoDownload(Picsavpath,notexsitsrcs)

	return 0
}

func delbylocal(dirpath string, picurls []string)[]string{
	var neededurls = make([]string, 0)
	for _,picurl := range(picurls){
		_, fileName := filepath.Split(picurl)
		localname := dirpath + "/" + fileName
		exsitmark, err :=  PathExists(localname)
		if err != nil{
			fmt.Println("delbylocal caught error:", err)
			continue
		}
		if exsitmark == false {
			neededurls = append(neededurls, picurl)
		}
	}
	return neededurls
}

func (w *workerwnd)UpdateCount(val int){
	if val == -1{
		w.failed += 1
	}else{
		w.succeed += 1
	}
	w.processed += 1
	w.pb.SetValue(w.processed)
	if w.total == w.processed{
		ni, _ :=  walk.NewNotifyIcon(w.main.Form())
		ni.SetVisible(true)
		ni.ShowInfo("Finshed", w.main.Title())
		w.rate.SetText("Compressing...")
		ni.Dispose()

		// compress
		w.Compress()
		w.rate.SetText(fmt.Sprintf("All:%d   Down:%d   Fail:%d", w.total, w.succeed, w.failed))
	}
}

func (w *workerwnd)Compress(){
	srcpath := PATHFINAL + "/" + w.title + "/" + DIRIMAGES + "/"
	zipname := PATHFINAL + "/" + w.title + "/" + w.title + ".zip"
	fmt.Println("srcpath:", srcpath)
	fmt.Println("output:", zipname)
	Zip(srcpath, zipname)
}

func threadGen(total int)int{
	var gates int
	if total <= 15{
		gates = 1
	}else if total <= 30{
		gates = 2
	}else if total <= 80{
		gates = 3
	}else if total <= 140{
		gates = 6
	}else if total <= 180{
		gates = 6
	}else if total <= 209{
		gates = 6
	}else{
		gates = 8
	}
	return gates
}

func (w *workerwnd)autoDownload(dirpath string, picurls []string){
	// 切片计算
	gates := threadGen(len(picurls))
	arry := splitArray(picurls, gates)

	fmt.Println("Temp save path:", dirpath)
	for i, url := range arry {
		go func(workid int, url []string){
			for _, picurl := range url {
				_, fileName := filepath.Split(picurl)
				localname := dirpath + "/" + fileName
				fmt.Println("Pulling >>>", localname, "   [", picurl, "]")
				err :=  SavePicFile(localname, picurl)
				if err != nil{
					fmt.Println("Download error", err)
					w.UpdateCount(-1)
				}else{
					w.UpdateCount(1)
				}
			}
		}(i, url)
	}
}

//写入文件
func SavePicFile(localname, url string)error{
	//读取url的信息
	resp,err := http.Get(url)
	if err!=nil{
		fmt.Println("SavePicfile http get error:", err)
		return err
	}

	if resp.Status == "200 OK"{

	}else if resp.Status == "404 Not Found"{
		switch strings.ToLower(filepath.Ext(url)){
		case ".png":
			localname = localname[:len(localname) - 3] + "jpg"
			url = url[:len(url) - 3] + "jpg"
		case ".jpg":
			localname = localname[:len(localname) - 3] + "png"
			url = url[:len(url) - 3] + "png"
		default:
			fmt.Println("unknwon  file external, error:", resp.Status)
			return errors.New(resp.Status)
		}

		fmt.Println("Repull as >>> ", localname)
		resp,err = http.Get(url)
		if err!=nil{
			return err
		}
		if resp.Status != "200 OK"{
			return errors.New(resp.Status)
		}
	}else{
		return errors.New(resp.Status)
	}

	//保存在本地绝对名字
	f,err :=os.Create(localname)
	if err!=nil{
		return err
	}
	defer f.Close()

	buf :=make([]byte,4096)
	for{
		n,err2 :=resp.Body.Read(buf)
		if n==0{
			break
		}
		if err2!=nil&&err2!=io.EOF{
			err = err2
			return err
		}
		// 图片文件落地本地文件
		f.Write(buf[:n])
	}
	return nil
}

type workerwnd struct{
	 main *walk.Dialog
	 cover *walk.ImageView
	 downinfo *walk.Label
	 rate *walk.Label
	 pb *walk.ProgressBar

	total, processed,succeed, failed int
	updatelock sync.Mutex
	 path, title string
}

type worker struct{
	gui *workerwnd
	downpath,url string
	total, downed int
}

func SpyWork(path, url string){
	w := &worker{
		downpath: path,
		url:url,
	}

	w.gui = workWndCreate(path, url)
	if w.gui == nil{
		return
	}

	w.gui.main.SetFocus()
	w.gui.main.Starting().Attach(func(){
		go w.gui.spyCover(path, url)
		go w.gui.spyPics(path, url)
	})

	go w.gui.main.Run()

	return
}
