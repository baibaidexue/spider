package main

import (
	"encoding/json"
	"fmt"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"os"
	"runtime"
	"strings"
)

type godWndMount struct {
	wnd *walk.MainWindow
	fullUrl *walk.TextEdit
	baseUrl, idUrl *walk.TextEdit
	downloadPath *walk.TextEdit
}

func main() {
	var k godWndMount

	runtime.GOMAXPROCS(runtime.NumCPU()/2)

	err := MainWindow{
		AssignTo: &k.wnd,
		Title:	"Spider of nya",
		Size: Size{Width: 380, Height: 250},
		Layout: VBox{},
		Children: []Widget{

			Composite{
				Layout: Grid{Columns: 2},
				Children:[]Widget{
					Label{Text: "BaseUrl"},
					TextEdit{MaxSize: Size{Height: 20}, AssignTo: &k.baseUrl},
					Label{Text: "mangaid"},
					TextEdit{MaxSize: Size{Height: 20}, AssignTo: &k.idUrl},
				},
			},
			Composite{
				Layout: Grid{Columns: 3},
				Children:[]Widget{
					Label{Text: "downloaddir"},
					PushButton{Text: "...", MinSize: Size{Width: 36, Height: 20}, OnClicked: func(){
						downPath := PATHDOWNLOAD
						if downPath == ""{
							walk.MsgBox(nil, "path", "Path find error", walk.MsgBoxOK)
						}
						k.downloadPath.SetText(downPath)
					}},
					TextEdit{MaxSize: Size{Height: 20}, AssignTo: &k.downloadPath},
				},
			},
			Composite{
				Layout: Grid{Columns: 2},
				Children:[]Widget{
					Label{Text: "Url"},
					TextEdit{AssignTo: &k.fullUrl, MinSize: Size{Height: 80}},
				},
			},
			PushButton{
					Text: "kiss",
					MaxSize: Size{Width: 200, Height: 36},
					MinSize: Size{Width: 200, Height: 36},
					OnClicked: func() {
						if len(k.fullUrl.Text())  > 0 {
							go SpyWork(k.downloadPath.Text(), k.fullUrl.Text())
						}else{
							if len(k.idUrl.Text()) == 0{
								walk.MsgBox(nil, "params", "Id needed", walk.MsgBoxOK)
								return
							}
							if len(k.baseUrl.Text()) == 0{
								walk.MsgBox(nil, "params", "BaseUrl needed", walk.MsgBoxOK)
								return
							}
							if len(k.downloadPath.Text()) == 0{
								walk.MsgBox(nil, "params", "Download path needed", walk.MsgBoxOK)
								return
							}
							url :=fmt.Sprintf("%v/%v/", k.baseUrl.Text(), k.idUrl.Text())
							go SpyWork(k.downloadPath.Text(), url)
						}
				},
			},
		},

	}.Create()
	if err != nil{
		walk.MsgBox(nil, "Main Fatal Error", err.Error(), walk.MsgBoxOK)
		return
	}

	k.loadConfig()
	k.wnd.Disposing().Attach(
		func(){
			k.saveConfig()
		})

	walk.Clipboard().ContentsChanged().Attach(func(){
		clipstr,_ := walk.Clipboard().Text()
		if len(clipstr) > 0{
			if strings.HasPrefix(strings.ToLower(clipstr), "http"){
				if strings.Contains(strings.ToLower(clipstr), `://zh`){
					if strings.Contains(strings.ToLower(clipstr), `hentai.`){
						// k.wnd.BringToTopMost()
						k.wnd.SetFocus()
						autoFetch(k.wnd.Form(),k.downloadPath.Text(), clipstr)
					}else if strings.Contains(strings.ToLower(clipstr), `bus.`){
						// k.wnd.BringToTopMost()
						k.wnd.SetFocus()
						autoFetch(k.wnd.Form(), k.downloadPath.Text(), clipstr)
					}
				}
			}
		}
	})

	k.wnd.Run()

	return
}

func autoFetch(parent walk.Form, downloadpath, tarurl string){
	if 6 == walk.MsgBox(parent, "checking", "Downloading?", walk.MsgBoxYesNo){
		// if walk.MsgBoxYesNo()
		go SpyWork(downloadpath, tarurl)
	}

	return
}

type spiderConf struct{
	BaseUrl		string
	MangaId 	string
	DownloadPath	string
	FullUrl		string
}

func (k *godWndMount)loadConfig(){
	fd, _ := os.Open("./conf.json")
	if fd != nil{
		defer fd.Close()
		decoder := json.NewDecoder(fd)

		conf := spiderConf{}
		err := decoder.Decode(&conf)
		if err != nil{

		}
		k.downloadPath.SetText(conf.DownloadPath)
		k.baseUrl.SetText(conf.BaseUrl)
		k.idUrl.SetText(conf.MangaId)
		k.fullUrl.SetText(conf.FullUrl)
	}
}

func (k *godWndMount)saveConfig(){
	fd, _ := os.OpenFile("./conf.json", os.O_WRONLY | os.O_CREATE, 0666)
	if fd != nil{
		defer fd.Close()

		s := &spiderConf{
			BaseUrl: k.baseUrl.Text(),
			MangaId: k.idUrl.Text(),
			DownloadPath: k.downloadPath.Text(),
			FullUrl: k.fullUrl.Text(),
		}
		fmt.Println(s)
		buf, _ := json.Marshal(s)
		fd.Write(buf)
	}else{
		fmt.Println("File open failed at saving data")
	}
}

func localPath()string{
	return PATHDOWNLOAD
}
