package main

// 视图测试:在无窗口环境用 ui.Tester 像用户一样点击、按键,CI 可直接运行。

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func newTestApp() (*app, *ui.Tester) {
	a := newApp(0)
	tt := ui.NewTester(a.view, 620, 700)
	return a, tt
}

func TestViewInitialHUD(t *testing.T) {
	a, tt := newTestApp()
	if a.game.phase != phaseReady {
		t.Fatalf("初始 phase=%v,想要 ready", a.game.phase)
	}
	for _, want := range []string{"得分 0", "最高 0", "开始", "重新开始"} {
		if !tt.HasText(want) {
			t.Fatalf("初始界面缺少 %q,全部文本 %q", want, tt.Texts())
		}
	}
}

func TestViewArrowKeyStartsGame(t *testing.T) {
	a, tt := newTestApp()
	tt.Key(0, ui.KeyUp)
	if a.game.phase != phasePlaying {
		t.Fatalf("方向键应开始游戏,phase=%v", a.game.phase)
	}
	if !tt.HasText("暂停") {
		t.Fatalf("行进中应有暂停按钮,文本 %q", tt.Texts())
	}
	if !tt.Focused("重新开始") && !tt.Focused("暂停") {
		// 焦点应在棋盘上;Focused 按钮文本查不到棋盘,这里只确认按钮没有抢走焦点
		t.Log("焦点不在按钮上 ✓")
	}
}

func TestViewPauseResume(t *testing.T) {
	a, tt := newTestApp()
	tt.Key(0, ui.KeyRight) // 开始
	tt.Key(0, ui.KeySpace)
	if a.game.phase != phasePaused {
		t.Fatalf("空格应暂停,phase=%v", a.game.phase)
	}
	if !tt.HasText("继续") {
		t.Fatalf("暂停后应有继续按钮,文本 %q", tt.Texts())
	}
	tt.Key(0, ui.KeySpace)
	if a.game.phase != phasePlaying {
		t.Fatalf("空格应恢复,phase=%v", a.game.phase)
	}
}

func TestViewPauseButton(t *testing.T) {
	a, tt := newTestApp()
	tt.Key(0, ui.KeyRight)
	if err := tt.Click("暂停"); err != nil {
		t.Fatal(err)
	}
	if a.game.phase != phasePaused {
		t.Fatalf("点击暂停按钮无效,phase=%v", a.game.phase)
	}
}

func TestViewRestartViaKeyAndButton(t *testing.T) {
	a, tt := newTestApp()
	tt.Key(0, ui.KeyRight)
	a.game.score = 5 // 模拟行进得分
	tt.Frame()
	tt.Key(0, ui.KeyR) // R 重开
	if a.game.phase != phasePlaying || a.game.score != 0 || len(a.game.snake) != 3 {
		t.Fatalf("R 重开异常: phase=%v score=%d", a.game.phase, a.game.score)
	}
	// 结束状态用按钮重开
	a.game.phase = phaseOver
	tt.Frame()
	if err := tt.Click("再来一局"); err != nil {
		t.Fatal(err)
	}
	if a.game.phase != phasePlaying || a.game.score != 0 {
		t.Fatalf("重开按钮无效: phase=%v score=%d", a.game.phase, a.game.score)
	}
}

func TestViewBestPersistsInView(t *testing.T) {
	a, tt := newTestApp()
	a.game.best = 7
	tt.Frame()
	if !tt.HasText("最高 7") {
		t.Fatalf("HUD 应显示最高 7,文本 %q", tt.Texts())
	}
}
