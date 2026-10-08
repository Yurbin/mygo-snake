package main

// 视图层:棋盘绘制、键盘输入、HUD 与状态遮罩。
// 视图每帧由 mygo 调用:advance() 先按时间推进游戏,再从状态构建界面。

import (
	"strconv"
	"time"

	"github.com/egoist/mygo/ui"
)

// 一帧内最多补进的 tick 数:窗口被遮挡较久后恢复时,避免一次追平太多步。
const maxCatchupTicks = 3

// 蛇身配色:头亮尾暗的绿色渐变,深浅色主题下都成立(棋盘底色用 Theme().Surface)。
var (
	snakeHeadColor = ui.Hex("#22c55e")
	snakeTailColor = ui.Hex("#15803d")
	eyeWhite       = ui.Hex("#ffffff")
	eyeDark        = ui.Hex("#052e16")
	overlayDim     = ui.RGBA(15, 23, 42, 0.55) // 遮罩压暗色,深浅主题通用
	overlayTitle   = ui.Hex("#ffffff")
	overlaySub     = ui.Hex("#cbd5e1")
)

func (a *app) view(c *ui.Context) {
	t := c.Theme()
	g := a.game
	a.advance(c, g)

	ui.Column(c).Fill().Padding(16).Gap(12).Children(func() {
		a.hud(c, t, g)
		ui.Box(c).Fill().Center().Children(func() {
			ui.Box(c).
				Fill().
				AspectRatio(1).
				Draw(a.paintBoard(t, g)).
				DrawOver(a.paintOverlay(t, g)).
				Focusable().
				AutoFocus().
				HandleInput(a.handleKey)
		})
		ui.Text(c, "方向键 / WASD 移动 · 空格 暂停 · R 重开 · "+versionLabel()).
			TextColor(t.TextMuted).FontSize(13).TextAlign(ui.Center)
	})
}

// hud 是顶栏:得分、最高分与操作按钮。
func (a *app) hud(c *ui.Context, t *ui.Theme, g *game) {
	ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
		ui.Textf(c, "得分 %d", g.score).FontSize(20).Bold()
		ui.Textf(c, "最高 %d", g.best).FontSize(15).TextColor(t.TextMuted)
		ui.Box(c).Grow(1)
		switch g.phase {
		case phasePlaying:
			if ui.Button(c, "暂停").Clicked() {
				g.togglePause()
			}
		case phaseOver:
			if ui.Button(c, "再来一局").Clicked() {
				a.restart()
			}
		case phasePaused:
			if ui.Button(c, "继续").Clicked() {
				g.togglePause()
			}
		default: // phaseReady
			if ui.Button(c, "开始").Clicked() {
				g.start()
			}
		}
		if ui.PrimaryButton(c, "重新开始").Clicked() {
			a.restart()
		}
	})
}

// advance 是游戏的时间驱动循环:
// 行进中在精确的 tick 时刻请求下一帧(c.After),Draw 回调只画不动状态。
func (a *app) advance(c *ui.Context, g *game) {
	now := c.Now()
	if g.phase == phasePlaying {
		if a.nextTick.IsZero() {
			a.nextTick = now.Add(g.Interval())
		} else {
			for i := 0; i < maxCatchupTicks && g.phase == phasePlaying && !now.Before(a.nextTick); i++ {
				g.tick()
				a.nextTick = a.nextTick.Add(g.Interval())
			}
			if g.phase == phasePlaying && now.After(a.nextTick) {
				a.nextTick = now.Add(g.Interval()) // 落后太多时直接从现在起算
			}
		}
		if g.phase == phasePlaying {
			c.After(a.nextTick.Sub(now))
		}
	}
	if g.phase != phasePlaying {
		a.nextTick = time.Time{}
	}
	if g.best > a.savedBest {
		a.saveBest()
	}
}

// handleKey 接管棋盘上的按键。返回 true 表示事件已消费。
func (a *app) handleKey(ev ui.InputEvent) bool {
	if ev.Kind != ui.InputKeyDown {
		return false
	}
	g := a.game
	switch ev.Key {
	case ui.KeyUp, ui.KeyW:
		g.start()
		g.turn(dirUp)
	case ui.KeyDown, ui.KeyS:
		g.start()
		g.turn(dirDown)
	case ui.KeyLeft, ui.KeyA:
		g.start()
		g.turn(dirLeft)
	case ui.KeyRight, ui.KeyD:
		g.start()
		g.turn(dirRight)
	case ui.KeySpace:
		switch g.phase {
		case phaseReady:
			g.start()
		case phasePlaying, phasePaused:
			g.togglePause()
		}
	case ui.KeyEnter:
		if g.phase == phaseOver {
			a.restart()
		} else {
			return false
		}
	case ui.KeyR:
		a.restart()
	default:
		return false
	}
	return true
}

// paintBoard 绘制棋盘、食物与蛇。
func (a *app) paintBoard(t *ui.Theme, g *game) func(p *ui.Painter, r ui.Rect) {
	return func(p *ui.Painter, r ui.Rect) {
		board := squareIn(r)
		p.Fill(board, t.Surface, 14)
		p.Stroke(board, t.Border, 14, 1)

		cell := board.W / float32(g.Cols)
		pad := cell * 0.11 // 格间留白

		// 食物:红苹果
		if g.food.X >= 0 {
			fr := insetRect(cellRect(board, cell, g.food), pad*1.15)
			p.Fill(fr, t.Danger, cell)
			// 高光
			hi := ui.Rect{X: fr.X + fr.W*0.18, Y: fr.Y + fr.H*0.14, W: fr.W * 0.26, H: fr.H * 0.26}
			p.Fill(hi, ui.RGBA(255, 255, 255, 0.45), cell)
		}

		// 蛇身:尾(暗)画起,头(亮)压在上面
		n := len(g.snake)
		for i := n - 1; i >= 0; i-- {
			f := float32(0)
			if n > 1 {
				f = float32(i) / float32(n-1)
			}
			col := mixColor(snakeHeadColor, snakeTailColor, f)
			p.Fill(insetRect(cellRect(board, cell, g.snake[i]), pad), col, cell*0.32)
		}

		// 蛇头眼睛:沿行进方向前移,左右各一
		hr := cellRect(board, cell, g.snake[0])
		d := dirVec[g.dir]
		fx, fy := float32(d.X), float32(d.Y)
		px, py := -fy, fx
		cx, cy := hr.X+hr.W/2, hr.Y+hr.H/2
		for _, s := range []float32{-1, 1} {
			ex := cx + fx*cell*0.14 + px*s*cell*0.17
			ey := cy + fy*cell*0.14 + py*s*cell*0.17
			p.Fill(circleRect(ex, ey, cell*0.12), eyeWhite, cell)
			p.Fill(circleRect(ex+fx*cell*0.03, ey+fy*cell*0.03, cell*0.055), eyeDark, cell)
		}
	}
}

// paintOverlay 在非行进状态下画半透明遮罩与提示文案。
func (a *app) paintOverlay(t *ui.Theme, g *game) func(p *ui.Painter, r ui.Rect) {
	return func(p *ui.Painter, r ui.Rect) {
		if g.phase == phasePlaying {
			return
		}
		board := squareIn(r)
		p.Fill(board, overlayDim, 14)

		title, sub := "", ""
		switch g.phase {
		case phaseReady:
			title, sub = "按方向键开始", "↑ ↓ ← → 或 WASD 移动 · 空格暂停"
		case phasePaused:
			title, sub = "已暂停", "按空格继续"
		case phaseOver:
			if g.score >= g.Cols*g.Rows-3 {
				title, sub = "通关!蛇占满了棋盘", "得分全部拿下"
			} else {
				title, sub = "游戏结束", "本局得分 "+strconv.Itoa(g.score)+" · 最高 "+strconv.Itoa(g.best)+" · 回车重开"
			}
		}
		ts := ui.Span{Text: title, Size: 30, Weight: 700, Color: overlayTitle}
		ss := ui.Span{Text: sub, Size: 15, Color: overlaySub}
		tw, th := p.MeasureText(0, ts)
		sw, _ := p.MeasureText(0, ss)
		midY := board.Y + board.H/2
		p.RichText(board.X+(board.W-tw)/2, midY-th-8, 0, ts)
		p.RichText(board.X+(board.W-sw)/2, midY+10, 0, ss)
	}
}

// squareIn 返回 r 内居中的最大正方形。
func squareIn(r ui.Rect) ui.Rect {
	size := min(r.W, r.H)
	return ui.Rect{X: r.X + (r.W-size)/2, Y: r.Y + (r.H-size)/2, W: size, H: size}
}

func cellRect(board ui.Rect, cell float32, c cell) ui.Rect {
	return ui.Rect{X: board.X + float32(c.X)*cell, Y: board.Y + float32(c.Y)*cell, W: cell, H: cell}
}

func insetRect(r ui.Rect, d float32) ui.Rect {
	return ui.Rect{X: r.X + d, Y: r.Y + d, W: r.W - 2*d, H: r.H - 2*d}
}

// circleRect 以 (x,y) 为圆心、rad 为半径的圆(Fill 用 radius=rad 的方框近似)。
func circleRect(x, y, rad float32) ui.Rect {
	return ui.Rect{X: x - rad, Y: y - rad, W: rad * 2, H: rad * 2}
}

func mixColor(a, b ui.Color, t float32) ui.Color {
	return ui.Color{
		R: mixByte(a.R, b.R, t),
		G: mixByte(a.G, b.G, t),
		B: mixByte(a.B, b.B, t),
		A: 255,
	}
}

func mixByte(x, y uint8, t float32) uint8 {
	return uint8(float32(x) + (float32(y)-float32(x))*t)
}
