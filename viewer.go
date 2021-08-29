package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/lxn/walk"
	"github.com/pkg/errors"

	// "github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func sortByTime(pl []os.FileInfo) []os.FileInfo {
	sort.Slice(pl, func(i, j int) bool {
		flag := false
		if pl[i].ModTime().After(pl[j].ModTime()) {
			flag = true
		} else if pl[i].ModTime().Equal(pl[j].ModTime()) {
			if pl[i].Name() < pl[j].Name() {
				flag = true
			}
		}
		return flag
	})
	return pl
}

func genMangaInfo(path string) (imgcount int, mangaSize int64) {
	rd, err := ioutil.ReadDir(path)
	if err != nil {
		return 0, 0
	}
	for _, fi := range rd {
		if fi.IsDir() {
			dirimg, dirsize := genMangaInfo(filepath.Join(path, fi.Name()))
			imgcount += dirimg
			mangaSize += dirsize
		}

		if strings.HasSuffix(fi.Name(), ".jpg") {
			imgcount++
		}
		if strings.HasSuffix(fi.Name(), ".png") {
			imgcount++
		}
		mangaSize += fi.Size()
	}
	return
}

func getMangaLists(path string) ([]string, error) {
	var list []string
	rd, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	readerInfos1 := sortByTime(rd)
	ext := []string{".jpg", ".png"}
	for _, fi := range readerInfos1 {
		if fi.IsDir() {
			for _, dname := range ext {
				filename := filepath.Join(path, fi.Name(), fi.Name()+dname)
				exist, _ := PathExists(filename)
				if exist {
					list = append(list, filename)
					break
				}
			}
		}
	}
	return list, nil
}
func openManga(item string) {
	fmt.Println(item)
	dir := filepath.Dir(item)
	first := filepath.Join(dir, "images")
	fmt.Println(first)
	var imagePath string
	for i := 1; i < 20; i++ {
		if ok, _ := PathExists(filepath.Join(first, fmt.Sprintf("%d.jpg", i))); ok {
			imagePath = filepath.Join(first, fmt.Sprintf("%d.jpg", i))
			break
		}
		if ok, _ := PathExists(filepath.Join(first, fmt.Sprintf("%d.png", i))); ok {
			imagePath = filepath.Join(first, fmt.Sprintf("%d.png", i))
			break
		}
	}
	if len(imagePath) > 0 {
		openExplorerFolder(imagePath)
	}
}
func (o *viewerPage) mangaLiked(item string) {
	for _, val := range o.TopContext.g.LikedMangas {
		if val == item {
			return
		}
	}
	o.TopContext.g.LikedMangas = append([]string{item}, o.TopContext.g.LikedMangas...)
}

func removeArrayItem(list []string, item string) (mangalist []string) {
	listlen := len(list)
	for i := 0; i < listlen; i++ {
		if list[i] != item {
			mangalist = append(mangalist, list[i])
		}
	}
	return
}

func (o *viewerPage) mangaUnLiked(item string) {
	o.TopContext.g.LikedMangas = removeArrayItem(o.TopContext.g.LikedMangas, item)
}

func (o *viewerPage) isLikedManga(item string) bool {
	for _, val := range o.TopContext.g.LikedMangas {
		if val == item {
			return true
		}
	}
	return false
}

const MangaLikedPrefix = "❤️"

func mangaNameAddLikedPrefix(ori string) string {
	return MangaLikedPrefix + ori
}
func mangaNameRemoveLikedPrefix(ori string) string {
	return ori[len(MangaLikedPrefix):]
}

func getUrlFromMeta(path string) (url string) {
	fmt.Println(path)
	metafile := filepath.Join(path, "meta.md")
	if ok, _ := PathExists(metafile); ok {

		fd, err := os.OpenFile(metafile, os.O_RDONLY, 0666)
		if err != nil {
			return
		}
		defer fd.Close()
		br := bufio.NewReader(fd)
		for {
			a, _, c := br.ReadLine()
			if c == io.EOF {
				break
			}
			linebuf := string(a)
			if strings.HasPrefix(linebuf, "http") {
				url = linebuf
				return
			}
		}
	}
	return
}

func (o *viewerPage) widgetCompose(item string) Widget {
	picname := filepath.Base(item)
	mangaName := picname[:len(picname)-4]
	mangaNameFinal := mangaName

	// Style change
	font := Font{Family: "Sogoe UI", Bold: false, PointSize: 11}
	if o.viewMode == VIEWMODE_ALL && o.isLikedManga(item) {
		mangaNameFinal = mangaNameAddLikedPrefix(mangaNameFinal)
		font = Font{Family: "Sogoe UI", Bold: true, PointSize: 11}
	}

	count, size := genMangaInfo(filepath.Dir(item))
	mangaInfo := fmt.Sprintf("Images:%v      MangaSize: %v MB", count-1, (size/1024)/1024)

	var unlikeStatus *walk.Action
	var likeStatus *walk.Action
	var mangaTitle *walk.TextLabel
	var imageField *walk.ImageView
	return Composite{
		Layout: VBox{},
		Children: []Widget{
			ImageView{
				AssignTo: &imageField,
				Name:     mangaName,
				MaxSize:  Size{Width: 190, Height: 280},
				MinSize:  Size{Width: 190, Height: 280},
				Image:    ".\\" + item,
				Margin:   0,
				Mode:     ImageViewModeZoom,
				ContextMenuItems: []MenuItem{
					Action{Text: "Read", OnTriggered: func() {
						openManga(item)
					}},
					Action{Text: "Open In Explorer", OnTriggered: func() {
						openExplorerFolder(filepath.Dir(item))
					}},
					Action{Text: "Copy Manga's Name", OnTriggered: func() {
						_ = walk.Clipboard().SetText(mangaName)
					}},
					Action{Text: "Copy Manga's Url", OnTriggered: func() {
						_ = walk.Clipboard().SetText(getUrlFromMeta(filepath.Dir(item)))
					}},
					Separator{},
					Action{AssignTo: &likeStatus, Checked: o.isLikedManga(item), Text: "Add To Favorite",
						OnTriggered: func() {
							o.mangaLiked(item)
							unlikeStatus.SetEnabled(true)
							_ = likeStatus.SetChecked(true)
							font, err := walk.NewFont("Sogoe UI", 11, walk.FontBold)
							if err == nil {
								mangaTitle.SetFont(font)
							}
							_ = mangaTitle.SetText(mangaName)
						}},
					Action{AssignTo: &unlikeStatus, Enabled: o.isLikedManga(item), Text: "Remove From Favorite", OnTriggered: func() {
						o.mangaUnLiked(item)
						_ = likeStatus.SetChecked(false)
						font, err := walk.NewFont("Sogoe UI", 11, 0)
						if err == nil {
							mangaTitle.SetFont(font)
						}
						_ = mangaTitle.SetText(mangaNameAddLikedPrefix(mangaName))
						unlikeStatus.SetEnabled(false)
					}},
					Separator{},
					Action{Text: "Refresh Status", OnTriggered: func() {
						count, size := genMangaInfo(filepath.Dir(item))
						mangaInfo := fmt.Sprintf("Images:%v      MangaSize: %v MB", count-1, (size/1024)/1024)
						newTips := fmt.Sprintf("%v\n\n\n%v", mangaName, mangaInfo)
						_ = imageField.SetToolTipText(newTips)
					}},
					Action{Text: "Download Again", OnTriggered: func() {
						url := getUrlFromMeta(filepath.Dir(item))
						if len(url) > 0 {
							go pullMangaProject(MangaSrcDir, url, o.TopContext, o.TopContext.g.EnableAutoCompress)
						}
					}},
					Separator{},
					Action{Text: "Delete From Disk", OnTriggered: func() {
						o.deleteManga(item)
					}},
				},
				ToolTipText: fmt.Sprintf("%v\n\n\n%v", mangaName, mangaInfo),
			},
			TextLabel{
				AssignTo:    &mangaTitle,
				Text:        mangaNameFinal,
				Font:        font,
				MaxSize:     Size{Width: 180, Height: 32},
				ToolTipText: mangaName,
			},
		},
	}
}

func (o *viewerPage) deleteMangaItem(item string) {
	if o.isLikedManga(item) {
		o.mangaUnLiked(item)
	}

	fmt.Println(o.datasource[:2])
	o.datasource = removeArrayItem(o.datasource, item)
	fmt.Println("after delete")
	fmt.Println(o.datasource[:2])
}

func (o *viewerPage) deleteManga(item string) {
	defer timeCost(time.Now(), runFuncName())

	mangaTopDir := filepath.Dir(item)
	log.Println("Prepare delete", mangaTopDir)
	o.deleteMangaItem(item)
	err := os.RemoveAll(mangaTopDir)
	if err != nil {
		log.Println(item, " delete failed:", err)
	}

	delete(o.mangaPageMap, o.listoff)
	o.refreshCurPage()
}

func (o *viewerPage) thumbnailListWidgetsInit(list []string) []Widget {
	defer timeCost(time.Now(), runFuncName())
	var widgets []Widget
	for _, item := range list {
		widgets = append(widgets, o.widgetCompose(item))
	}
	return widgets
}

type mangaViewPage struct {
	*walk.Composite
}

const VIEWMODE_ALL = "All"
const VIEWMODE_LIKED = "Liked"

type viewerPage struct {
	TopContext *AppMainWindow
	*walk.Composite

	viewMode   string
	mangaCom   *walk.Composite
	datasource []string

	loadCombox                  *walk.ComboBox
	perLoadCnt                  int
	listoff                     int
	mangaPageMap                map[int]Page
	currentPage                 Page
	currentPageChangedPublisher walk.EventPublisher

	prebut *walk.PushButton
	nexbut *walk.PushButton
	hombut *walk.PushButton
	endbut *walk.PushButton
}

func (o *viewerPage) pageThumbnailListWidgetsInit(head, tail int) []Widget {
	return o.thumbnailListWidgetsInit(o.datasource[head:tail])
}

func (o *viewerPage) getViewRange(off int) (data []string, err error) {
	head := off
	tail := off + o.TopContext.g.MangaPerPage
	log.Printf("head:%v tail:%v  sourcelen:%v", head, tail, len(o.datasource))
	if head < 0 {
		err = errors.New(" head on left of list head")
		return
	}
	if head >= len(o.datasource) {
		err = errors.New("head on right of list tail")
		return
	}

	if tail > len(o.datasource) {
		tail = len(o.datasource)
	}
	data = o.datasource[head:tail]
	return
}

func (o *viewerPage) newMangaReaderPage(parent walk.Container, publishOff int) (Page, error) {
	defer timeCost(time.Now(), runFuncName())
	lists, err := o.getViewRange(publishOff)
	if err != nil {
		return nil, err
	}

	m := new(mangaViewPage)
	if err := (Composite{
		AssignTo: &m.Composite,
		Name:     "MangaPage",
		Layout:   VBox{},

		Children: []Widget{
			ScrollView{
				Layout:   Flow{},
				Children: o.thumbnailListWidgetsInit(lists),
			},
			Label{
				Font: Font{Bold: true, PointSize: 12},
				Text: fmt.Sprintf("Pos: %v/%v  Rendered:%v", publishOff, len(o.datasource), o.TopContext.g.MangaPerPage),
			},
		},
	}).Create(NewBuilder(parent)); err != nil {
		return nil, err
	}

	if err := walk.InitWrapperWindow(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (o *viewerPage) reloadMangaListOfPath(path string) {
	list, err := getMangaLists(path)
	if err != nil {
		log.Printf("Viewer get manga's thumbnails failed: %v", err)
		return
	}
	o.datasource = list
}

func (o *viewerPage) reloadMangaListOfGlobal() {
	o.datasource = o.TopContext.g.LikedMangas
}

func (o *viewerPage) reloadMangaList() {
	defer timeCost(time.Now(), runFuncName())
	switch o.viewMode {
	case VIEWMODE_ALL:
		o.reloadMangaListOfPath(MangaSrcDir)
	case VIEWMODE_LIKED:
		o.reloadMangaListOfGlobal()
	}
}

func (o *viewerPage) switchMangaPage(id int) error {
	if _, ok := o.mangaPageMap[id]; !ok {
		newMangaPage, err := o.newMangaReaderPage(o.mangaCom, id)
		if err != nil {
			return err
		}
		o.mangaPageMap[id] = newMangaPage
	}

	publishPage := o.mangaPageMap[id]
	if o.currentPage == publishPage {
		return nil
	}

	if o.currentPage != nil {
		o.currentPage.SetVisible(false)
	}

	o.currentPage = publishPage
	o.currentPage.SetVisible(true)
	o.currentPage.SetFocus()
	o.currentPageChangedPublisher.Publish()

	return nil
}

func (o *viewerPage) avoidAllPage() {
	// for _, mangapag := range o.mangaPageMap {
	// 	mangapag.Dispose()
	// }
	// gc will clean the old map
	o.mangaPageMap = make(map[int]Page)
	return
}

func (o *viewerPage) trySwitchPage(exp int) {
	err := o.switchMangaPage(exp)
	if err != nil {
		log.Println(err)
		return
	}
	o.listoff = exp
}

func (o *viewerPage) switchFirPage() {
	o.trySwitchPage(0)
}

func (o *viewerPage) switchLastPage() {
	overtime := len(o.datasource) / o.perLoadCnt
	exp := overtime * o.perLoadCnt
	if exp == len(o.datasource) {
		exp -= o.TopContext.g.MangaPerPage
	}
	o.trySwitchPage(exp)
}

func (o *viewerPage) switchPrevPage() {
	exp := o.listoff - o.perLoadCnt
	o.trySwitchPage(exp)
}

func (o *viewerPage) switchNextPage() {
	exp := o.listoff + o.perLoadCnt
	o.trySwitchPage(exp)
}

func (o *viewerPage) refreshCurPage() {
	exp := o.listoff
	o.trySwitchPage(exp)
}

func newViewContainerPage(parent walk.Container, mw *AppMainWindow, viewMode string) (Page, error) {
	o := new(viewerPage)

	o.TopContext = mw
	o.viewMode = viewMode
	o.mangaPageMap = make(map[int]Page)
	o.reloadMangaList()
	o.listoff = 0
	o.perLoadCnt = mw.g.MangaPerPage

	if err := (Composite{
		AssignTo: &o.Composite,
		Name:     "viewerPage",
		Layout:   VBox{},
		Children: []Widget{
			Composite{
				// OnKeyDown: func(key walk.Key) {
				// 	switch key {
				// 	case walk.KeyLeft:
				// 		fallthrough
				// 	case walk.KeyPrior:
				// 		fmt.Println("catch key prev")
				// 		o.prebut.SetChecked(true)
				//
				// 	case walk.KeyRight:
				// 		fallthrough
				// 	case walk.KeyNext:
				// 		fmt.Println("catch key next")
				// 		// o.nexbut.SetChecked(true)
				// 		o.nexbut.WndProc(o.Handle(), win.WM_COMMAND, win.BN_CLICKED, 1)
				// 	case walk.KeyHome:
				// 		fmt.Println("catch key HOME")
				// 		o.switchFirPage()
				// 	case walk.KeyEnd:
				// 		fmt.Println("catch key END")
				// 		o.switchLastPage()
				// 	}
				// },
				Layout: HBox{},
				Children: []Widget{
					ComboBox{
						AssignTo: &o.loadCombox,
						Model:    []string{"5", "10", "15", "20", "25", "30"},

						OnCurrentIndexChanged: func() {
							switch o.loadCombox.CurrentIndex() {
							case 0:
								o.perLoadCnt = 5
							case 1:
								o.perLoadCnt = 10
							case 2:
								o.perLoadCnt = 15
							case 3:
								o.perLoadCnt = 20
							case 4:
								o.perLoadCnt = 25
							case 5:
								o.perLoadCnt = 30
							default:
								o.perLoadCnt = 15
							}
							mw.g.MangaPerPage = o.perLoadCnt
						},
					},
					PushButton{
						Text: "Reload",
						// MinSize: Size{Width: 100, Height: 80},
						OnClicked: func() {
							log.Println("reload manga datasource")
							o.reloadMangaList()
							o.avoidAllPage()
							o.refreshCurPage()
						},
					},
					HSpacer{},
					PushButton{
						Text:        "<<",
						AssignTo:    &o.hombut,
						ToolTipText: "goto first page",
						//MinSize: Size{Width: 100, Height: 80},
						OnClicked: func() {
							o.switchFirPage()
						},
					},
					HSpacer{},
					PushButton{
						Text:        "<",
						AssignTo:    &o.prebut,
						ToolTipText: "previous page",
						//MinSize: Size{Width: 100, Height: 80},
						OnClicked: func() {
							o.switchPrevPage()
						},
					},
					PushButton{
						Text:        ">",
						AssignTo:    &o.nexbut,
						ToolTipText: "next page",
						//MinSize: Size{Width: 100, Height: 80},
						OnClicked: func() {
							o.switchNextPage()
						},
					},

					HSpacer{},

					PushButton{
						Text:        ">>",
						AssignTo:    &o.endbut,
						ToolTipText: "goto end page",
						//MinSize: Size{Width: 100, Height: 80},
						OnClicked: func() {
							o.switchLastPage()
						},
					},
					HSpacer{},
				},
			},
			Composite{
				AssignTo: &o.mangaCom,
				Name:     "mangaCom",
				Layout:   HBox{MarginsZero: true, SpacingZero: true},
			},
		},
	}).Create(NewBuilder(parent)); err != nil {
		return nil, err
	}

	if err := walk.InitWrapperWindow(o); err != nil {
		return nil, err
	}

	o.switchFirPage()
	return o, nil
}

func newViewAllPage(parent walk.Container, mw *AppMainWindow) (Page, error) {
	defer timeCost(time.Now(), runFuncName())
	return newViewContainerPage(parent, mw, VIEWMODE_ALL)
}

func newViewLikedPage(parent walk.Container, mw *AppMainWindow) (Page, error) {
	defer timeCost(time.Now(), runFuncName())
	return newViewContainerPage(parent, mw, VIEWMODE_LIKED)
}
