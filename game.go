package main

// 贪吃蛇核心逻辑:纯状态机,不依赖 UI,可在无窗口环境(单元测试、CI)运行。

import (
	"math/rand"
	"time"
)

// cell 是网格坐标,左上角为 (0,0)。
type cell struct{ X, Y int }

// direction 是蛇的行进方向。
type direction int

const (
	dirUp direction = iota
	dirRight
	dirDown
	dirLeft
)

var dirVec = [4]cell{
	dirUp:    {X: 0, Y: -1},
	dirRight: {X: 1, Y: 0},
	dirDown:  {X: 0, Y: 1},
	dirLeft:  {X: -1, Y: 0},
}

func (d direction) opposite() direction { return direction((int(d) + 2) % 4) }

// phase 是游戏的阶段。
type phase int

const (
	phaseReady   phase = iota // 就绪:按方向键开始
	phasePlaying              // 行进中
	phasePaused               // 已暂停
	phaseOver                 // 已结束(撞墙、撞自己,或占满棋盘通关)
)

const (
	// 每吃 foodsPerLevel 个食物提速一档。
	foodsPerLevel = 5
	baseInterval = 170 * time.Millisecond
	minInterval  = 70 * time.Millisecond
	levelStep    = 14 * time.Millisecond
	// turnQueueCap 限制转向队列长度,防止一次连按把方向缓存挤爆。
	turnQueueCap = 2
)

// game 持有贪吃蛇的完整状态。
type game struct {
	Cols, Rows int

	snake []cell // 蛇头在前
	dir   direction
	queue []direction // 待应用的转向队列,防一帧内 180° 回头

	food  cell
	score int
	best  int // 历史最高分,跨局保留

	phase phase
	rng   *rand.Rand
}

func newGame(cols, rows int, best int, seed int64) *game {
	g := &game{Cols: cols, Rows: rows, best: best, rng: rand.New(rand.NewSource(seed))}
	g.reset()
	return g
}

// reset 恢复初始局面(保留 best)。
func (g *game) reset() {
	cx, cy := g.Cols/2, g.Rows/2
	g.snake = []cell{{X: cx, Y: cy}, {X: cx - 1, Y: cy}, {X: cx - 2, Y: cy}}
	g.dir = dirRight
	g.queue = g.queue[:0]
	g.score = 0
	g.phase = phaseReady
	g.placeFood()
}

// level 返回当前速度档位(从 0 开始)。
func (g *game) level() int { return g.score / foodsPerLevel }

// Interval 返回当前两个 tick 之间的间隔。
func (g *game) Interval() time.Duration {
	d := baseInterval - time.Duration(g.level())*levelStep
	if d < minInterval {
		d = minInterval
	}
	return d
}

// turn 入队一个转向。与当前方向(或队列末尾方向)相同、相反的输入,以及超长队列,都被忽略。
func (g *game) turn(d direction) {
	if g.phase == phaseOver {
		return
	}
	last := g.dir
	if n := len(g.queue); n > 0 {
		last = g.queue[n-1]
	}
	if d == last || d == last.opposite() {
		return
	}
	if len(g.queue) >= turnQueueCap {
		return
	}
	g.queue = append(g.queue, d)
}

// start 让就绪状态进入行进(首个方向键)。
func (g *game) start() {
	if g.phase == phaseReady {
		g.phase = phasePlaying
	}
}

// togglePause 在行进与暂停之间切换。
func (g *game) togglePause() {
	switch g.phase {
	case phasePlaying:
		g.phase = phasePaused
	case phasePaused:
		g.phase = phasePlaying
	}
}

// restart 重开一局,直接进入行进。
func (g *game) restart() {
	g.reset()
	g.phase = phasePlaying
}

// tick 推进一步:出队转向、前进、吃食物或判定死亡。
func (g *game) tick() {
	if g.phase != phasePlaying {
		return
	}
	if len(g.queue) > 0 {
		g.dir = g.queue[0]
		g.queue = g.queue[1:]
	}
	head := g.snake[0]
	next := cell{X: head.X + dirVec[g.dir].X, Y: head.Y + dirVec[g.dir].Y}

	// 撞墙
	if next.X < 0 || next.X >= g.Cols || next.Y < 0 || next.Y >= g.Rows {
		g.die()
		return
	}
	// 撞自己:不吃食物时尾巴会移走,尾格可以进入
	body := g.snake
	eat := next == g.food
	if !eat {
		body = g.snake[:len(g.snake)-1]
	}
	for _, c := range body {
		if c == next {
			g.die()
			return
		}
	}

	g.snake = append([]cell{next}, g.snake...)
	if eat {
		g.score++
		if g.score > g.best {
			g.best = g.score
		}
		g.placeFood()
	} else {
		g.snake = g.snake[:len(g.snake)-1]
	}
}

func (g *game) die() {
	g.phase = phaseOver
	g.queue = g.queue[:0]
}

// placeFood 在空格上随机放置食物;蛇占满棋盘时通关。
func (g *game) placeFood() {
	free := g.Cols*g.Rows - len(g.snake)
	if free <= 0 {
		g.food = cell{X: -1, Y: -1}
		g.phase = phaseOver
		return
	}
	for {
		c := cell{X: g.rng.Intn(g.Cols), Y: g.rng.Intn(g.Rows)}
		if !g.occupied(c) {
			g.food = c
			return
		}
	}
}

func (g *game) occupied(c cell) bool {
	for _, s := range g.snake {
		if s == c {
			return true
		}
	}
	return false
}
