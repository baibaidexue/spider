package main

import (
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type settingPage struct {
	*walk.Composite
	enableClipBoardCheck *walk.CheckBox
	enableAutoCompress   *walk.CheckBox
}

func newSettingPage(parent walk.Container, mwc *AppMainWindow) (Page, error) {
	defer timeCost(time.Now(), runFuncName())
	o := new(settingPage)
	err := Composite{
		AssignTo: &o.Composite,
		Layout:   VBox{Alignment: AlignHNearVFar},
		Children: []Widget{
			PushButton{
				Text: "Open download folder",
				// Font:    Font{Family: "Sogoe UI", Bold: false, PointSize: 12},
				MinSize: Size{Width: 200, Height: 60},
				OnClicked: func() {
					openExplorerFolder(MangaSrcDir)
				},
			},
			Composite{
				Layout: Grid{Columns: 2, Alignment: AlignHNearVNear},
				Children: []Widget{
					Label{
						Text:        "Enable clipboard listen",
						Font:        Font{Family: "Sogoe UI", Bold: false, PointSize: 12},
						ToolTipText: "Check if clipboard has a downloadable url",
					},
					CheckBox{
						AssignTo:    &o.enableClipBoardCheck,
						ToolTipText: "Check if clipboard has a downloadable url",
						Checked:     mwc.g.EnableClipBoardListen,
						OnCheckedChanged: func() {
							mwc.g.EnableClipBoardListen = o.enableClipBoardCheck.Checked()
						},
					},
					Label{
						Text:        "Enable auto compress",
						Font:        Font{Family: "Sogoe UI", Bold: false, PointSize: 12},
						ToolTipText: "while download complete start image compress to zip",
					},
					CheckBox{
						AssignTo:    &o.enableAutoCompress,
						ToolTipText: "while download complete start image compress to zip",
						Checked:     mwc.g.EnableAutoCompress,
						OnCheckedChanged: func() {
							mwc.g.EnableAutoCompress = o.enableAutoCompress.Checked()
						},
					},
				},
			},

			PushButton{
				Text:    "Save config",
				MinSize: Size{Width: 200, Height: 60},
				OnClicked: func() {
					mwc.saveConfig()
					NotifyBarCall("Config", "Config saved ok.", mwc.Form())
				},
			},
			VSpacer{},
		},
	}.Create(NewBuilder(parent))
	if err != nil {
		walk.MsgBox(nil, "Spider page Create Error", err.Error(), walk.MsgBoxOK)
		return nil, err
	}

	if err := walk.InitWrapperWindow(o); err != nil {
		return nil, err
	}

	return o, nil
}
