package main

import (
	"encoding/json"
	"fmt"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
	"os"
	"strings"
)

type godWndMount struct {
	wnd            *walk.MainWindow
	fullUrl        *walk.TextEdit
	baseUrl, idUrl *walk.TextEdit
	downloadPath   *walk.TextEdit
}

func main() {
	var k godWndMount
	err := MainWindow{
		AssignTo: &k.wnd,
		Title:    "Spider nya",
		Size:     Size{Width: 380, Height: 250},
		Icon:     "icon/zhlogo.png",
		Layout:   VBox{},
		Children: []Widget{
			Composite{
				Layout: Grid{Columns: 3},
				Children: []Widget{
					Label{Text: "local save dir"},
					PushButton{Text: "...", MinSize: Size{Width: 36, Height: 20}, OnClicked: func() {
						downPath := PATHFINAL
						if len(downPath) == 0 {
							walk.MsgBox(nil, "local save dir", "Local save path given is illegal!", walk.MsgBoxOK)
							return
						}
						k.downloadPath.SetText(downPath)
					}},
					TextEdit{MaxSize: Size{Height: 20}, AssignTo: &k.downloadPath},
				},
			},
			PushButton{Text: "open local folder", OnClicked: func() {
				openExplorerFolder(k.downloadPath.Text())
			}},
			Composite{
				Layout: Grid{Columns: 2},
				Children: []Widget{
					Label{Text: "Url"},
					TextEdit{AssignTo: &k.fullUrl, MinSize: Size{Height: 80}},
				},
			},
			PushButton{
				Text:    "Pull manga",
				MaxSize: Size{Width: 200, Height: 36},
				MinSize: Size{Width: 200, Height: 36},
				OnClicked: func() {
					if len(k.fullUrl.Text()) > 0 {
						go pullMangaProject(k.downloadPath.Text(), k.fullUrl.Text())
					}
				},
			},
		},
	}.Create()
	if err != nil {
		walk.MsgBox(nil, "Main Window Create Error", err.Error(), walk.MsgBoxOK)
		return
	}

	k.loadConfig()
	k.wnd.Disposing().Attach(
		func() {
			k.saveConfig()
		})

	walk.Clipboard().ContentsChanged().Attach(func() {
		clipstr, _ := walk.Clipboard().Text()
		if len(clipstr) > 0 {
			if strings.HasPrefix(strings.ToLower(clipstr), "http") {
				if strings.Contains(strings.ToLower(clipstr), `://zh`) {
					if strings.Contains(strings.ToLower(clipstr), `hentai.`) {
						win.SetForegroundWindow(k.wnd.Handle())
						askAutoPullManga(k.wnd.Form(), k.downloadPath.Text(), clipstr)
					} else if strings.Contains(strings.ToLower(clipstr), `bus.`) {
						win.SetForegroundWindow(k.wnd.Handle())
						askAutoPullManga(k.wnd.Form(), k.downloadPath.Text(), clipstr)
					}
				}
			}
		}
	})

	k.wnd.Run()

	return
}

func askAutoPullManga(parent walk.Form, localSavePath, mangaUrl string) {
	if win.IDYES == walk.MsgBox(parent, "Downloading?", mangaUrl, walk.MsgBoxYesNo) {
		go pullMangaProject(localSavePath, mangaUrl)
	}
}

type spiderConf struct {
	DownloadPath string
	FullUrl      string
}

func (k *godWndMount) loadConfig() {
	fd, _ := os.Open("conf.json")
	if fd != nil {
		defer fd.Close()
		decoder := json.NewDecoder(fd)

		conf := spiderConf{}
		err := decoder.Decode(&conf)
		if err != nil {

		}
		k.downloadPath.SetText(conf.DownloadPath)
		k.fullUrl.SetText(conf.FullUrl)
	}
}

func (k *godWndMount) saveConfig() {
	fd, err := os.OpenFile("conf.json", os.O_WRONLY|os.O_CREATE, 0666)
	if fd != nil {
		defer fd.Close()

		s := &spiderConf{
			DownloadPath: k.downloadPath.Text(),
			FullUrl:      k.fullUrl.Text(),
		}
		fmt.Println(s)
		buf, _ := json.Marshal(s)
		fd.Write(buf)
	} else {
		fmt.Println("Config file open failed ", err.Error())
	}
}
