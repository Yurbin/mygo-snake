//go:build linux

package main

// 麒麟 V10 SP1(内网)交付经验(wails-snake 项目真机验证):
//   - 普通用户没有 root,系统级安装(deb/rpm、写 /usr/share)一律不可行;
//   - 一切桌面集成走用户级 freedesktop 路径,零 root;
//   - 每次启动幂等重写菜单项,Exec 指向当前真实路径,移动目录后重跑即自愈;
//   - --uninstall 删除自己写的用户级文件,彻底删除 = 卸载 + 删程序目录。

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	desktopAppID = "mygo-snake"
	desktopName  = "贪吃蛇"
)

// desktopStartup 处理 --uninstall 等桌面集成参数。返回 true 表示已处理完毕,程序应退出。
func desktopStartup() bool {
	if len(os.Args) < 2 {
		return false
	}
	switch os.Args[1] {
	case "--uninstall":
		uninstallDesktop()
		return true
	case "--install-menu":
		// 手动重新集成菜单(平时启动即自动做,此参数备用)
		installDesktopEntry()
		return true
	}
	return false
}

// installDesktopEntry 幂等写用户级菜单项与图标(freedesktop 标准,UKUI 兼容)。
func installDesktopEntry() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	exe, err := os.Executable()
	if err != nil {
		return
	}
	// Exec 用符号链接解析后的绝对路径并加引号:目录被移动/含中文空格都能正常启动
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		if abs, err := filepath.Abs(resolved); err == nil {
			exe = abs
		}
	}

	appDir := filepath.Join(home, ".local", "share", "applications")
	iconDir := filepath.Join(home, ".local", "share", "icons", "hicolor", "1024x1024", "apps")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return
	}
	if err := os.MkdirAll(iconDir, 0o755); err == nil {
		_ = os.WriteFile(filepath.Join(iconDir, desktopAppID+".png"), iconPNG, 0o644)
	}

	// StartupWMClass 与 GTK 默认 WMClass(= 程序名)对齐,任务栏才能正确分组
	entry := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=%s
Comment=贪吃蛇小游戏 · MyGo 原生 UI
Exec="%s"
Icon=%s
StartupWMClass=%s
Terminal=false
Categories=Game;ArcadeGame;
`, desktopName, exe, desktopAppID, desktopAppID)

	_ = os.WriteFile(filepath.Join(appDir, desktopAppID+".desktop"), []byte(entry), 0o644)

	// 卸载入口:Terminal=true 的第二项,方便非开发者使用
	uninstall := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=卸载%s
Comment=删除贪吃蛇的菜单项与用户数据
Exec=bash -c '%s --uninstall; echo; echo 按回车键关闭; read'
Icon=%s
Terminal=true
Categories=Game;
`, desktopName, exe, desktopAppID)
	_ = os.WriteFile(filepath.Join(appDir, desktopAppID+"-uninstall.desktop"), []byte(uninstall), 0o644)
}

// uninstallDesktop 删除本程序写入的全部用户级文件;随后提示删除程序目录。
func uninstallDesktop() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	removed := 0
	rm := func(path string) {
		if os.Remove(path) == nil {
			removed++
			fmt.Println("已删除", path)
		}
	}
	rm(filepath.Join(home, ".local", "share", "applications", desktopAppID+".desktop"))
	rm(filepath.Join(home, ".local", "share", "applications", desktopAppID+"-uninstall.desktop"))
	rm(filepath.Join(home, ".local", "share", "icons", "hicolor", "1024x1024", "apps", desktopAppID+".png"))
	rm(filepath.Join(stateDir(), "state.json"))
	rm(filepath.Join(stateDir(), "window-state.json"))

	fmt.Printf("卸载完成(删除 %d 个文件)。程序目录请手动删除。\n", removed)
}
