package main

// 最高分持久化:JSON 存到用户数据目录,不写安装目录(可能只读)。

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/egoist/mygo"
)

const appIdentifier = "mygo-snake"

type savedState struct {
	Best int `json:"best"`
}

// stateDir 返回用户数据目录;mygo 路径不可用时退回 ~/.config/<appid>。
func stateDir() string {
	if dir, err := mygo.App.Path(mygo.PathUserData); err == nil && dir != "" {
		return dir
	}
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, appIdentifier)
	}
	return ""
}

func loadBest() int {
	dir := stateDir()
	if dir == "" {
		return 0
	}
	b, err := os.ReadFile(filepath.Join(dir, "state.json"))
	if err != nil {
		return 0
	}
	var s savedState
	if json.Unmarshal(b, &s) != nil {
		return 0
	}
	if s.Best < 0 {
		return 0
	}
	return s.Best
}

// saveBest 在最高分变化时落盘。失败只影响持久化,不影响游戏。
func (a *app) saveBest() {
	dir := stateDir()
	if dir == "" {
		return
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	b, err := json.Marshal(savedState{Best: a.game.best})
	if err != nil {
		return
	}
	if os.WriteFile(filepath.Join(dir, "state.json"), b, 0o644) == nil {
		a.savedBest = a.game.best
	}
}
