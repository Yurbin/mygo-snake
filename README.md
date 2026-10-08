# 贪吃蛇 · mygo-snake

用 [MyGo](https://github.com/egoist/mygo) **原生 UI** 编写的贪吃蛇桌面游戏:纯 Go、无 webview、无 cgo,界面由框架在 GPU(无 GPU 时自动回退 CPU)上直接绘制。

目标平台:**Linux x64 · 银河麒麟 V10 SP1**,交付形态:**绿色免安装包**(零 root、零联网)。

## 玩法

| 按键 | 作用 |
|---|---|
| 方向键 / WASD | 移动(按任意方向键开始) |
| 空格 | 暂停 / 继续 |
| R | 重新开始 |
| 回车 | 结束后重开 |

吃到苹果 +1 分,每 5 个提速一档;最高分持久化在 `~/.config/mygo-snake/`。

## 开发

```sh
go tool mygo dev   # 边改边跑,自动重建重启
go test ./...      # 逻辑与视图测试(无窗口)
```

视觉快照(人工核对用):

```sh
SNAKE_SNAPSHOT_DIR=/tmp/shots go test -run TestSnapshots .
```

图标重新生成:`go run tools/genicon/main.go`

## 构建

纯 Go 无 cgo,任意平台可直接交叉编译:

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath \
  -ldflags "-s -w -X main.version=0.1.0" -o build/linux-amd64/mygo-snake .
```

## CI / 发布(GitHub Actions)

- **push / PR** → `.github/workflows/ci.yml`:vet + 无窗口测试 + linux/amd64 交叉构建冒烟 + [麒麟兼容性自检](scripts/check-elf.sh)(动态依赖仅限 glibc 三件套、GLIBC 符号 ≤ 2.31)
- **推 tag `vX.Y.Z`** → `.github/workflows/release.yml`:校验 tag 与 `mygo.json` 版本一致 → 生产构建 → 组装绿色包 + sha256 → 挂到 GitHub Release

```sh
git tag v0.1.0 && git push origin v0.1.0
```

## 绿色免安装包

```
mygo-snake_<版本>_linux_amd64_portable/
├── mygo-snake        # 二进制
├── mygo-snake.png    # 图标
├── check-deps.sh     # 断言式依赖自检
└── 启动说明.txt       # 面向最终用户的中文说明
```

交付流程:外网下载 `tar.gz` + `.sha256` → 摆渡进内网 → `sha256sum -c` → 解压 → `./check-deps.sh` → 运行,全程零联网零 root。

## 麒麟 V10 SP1 兼容性

- **glibc**:二进制仅引用 glibc 2.2.5 时代的符号(dlopen 家族),远低于麒麟 2.31;构建期不链接任何 GUI 库
- **运行期依赖**:仅 GTK3 基础栈(`libgtk-3 / libglib-2.0 / libgobject-2.0 / libgio-2.0 / libcairo / libgdk_pixbuf`),由 mygo 经 purego 按符号 `dlopen` 目标机自带库——麒麟 V10 SP1 桌面版自带 GTK 3.24,无需安装任何东西
- **无需 WebKitGTK**(对比 wails 方案)
- **渲染**:有 GPU 走 OpenGL(GtkGLArea),虚机/软渲染自动回退 CPU(GtkDrawingArea)
- **X11 / Wayland 双支持**(UKUI 为 X11)
- **开始菜单集成**:程序运行时幂等写用户级 `~/.local/share/applications`(无需 root),移动目录后重跑自动修正;`--uninstall` 一键卸载

## 许可

MIT
