package main

// 视觉快照:设置 SNAKE_SNAPSHOT_DIR=目录 后运行本测试,导出关键界面的 PNG,
// 供人工核对视觉效果(CI 不设置该变量,自动跳过)。

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
)

func snapshot(t *testing.T, tt *ui.Tester, name string) {
	t.Helper()
	dir := os.Getenv("SNAKE_SNAPSHOT_DIR")
	if dir == "" {
		t.Skip("未设置 SNAKE_SNAPSHOT_DIR,跳过快照")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(dir, name+".png"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, tt.Image()); err != nil {
		t.Fatal(err)
	}
}

func TestSnapshots(t *testing.T) {
	dir := os.Getenv("SNAKE_SNAPSHOT_DIR")
	if dir == "" {
		t.Skip("未设置 SNAKE_SNAPSHOT_DIR,跳过快照")
	}

	// 就绪
	a, tt := newTestApp()
	tt.Frame()
	snapshot(t, tt, "1-ready")

	// 行进中:构造一局有比分、蛇较长的局面
	a.game.phase = phasePlaying
	a.game.score = 13
	a.game.best = 21
	a.game.snake = []cell{
		{12, 12}, {11, 12}, {10, 12}, {10, 13}, {10, 14},
		{11, 14}, {12, 14}, {13, 14}, {14, 14}, {14, 13},
		{14, 12}, {14, 11}, {13, 11},
	}
	a.game.dir = dirRight
	a.game.food = cell{X: 16, Y: 12}
	tt.Frame()
	snapshot(t, tt, "2-playing")

	// 暂停
	a.game.togglePause()
	tt.Frame()
	snapshot(t, tt, "3-paused")

	// 结束
	a.game.phase = phaseOver
	tt.Frame()
	snapshot(t, tt, "4-over")

	// 深色主题
	a2, tt2 := newTestApp()
	tt2.SetDark(true)
	a2.game.phase = phasePlaying
	a2.game.score = 5
	a2.game.snake = []cell{{12, 12}, {11, 12}, {10, 12}, {10, 13}, {10, 14}, {11, 14}, {12, 14}}
	a2.game.dir = dirUp
	a2.game.food = cell{X: 12, Y: 8}
	tt2.Frame()
	snapshot(t, tt2, "5-playing-dark")
}
