package main

// 贪吃蛇 · MyGo 原生 UI
// 纯 Go、无 webview:窗口由 mygo 在 GPU(或回退的 CPU)上直接绘制。

import (
	"log"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// version 在 CI 发版时通过 -ldflags -X 注入(如 0.1.0),默认 dev。
var version = "dev"

func versionLabel() string {
	if version == "dev" {
		return "开发版"
	}
	return "v" + version
}

const (
	boardCols = 24
	boardRows = 24
)

// app 是窗口显示的全部状态:视图每帧从它重建界面。
type app struct {
	game      *game
	nextTick  time.Time // 行进中下一次 tick 的时刻;非行进时为零值
	savedBest int       // 已落盘的最高分
}

func newApp(best int) *app {
	g := newGame(boardCols, boardRows, best, time.Now().UnixNano())
	return &app{game: g, savedBest: best}
}

func (a *app) restart() {
	a.game.restart()
	a.nextTick = time.Time{}
}

func main() {
	if desktopStartup() { // linux:--uninstall / --install-menu
		return
	}
	a := newApp(loadBest())
	mygo.App.WhenReady(func() {
		installDesktopEntry() // linux 用户级菜单集成(幂等);其他平台 no-op
		w := mygo.NewWindow(mygo.WindowOptions{
			Title:     "贪吃蛇",
			Width:     620,
			Height:    700,
			MinWidth:  500,
			MinHeight: 560,
			StateKey:  "main", // 记住上次窗口位置与大小
			Content:   ui.View(a.view),
		})
		_ = w.SetIcon(iconPNG) // 标题栏/任务栏图标(Linux 为 GTK 窗口图标)
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
