package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
)

const (
	MangaBoxTitle = "Mangabox v1.5"
	MangaSrcDir   = "star"
	DIRIMAGES     = "images"
)

func main() {
	mw := new(AppMainWindow)
	mw.loadConfig()

	cfg := &MultiPageMainWindowConfig{
		Name:       MangaBoxTitle,
		InitSize:   Size{Width: 1250, Height: 950},
		HeaderIcon: "icon/zhlogo.png",
		Title:      MangaBoxTitle,
		MenuItems: []MenuItem{
			Menu{
				Text: "&Edit",
				Items: []MenuItem{
					Action{
						Text:        "&exit",
						OnTriggered: func() { mw.Close() },
					},
				},
			},
			Menu{
				Text: "&Help",
				Items: []MenuItem{
					Action{
						Text:        "&manual",
						OnTriggered: func() { mw.aboutManualTriggered() },
					},
					Action{
						Text:        "&about data store",
						OnTriggered: func() { mw.aboutDataSaveTriggered() },
					},
					Action{
						Text:        "about &mangabox",
						OnTriggered: func() { mw.aboutVersionTriggered() },
					},
				},
			},
		},
		OnCurrentPageChanged: func() {
			mw.updateTitle(mw.CurrentPageTitle())
		},
		PageCfgs: []PageConfig{
			{"New", "icon/plus.png", newSpiderPage},
			{"Local", "icon/go-home.png", newViewAllPage},
			{"Liked", "icon/emblem-favorite.png", newViewLikedPage},
			{"Setting", "icon/document-properties.png", newSettingPage},
		},
	}

	mpmw, err := CreateMainWindow(cfg, mw)
	if err != nil {
		walk.MsgBox(nil, "error at start lemon", err.Error(), walk.MsgBoxOK)
		panic(err)
	}

	mw.mainWindowConf = mpmw

	walk.Clipboard().ContentsChanged().Attach(func() {
		if mw.g.EnableClipBoardListen == false {
			return
		}
		if clipstr, _ := walk.Clipboard().Text(); len(clipstr) > 0 {
			if canBeMangaUrl(clipstr) {
				bringWindowTop(mpmw.Handle())
				askIfSaveManga(mpmw.Form(), MangaSrcDir, clipstr, mw, mw.g.EnableAutoCompress)
			}
		}
	})

	mw.Disposing().Attach(func() { mw.saveConfig() })
	mw.Run()
}

type AppMainWindow struct {
	*mainWindowConf

	g *globalSetting
}

func (mw *AppMainWindow) aboutVersionTriggered() {
	walk.MsgBox(mw,
		"About MangaBox Ver 1.5",
		"Manga download tool cast by Marine.",
		walk.MsgBoxOK|walk.MsgBoxIconInformation)
}

func (mw *AppMainWindow) aboutManualTriggered() {
	walk.MsgBox(mw,
		"MangaBox Manual",
		"Use my mentor under my guide",
		walk.MsgBoxOK|walk.MsgBoxIconInformation)
}

func (mw *AppMainWindow) aboutDataSaveTriggered() {
	walk.MsgBox(mw,
		"About Data store",
		fmt.Sprintf("User data will save to or load from \"star\" located with \"spider.exe\", named with manga's name."),
		walk.MsgBoxOK|walk.MsgBoxIconInformation)
}

func (mw *AppMainWindow) updateTitle(prefix string) {

	var buf bytes.Buffer
	buf.WriteString(MangaBoxTitle)
	if prefix != "" {
		buf.WriteString(" - ")
		buf.WriteString(prefix)
	}

	mw.SetTitle(buf.String())
}

type globalSetting struct {
	DownloadPath          string
	FullUrl               string
	EnableAutoCompress    bool
	EnableClipBoardListen bool
	MangaPerPage          int

	LikedMangas []string
}

func (k *AppMainWindow) loadConfig() {
	defer timeCost(time.Now(), runFuncName())
	conf := globalSetting{}

	fd, err := os.Open("conf.json")
	if err != nil {
		conf.DownloadPath = MangaSrcDir
		conf.EnableAutoCompress = true
		conf.EnableClipBoardListen = true
		conf.MangaPerPage = 10
		conf.FullUrl = "https://zha.qqhentai.com/g/369120/"
	} else {
		_ = fd.Close()
		decoder := json.NewDecoder(fd)
		err := decoder.Decode(&conf)
		if err != nil {

		}
		// const
		conf.DownloadPath = MangaSrcDir
		if conf.MangaPerPage == 0 {
			conf.MangaPerPage = 10
		}
	}
	k.g = &conf
}

func (k *AppMainWindow) saveConfig() {
	fd, err := os.OpenFile("conf.json", os.O_WRONLY|os.O_CREATE, 0666)
	if err == nil {
		buf, _ := json.Marshal(k.g)
		_, _ = fd.Write(buf)
		_ = fd.Close()
	} else {
		fmt.Println("Config file open failed ", err.Error())
	}
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

func askIfSaveManga(parent walk.Form, localSavePath, mangaUrl string, topContext *AppMainWindow, autoCompress bool) {
	if win.IDYES == walk.MsgBox(parent, "Downloading?", mangaUrl, walk.MsgBoxYesNo) {
		go pullMangaProject(localSavePath, mangaUrl, topContext, autoCompress)
	}
}
