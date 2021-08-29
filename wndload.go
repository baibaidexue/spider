// Copyright 2017 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type MultiPageMainWindowConfig struct {
	Name                 string
	Enabled              Property
	Visible              Property
	HeaderIcon           Property
	Font                 Font
	InitSize             Size
	MinSize              Size
	MaxSize              Size
	ContextMenuItems     []MenuItem
	OnKeyDown            walk.KeyEventHandler
	OnKeyPress           walk.KeyEventHandler
	OnKeyUp              walk.KeyEventHandler
	OnMouseDown          walk.MouseEventHandler
	OnMouseMove          walk.MouseEventHandler
	OnMouseUp            walk.MouseEventHandler
	OnSizeChanged        walk.EventHandler
	OnCurrentPageChanged walk.EventHandler
	Title                string
	Size                 Size
	MenuItems            []MenuItem
	ToolBar              ToolBar
	PageCfgs             []PageConfig
}

type PageConfig struct {
	Title   string
	Image   string
	NewPage NewPageFunc
}

type NewPageFunc func(parent walk.Container, mw *AppMainWindow) (Page, error)

type Page interface {
	// Provided by Walk
	walk.Container
	Parent() walk.Container
	SetParent(parent walk.Container) error
}

type mainWindowConf struct {
	topContext *AppMainWindow

	*walk.MainWindow
	navTB                       *walk.ToolBar
	pageCom                     *walk.Composite
	action2NewPage              map[*walk.Action]NewPageFunc
	actionPageMap               map[*walk.Action]Page
	pageActions                 []*walk.Action
	currentAction               *walk.Action
	currentPage                 Page
	currentPageChangedPublisher walk.EventPublisher
	closehandler                walk.CloseEvent
}

func CreateMainWindow(cfg *MultiPageMainWindowConfig, mw *AppMainWindow) (*mainWindowConf, error) {
	mpmw := &mainWindowConf{
		topContext:     mw,
		action2NewPage: make(map[*walk.Action]NewPageFunc),
		actionPageMap:  make(map[*walk.Action]Page),
	}

	if err := (MainWindow{
		AssignTo:         &mpmw.MainWindow,
		Name:             cfg.Name,
		Title:            cfg.Title,
		Icon:             cfg.HeaderIcon,
		Enabled:          cfg.Enabled,
		Visible:          cfg.Visible,
		Font:             cfg.Font,
		Size:             cfg.InitSize,
		MinSize:          cfg.MinSize,
		MaxSize:          cfg.MaxSize,
		MenuItems:        cfg.MenuItems,
		ToolBar:          cfg.ToolBar,
		ContextMenuItems: cfg.ContextMenuItems,
		OnKeyDown:        cfg.OnKeyDown,
		OnKeyPress:       cfg.OnKeyPress,
		OnKeyUp:          cfg.OnKeyUp,
		OnMouseDown:      cfg.OnMouseDown,
		OnMouseMove:      cfg.OnMouseMove,
		OnMouseUp:        cfg.OnMouseUp,
		OnSizeChanged:    cfg.OnSizeChanged,
		Layout:           HBox{MarginsZero: true, SpacingZero: true},
		Children: []Widget{
			ScrollView{
				HorizontalFixed: true,
				Layout:          VBox{MarginsZero: true},
				MaxSize:         Size{Width: 60},
				Children: []Widget{
					Composite{
						Layout: VBox{MarginsZero: true},
						Children: []Widget{
							ToolBar{
								AssignTo:           &mpmw.navTB,
								Row:                1,
								MaxSize:            Size{Width: 60},
								AlwaysConsumeSpace: true,
								Orientation:        Vertical,
								ButtonStyle:        ToolBarButtonImageAboveText,
							},
						},
					},
				},
			},
			Composite{
				AssignTo: &mpmw.pageCom,
				Name:     "pageCom",
				Layout:   HBox{MarginsZero: true, SpacingZero: true},
			},
		},
	}).Create(); err != nil {
		return nil, err
	}

	succeeded := false
	defer func() {
		if !succeeded {
			mpmw.Dispose()
		}
	}()

	// 根据预设页面参数创建action对象
	for _, pc := range cfg.PageCfgs {
		fmt.Println(pc.Title)
		action, err := mpmw.newPageAction(pc.Title, pc.Image, pc.NewPage)
		if err != nil {
			return nil, err
		}

		mpmw.pageActions = append(mpmw.pageActions, action)
	}

	// 将上面设置的 pageActions 设置到toolbar的变量的 action列表中
	if err := mpmw.setNavToolBar(); err != nil {
		return nil, err
	}

	// 初始化时设置第一个页面显示
	if len(mpmw.pageActions) > 0 {
		// TODO 修改回初始值 0
		if err := mpmw.switchPage(mpmw.pageActions[1]); err != nil {
			return nil, err
		}
	}

	// 当动作触发翻页时，将自定义的函数添加到这个事件的触发处理函数中
	if cfg.OnCurrentPageChanged != nil {
		mpmw.CurrentPageChanged().Attach(cfg.OnCurrentPageChanged)
	}

	// // 当窗体关闭动作触发时，执行下面函数中的动作
	// mpmw.Disposing().Attach(func() {
	// 	ret := RunCheckDataSaveWnd(mpmw.Form())
	// 	if ret {
	// 		mpmw.SavePageData()
	// 	}
	// })

	succeeded = true

	return mpmw, nil
}

// func (mw *mainWindowConf) SavePageData() {
// 	for _, curAct := range mw.pageActions {
// 		curPage := mw.actionPageMap[curAct]
// 		curPage.SaveDataLocal()
// 	}
// 	NotifyBarCall("Saved succeed.", "All configured data saved. Exit ok. ", mw)
// }

func RunCheckDataSaveWnd(owner walk.Form) bool {
	var dlg *walk.Dialog
	var acceptPB, cancelPB *walk.PushButton
	var ret bool

	_, _ = Dialog{
		AssignTo:      &dlg,
		Title:         "Data save check",
		DefaultButton: &acceptPB,
		CancelButton:  &cancelPB,
		Icon:          "button/save.png",
		Layout:        VBox{},
		MinSize:       Size{Width: 100, Height: 80},
		Children: []Widget{
			Composite{
				Layout: HBox{},
				Children: []Widget{
					// Composite{
					// 	Layout: HBox{},
					// 	Children: []Widget{
					// 		Label{
					// 			Text: "Saving data"
					// 		},
					// 	},
					// },
					Composite{
						Layout: HBox{},
						Children: []Widget{
							// HSpacer{},
							PushButton{
								AssignTo: &acceptPB,
								Text:     "Save data",
								OnClicked: func() {
									dlg.Accept()
									ret = true
								},
							},
							PushButton{
								AssignTo: &cancelPB,
								Text:     "Don't save",
								OnClicked: func() {
									dlg.Cancel()
									ret = false
								},
							},
						},
					},
				},
			},
		},
	}.Run(owner)

	return ret
}

func (mw *mainWindowConf) CurrentPage() Page {
	return mw.currentPage
}

func (mw *mainWindowConf) CurrentPageTitle() string {
	if mw.currentAction == nil {
		return ""
	}

	return mw.currentAction.Text()
}

func (mw *mainWindowConf) CurrentPageChanged() *walk.Event {
	return mw.currentPageChangedPublisher.Event()
}

func (mw *mainWindowConf) newPageAction(title, image string, newPage NewPageFunc) (*walk.Action, error) {
	img, err := walk.ImageFrom(image)
	if err != nil {
		return nil, err
	}

	action := walk.NewAction()

	_ = action.SetCheckable(true)
	_ = action.SetExclusive(true)
	_ = action.SetImage(img)
	_ = action.SetText(title)

	mw.action2NewPage[action] = newPage

	Page, err := newPage(mw.pageCom, mw.topContext)
	if err != nil {
		return nil, err
	}
	mw.actionPageMap[action] = Page
	Page.SetVisible(false)

	// 绑定 action 的触发处理事件
	action.Triggered().Attach(func() {
		// mpmw.setCurrentAction(action)
		_ = mw.switchPage(action)
	})

	return action, nil
}

func (mw *mainWindowConf) switchPage(action *walk.Action) error {
	_ = mw.SetFocus()

	prevPage := mw.actionPageMap[mw.currentAction]
	if prevPage != nil {
		prevPage.SetVisible(false)
	}

	// 标识接下来的函数动作不再需要处理了吗？
	action.SetChecked(true)

	// 将当前的显示区域的页面更新为新的页面元素，并且调用渲染函数重新渲染界面
	mw.currentPage = mw.actionPageMap[action]
	mw.currentPage.SetVisible(true)
	mw.currentAction = action
	mw.currentPageChangedPublisher.Publish()

	return nil
}

func (mw *mainWindowConf) setCurrentAction(action *walk.Action) error {
	defer func() {
		if !mw.pageCom.IsDisposed() {
			mw.pageCom.RestoreState()
		}
	}()

	mw.SetFocus()

	// 释放当前的分页系统资源
	if prevPage := mw.currentPage; prevPage != nil {
		mw.pageCom.SaveState()
		prevPage.SetVisible(false)
		prevPage.(walk.Widget).SetParent(nil)
		prevPage.Dispose()
	}

	// 创建新页面的函数
	newPage := mw.action2NewPage[action]
	page, err := newPage(mw.pageCom, mw.topContext)
	if err != nil {
		return err
	}

	// 标识接下来的函数动作不再需要处理了吗？
	action.SetChecked(true)

	// 将当前的显示区域的页面更新为新的页面元素，并且调用渲染函数重新渲染界面
	mw.currentPage = page
	mw.currentAction = action
	mw.currentPageChangedPublisher.Publish()

	return nil
}

func (mw *mainWindowConf) setNavToolBar() error {
	mw.navTB.SetSuspended(true)
	defer mw.navTB.SetSuspended(false)

	actions := mw.navTB.Actions()

	if err := actions.Clear(); err != nil {
		return err
	}

	for _, action := range mw.pageActions {
		if err := actions.Add(action); err != nil {
			return err
		}
	}

	// 设置当前action为toolbar中的其中一个action
	if mw.currentAction != nil {
		if !actions.Contains(mw.currentAction) { // 如果主窗体的action不在navtoolbar的action列表中
			for _, action := range mw.pageActions {
				if action != mw.currentAction {
					if err := mw.setCurrentAction(action); err != nil {
						return err
					}

					break
				}
			}
		}
	}

	return nil
}
