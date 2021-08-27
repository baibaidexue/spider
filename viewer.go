package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/lxn/walk"
	// "github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func getMangaIcons(path string) ([]string, error) {
	var list []string
	rd, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, fi := range rd {
		if fi.IsDir() {
			filename := filepath.Join(path, fi.Name(), fi.Name()+".jpg")
			exist, err := PathExists(filename)
			if err != nil {
				continue
			}
			if exist {
				// fmt.Printf("findout:%v\n", filename)
				list = append(list, filename)
			}
		}
	}
	return list, nil
}
func callImage(item string) {
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
	}
	if len(imagePath) > 0 {
		openExplorerFolder(imagePath)
	}
}

func widgetCompose(item string) Widget {
	return Composite{
		Layout: VBox{},
		Children: []Widget{
			ImageView{
				// AssignTo:
				Name:    item,
				MaxSize: Size{210, 300},
				MinSize: Size{210, 300},
				Image:   ".\\" + item,
				Margin:  2,
				Mode:    ImageViewModeZoom,
				OnMouseDown: func(x, y int, button walk.MouseButton) {
					callImage(item)
				},
				ToolTipText: filepath.Base(item),
			},
			Label{
				Text:        filepath.Base(item),
				Font:        Font{Bold: true, PointSize: 12},
				MaxSize:     Size{210, 30},
				ToolTipText: filepath.Base(item),
			},
		},
	}
}

func thumbnailListWidgetsInit(list []string) []Widget {
	var widgets []Widget
	for _, item := range list {
		widgets = append(widgets, widgetCompose(item))
	}
	return widgets
}

func viewer(parent walk.Form, downloadPath string) {

	go func() {
		var handler *walk.Dialog

		thumbnailList, err := getMangaIcons(downloadPath)
		if err != nil {
			log.Printf("Viewer get manga's icon failed:%v", err)
			return
		}
		widgets := thumbnailListWidgetsInit(thumbnailList[:15])
		if len(widgets) == 0 {
			log.Println("No thumbnail can be found!")
			return
		}
		err = Dialog{
			AssignTo: &handler,
			Title:    "MangaView",
			// Size:     Size{400, 600},
			MinSize: Size{900, 800},
			Layout:  VBox{},
			Children: []Widget{
				ScrollView{
					Layout:   Flow{},
					Children: widgets},
			},
		}.Create(parent)
		if err != nil {
			log.Println(err)
		}

		handler.Run()
	}()
}
