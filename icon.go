package main

// 应用图标(构建时嵌入,运行时用于窗口图标与 Linux 用户级菜单图标)。

import _ "embed"

//go:embed resources/icon.png
var iconPNG []byte
