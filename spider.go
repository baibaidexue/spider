package main

import (
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type spiderPage struct {
	*walk.Composite
	fullUrl *walk.TextEdit
}

func newSpiderPage(parent walk.Container, mw *AppMainWindow) (Page, error) {
	defer timeCost(time.Now(), runFuncName())
	o := new(spiderPage)

	err := Composite{
		AssignTo: &o.Composite,
		Layout:   VBox{},
		Children: []Widget{
			PushButton{
				Text:    "open download folder",
				MinSize: Size{Width: 200, Height: 30},
				OnClicked: func() {
					openExplorerFolder(MangaSrcDir)
				},
			},
			TextEdit{
				AssignTo:    &o.fullUrl,
				MinSize:     Size{Height: 60},
				Font:        Font{PointSize: 14, Bold: false},
				ToolTipText: "one url once ",
				Text:        mw.g.FullUrl,
				OnTextChanged: func() {
					mw.g.FullUrl = o.fullUrl.Text()
				},
			},

			PushButton{
				Text:    "Download",
				Font:    Font{PointSize: 15, Bold: true},
				MaxSize: Size{Width: 200, Height: 100},
				MinSize: Size{Width: 200, Height: 80},
				OnClicked: func() {
					if len(o.fullUrl.Text()) > 0 {
						go pullMangaProject(MangaSrcDir, o.fullUrl.Text(), mw, mw.g.EnableAutoCompress)
					}
				},
			},
		},
	}.Create(NewBuilder(parent))
	if err != nil {
		walk.MsgBox(nil, "Spider page Create Error", err.Error(), walk.MsgBoxOK)
		return nil, err
	}

	_ = o.fullUrl.SetText(mw.g.FullUrl)

	if err := walk.InitWrapperWindow(o); err != nil {
		return nil, err
	}

	return o, nil
}
