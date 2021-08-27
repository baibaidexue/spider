package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
)

type godWndMount struct {
	wnd                  *walk.MainWindow
	fullUrl              *walk.TextEdit
	downloadPath         *walk.TextEdit
	enableClipBoardCheck *walk.CheckBox
	enableAutoCompress   *walk.CheckBox
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
					Label{Text: "download path"},
					PushButton{Text: "...", Enabled: false, MinSize: Size{Width: 36, Height: 20}, OnClicked: func() {
						downPath := PATHFINAL
						_ = k.downloadPath.SetText(downPath)
					}},
					TextEdit{MaxSize: Size{Height: 20}, Enabled: false, AssignTo: &k.downloadPath},
				},
			},
			PushButton{
				Text:    "open download folder",
				MinSize: Size{Width: 200, Height: 30},
				OnClicked: func() {
					openExplorerFolder(k.downloadPath.Text())
				}},
			Composite{
				Layout: Grid{Columns: 1},
				Children: []Widget{
					Label{
						Text:    "",
						MinSize: Size{Height: 10},
					},
				},
			},
			Composite{
				Layout: Grid{Columns: 2},
				Children: []Widget{
					Label{
						Text: "Enable clipboard listen",
						// MinSize: Size{Height: 10},
						ToolTipText: "Check if clipboard has a downloadable url",
					},
					CheckBox{
						AssignTo:    &k.enableClipBoardCheck,
						ToolTipText: "Check if clipboard has a downloadable url",
					},
					Label{
						Text: "Enable auto compress",
						// MinSize: Size{Height: 10},
						ToolTipText: "while download complete start image compress to zip",
					},
					CheckBox{
						AssignTo:    &k.enableAutoCompress,
						ToolTipText: "while download complete start image compress to zip",
					},
				},
			},

			Composite{
				Layout: Grid{Columns: 1},
				Children: []Widget{
					Label{
						Text:    "",
						MinSize: Size{Height: 10},
					},
				},
			},
			Composite{
				Layout: Grid{Columns: 2},
				Children: []Widget{
					TextEdit{
						AssignTo:    &k.fullUrl,
						MinSize:     Size{Height: 60},
						Font:        Font{PointSize: 12, Bold: false},
						ToolTipText: "one url once ",
					},
				},
			},
			PushButton{
				Text:    "Download",
				Font:    Font{PointSize: 14},
				MaxSize: Size{Width: 200, Height: 60},
				MinSize: Size{Width: 200, Height: 40},
				OnClicked: func() {
					if len(k.fullUrl.Text()) > 0 {
						go pullMangaProject(k.downloadPath.Text(), k.fullUrl.Text(), k.enableAutoCompress.Checked())
					}
				},
			},
			PushButton{
				Text:    "Manga Viewer",
				Font:    Font{PointSize: 14},
				MaxSize: Size{Width: 200, Height: 60},
				MinSize: Size{Width: 200, Height: 40},
				OnClicked: func() {
					viewer(k.wnd.Form(), k.downloadPath.Text())
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
		if k.enableClipBoardCheck.Checked() == false {
			return
		}
		if clipstr, _ := walk.Clipboard().Text(); len(clipstr) > 0 {
			if canBeMangaUrl(clipstr) {
				bringWindowTop(k.wnd.Handle())
				askIfSaveManga(k.wnd.Form(), k.downloadPath.Text(), clipstr, k.enableAutoCompress.Checked())
			}
		}
	})

	k.wnd.Run()
	return
}

func bringWindowTop(hwnd win.HWND) {
	win.SetWindowPos(hwnd, win.HWND_TOPMOST, 0, 0, 0, 0, win.SWP_NOMOVE|win.SWP_NOSIZE)
	win.SetWindowPos(hwnd, win.HWND_NOTOPMOST, 0, 0, 0, 0, win.SWP_SHOWWINDOW|win.SWP_NOSIZE|win.SWP_NOMOVE)
	win.SetForegroundWindow(hwnd)
	win.SetFocus(hwnd)
	win.SetActiveWindow(hwnd)
}

func canBeMangaUrl(clipstr string) bool {
	lowerStr := strings.ToLower(clipstr)
	if strings.HasPrefix(lowerStr, "http") &&
		strings.Contains(lowerStr, `://`) {
		if strings.Contains(lowerStr, `hentai.com`) ||
			strings.Contains(lowerStr, `hentai.site`) {
			return true
		}
	}
	return false
}

func askIfSaveManga(parent walk.Form, localSavePath, mangaUrl string, autoCompress bool) {
	if win.IDYES == walk.MsgBox(parent, "Downloading?", mangaUrl, walk.MsgBoxYesNo) {
		go pullMangaProject(localSavePath, mangaUrl, autoCompress)
	}
}

type spiderConf struct {
	DownloadPath          string
	FullUrl               string
	EnableAutoCompress    bool
	EnableClipBoardListen bool
}

func (k *godWndMount) loadConfig() {
	fd, _ := os.Open("conf.json")
	if fd != nil {
		decoder := json.NewDecoder(fd)
		conf := spiderConf{}
		err := decoder.Decode(&conf)
		if err != nil {

		}
		_ = fd.Close()
		// const
		conf.DownloadPath = "star"

		_ = k.downloadPath.SetText(conf.DownloadPath)
		_ = k.fullUrl.SetText(conf.FullUrl)
		k.enableAutoCompress.SetChecked(conf.EnableAutoCompress)
		k.enableClipBoardCheck.SetChecked(conf.EnableClipBoardListen)
	}
}

func (k *godWndMount) saveConfig() {
	fd, err := os.OpenFile("conf.json", os.O_WRONLY|os.O_CREATE, 0666)
	if err == nil {

		s := &spiderConf{
			DownloadPath:          k.downloadPath.Text(),
			FullUrl:               k.fullUrl.Text(),
			EnableAutoCompress:    k.enableAutoCompress.Checked(),
			EnableClipBoardListen: k.enableClipBoardCheck.Checked(),
		}
		fmt.Println(s)
		buf, _ := json.Marshal(s)
		_, _ = fd.Write(buf)
		_ = fd.Close()
	} else {
		fmt.Println("Config file open failed ", err.Error())
	}
}
