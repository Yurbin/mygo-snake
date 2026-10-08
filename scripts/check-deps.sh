#!/bin/sh
# 贪吃蛇 依赖自检(断言模式):缺失才报错并给出内网源安装指引,不尝试自动安装。
# 用法: ./check-deps.sh

fail=0

for lib in libgtk-3.so.0 libglib-2.0.so.0 libgobject-2.0.so.0 libgio-2.0.so.0 libcairo.so.2 libgdk_pixbuf-2.0.so.0; do
  if ldconfig -p 2>/dev/null | grep -q "$lib"; then
    echo "[OK] $lib"
  else
    echo "[缺失] $lib —— 请从内网软件源安装: sudo apt install libgtk-3-0"
    fail=1
  fi
done

if [ -n "$DISPLAY" ] || [ -n "$WAYLAND_DISPLAY" ]; then
  echo "[OK] 图形会话"
else
  echo "[警告] 未检测到图形会话(DISPLAY/WAYLAND_DISPLAY 均为空)"
  echo "       请在桌面环境的终端中运行本程序"
  fail=1
fi

if [ "$fail" -eq 0 ]; then
  echo "依赖自检通过,运行 ./mygo-snake 开始游戏"
fi
exit $fail
