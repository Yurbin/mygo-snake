//go:build !linux

package main

// 非 Linux 平台的桌面集成占位:macOS/Windows 开发时不需要菜单集成。

func desktopStartup() bool { return false }

func installDesktopEntry() {}
