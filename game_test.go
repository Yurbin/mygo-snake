package main

// 游戏核心逻辑的单元测试:无窗口、无 UI 依赖,CI 可直接运行。

import "testing"

func testGame() *game {
	return newGame(24, 24, 0, 1) // 固定种子,食物位置可复现
}

func head(g *game) cell { return g.snake[0] }

func TestInitialLayout(t *testing.T) {
	g := testGame()
	if got := len(g.snake); got != 3 {
		t.Fatalf("初始蛇长 %d,想要 3", got)
	}
	if head(g) != (cell{X: 12, Y: 12}) {
		t.Errorf("蛇头初始位置 %v,想要 (12,12)", head(g))
	}
	if g.dir != dirRight || g.phase != phaseReady || g.score != 0 {
		t.Errorf("初始状态异常: dir=%v phase=%v score=%d", g.dir, g.phase, g.score)
	}
	if g.occupied(g.food) {
		t.Errorf("食物 %v 落在蛇身上", g.food)
	}
}

func TestTickMovesAndFollows(t *testing.T) {
	g := testGame()
	g.start()
	g.tick()
	if head(g) != (cell{X: 13, Y: 12}) {
		t.Fatalf("tick 后蛇头 %v,想要 (13,12)", head(g))
	}
	if got := len(g.snake); got != 3 {
		t.Fatalf("不吃食物时蛇长应不变,得到 %d", got)
	}
	// 转向下再走一步,验证方向生效
	g.turn(dirDown)
	g.tick()
	if head(g) != (cell{X: 13, Y: 13}) {
		t.Fatalf("转向后蛇头 %v,想要 (13,13)", head(g))
	}
}

func TestTickOnlyWhenPlaying(t *testing.T) {
	g := testGame()
	g.tick() // phaseReady,tick 应为空操作
	if head(g) != (cell{X: 12, Y: 12}) {
		t.Fatalf("就绪状态下不应移动,蛇头 %v", head(g))
	}
}

func TestWallCollision(t *testing.T) {
	g := newGame(8, 8, 0, 1)
	g.start()
	for i := 0; i < 20 && g.phase == phasePlaying; i++ {
		g.tick()
	}
	if g.phase != phaseOver {
		t.Fatalf("撞墙后应结束,phase=%v 蛇头=%v", g.phase, head(g))
	}
}

func TestSelfCollision(t *testing.T) {
	// 头 (5,5) 向左,下一步 (4,5) 是身体(非尾格):应判死。
	g := testGame()
	g.snake = []cell{
		{X: 5, Y: 5}, {X: 4, Y: 5}, {X: 3, Y: 5}, {X: 2, Y: 5},
		{X: 2, Y: 6}, {X: 3, Y: 6}, {X: 4, Y: 6}, {X: 5, Y: 6}, {X: 6, Y: 6}, {X: 6, Y: 5},
	}
	g.dir = dirLeft
	g.food = cell{X: 0, Y: 0}
	g.phase = phasePlaying
	g.tick()
	if g.phase != phaseOver {
		t.Fatalf("撞到自己后应结束,phase=%v", g.phase)
	}
	if head(g) != (cell{X: 5, Y: 5}) {
		t.Fatalf("死亡时蛇头不应移动,得到 %v", head(g))
	}
}

func TestTailCellIsEnterableWhenNotEating(t *testing.T) {
	// 头 (5,6) 向右,下一步 (6,6) 恰是尾格;不吃食物时尾格将让出,允许进入。
	g := testGame()
	g.snake = []cell{
		{X: 5, Y: 6}, {X: 4, Y: 6}, {X: 3, Y: 6}, {X: 3, Y: 5},
		{X: 4, Y: 5}, {X: 5, Y: 5}, {X: 6, Y: 5}, {X: 6, Y: 6},
	}
	g.dir = dirRight
	g.food = cell{X: 0, Y: 0}
	g.phase = phasePlaying
	g.tick()
	if g.phase == phaseOver {
		t.Fatal("不吃食物时进入尾格不应判定死亡")
	}
	if head(g) != (cell{X: 6, Y: 6}) {
		t.Fatalf("蛇头 %v,想要 (6,6)", head(g))
	}
	if got := len(g.snake); got != 8 {
		t.Fatalf("蛇长 %d,想要 8(不增长)", got)
	}
}

func TestEatGrowsAndScores(t *testing.T) {
	g := testGame()
	g.start()
	g.food = cell{X: 13, Y: 12} // 正前方
	g.tick()
	if g.score != 1 {
		t.Fatalf("得分 %d,想要 1", g.score)
	}
	if got := len(g.snake); got != 4 {
		t.Fatalf("吃食物后蛇长 %d,想要 4", got)
	}
	if g.best != 1 {
		t.Fatalf("best %d,想要 1(首超即刷新)", g.best)
	}
}

func TestNoReverseTurn(t *testing.T) {
	g := testGame() // 初始向右
	g.turn(dirLeft) // 反向:忽略
	if len(g.queue) != 0 {
		t.Fatal("反向输入应被忽略")
	}
	g.turn(dirUp)
	g.turn(dirDown) // 与队列末尾(上)相反:忽略
	if len(g.queue) != 1 || g.queue[0] != dirUp {
		t.Fatalf("队列 %v,想要 [up]", g.queue)
	}
	g.turn(dirUp) // 与队列末尾相同:忽略
	if len(g.queue) != 1 {
		t.Fatalf("重复输入应被忽略,队列 %v", g.queue)
	}
	g.turn(dirLeft)
	g.turn(dirDown) // 队列已满(2):忽略
	if len(g.queue) != 2 {
		t.Fatalf("队列长度 %d,想要 2(封顶)", len(g.queue))
	}
}

func TestQueueAppliedPerTick(t *testing.T) {
	g := testGame()
	g.start()
	g.turn(dirUp)
	g.turn(dirLeft)
	g.tick() // 应用 up
	if g.dir != dirUp {
		t.Fatalf("第一个 tick 后 dir=%v,想要 up", g.dir)
	}
	g.tick() // 应用 left
	if g.dir != dirLeft {
		t.Fatalf("第二个 tick 后 dir=%v,想要 left", g.dir)
	}
	if head(g) != (cell{X: 11, Y: 11}) {
		t.Fatalf("蛇头 %v,想要 (11,11)", head(g))
	}
}

func TestPauseAndRestart(t *testing.T) {
	g := testGame()
	g.start()
	g.togglePause()
	g.tick() // 暂停中,空操作
	if head(g) != (cell{X: 12, Y: 12}) {
		t.Fatal("暂停时不应移动")
	}
	g.togglePause()
	g.tick()
	if head(g) != (cell{X: 13, Y: 12}) {
		t.Fatal("恢复后应继续移动")
	}
	g.restart()
	if g.score != 0 || g.phase != phasePlaying || len(g.snake) != 3 {
		t.Fatalf("重开后状态异常: score=%d phase=%v", g.score, g.phase)
	}
}

func TestBestCarriesOver(t *testing.T) {
	g := newGame(24, 24, 10, 1)
	g.start()
	g.food = cell{X: 13, Y: 12}
	g.tick()
	if g.best != 10 {
		t.Fatalf("未破纪录时 best 应保持 10,得到 %d", g.best)
	}
	g.restart()
	if g.best != 10 {
		t.Fatalf("重开后 best 应保留,得到 %d", g.best)
	}
}

func TestIntervalSpeedup(t *testing.T) {
	g := testGame()
	if got := g.Interval(); got != baseInterval {
		t.Fatalf("初始间隔 %v,想要 %v", got, baseInterval)
	}
	g.score = foodsPerLevel
	if got := g.Interval(); got != baseInterval-levelStep {
		t.Fatalf("1 档间隔 %v", got)
	}
	g.score = 100 // 远超上限
	if got := g.Interval(); got != minInterval {
		t.Fatalf("封底间隔 %v,想要 %v", got, minInterval)
	}
}

func TestPlaceFoodNeverOnSnake(t *testing.T) {
	for seed := int64(1); seed <= 50; seed++ {
		g := newGame(6, 6, 0, seed)
		if g.occupied(g.food) {
			t.Fatalf("seed=%d 食物 %v 落在蛇上", seed, g.food)
		}
	}
}

func TestWinByFillingBoard(t *testing.T) {
	// 4x4 棋盘,蛇占 15 格(头在前、盘成回字形),吃掉最后一个空格的食物后通关。
	g := newGame(4, 4, 0, 1)
	g.snake = []cell{
		{1, 3}, {2, 3}, {3, 3}, {3, 2},
		{2, 2}, {1, 2}, {0, 2}, {0, 1},
		{1, 1}, {2, 1}, {3, 1}, {3, 0},
		{2, 0}, {1, 0}, {0, 0},
	}
	g.dir = dirLeft
	g.food = cell{X: 0, Y: 3} // 最后一个空格
	g.phase = phasePlaying
	g.tick()
	if g.phase != phaseOver {
		t.Fatalf("占满棋盘应通关,phase=%v", g.phase)
	}
	if g.score != 1 || g.best != 1 {
		t.Fatalf("通关计分异常: score=%d best=%d", g.score, g.best)
	}
}
