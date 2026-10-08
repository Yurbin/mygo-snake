#!/usr/bin/env bash
# 在 ubuntu:20.04 容器(银河麒麟 V10 SP1 同基座:GTK 3.24 / glibc 2.31)内冒烟运行 mygo-snake。
# 由 .github/workflows/ci.yml 的 kylin-smoke job 以
#   docker run --rm -v "$PWD":/src:ro ubuntu:20.04 bash /src/scripts/ci/smoke-kylin.sh
# 调用。覆盖三个风险点:
#   1) mygo 在老版 GTK 3.24 上的 purego 符号兼容(缺符号会启动即崩);
#   2) 无 GPU(Xvfb 无 GL)时的 CPU 渲染回退;
#   3) 用户级开始菜单集成与 --uninstall。

set -euo pipefail
export DEBIAN_FRONTEND=noninteractive

# ubuntu:20.04 已 EOL,源须指向 old-releases 才能装包(仅 CI 自举,不影响交付物;
# 麒麟 V10 SP1 自身仍在维护期,内网源可用)。
sed -i 's|//archive.ubuntu.com|//old-releases.ubuntu.com|g; s|//security.ubuntu.com|//old-releases.ubuntu.com|g' /etc/apt/sources.list
apt-get update -qq
apt-get install -y -qq --no-install-recommends libgtk-3-0 libglib2.0-0 xvfb dbus-x11 file >/dev/null

# /src 为只读挂载,二进制复制到 /tmp 运行
cp /src/mygo-snake /tmp/mygo-snake
chmod +x /tmp/mygo-snake

xvfb-run -a bash -euxc '
  # GTK3 桌面应用在无 session dbus 的容器里可能阻塞在 portal/通知调用上,提供一条会话总线
  eval "$(dbus-launch --sh-syntax)"

  # 1) 依赖自检脚本应全部通过(与麒麟 V10 SP1 同版 GTK 3.24 基础栈)
  /src/scripts/check-deps.sh

  # 2) 应用应存活满 8 秒直到被 timeout 终止(124=TERM 生效;137=TERM 未及退出被 KILL 兜底)
  set +e
  timeout -k 5 8 /tmp/mygo-snake > /tmp/out.log 2>&1
  code=$?
  set -e
  cat /tmp/out.log
  if grep -qiE "panic|fatal error" /tmp/out.log; then
    echo "[错误] 运行日志含 panic/fatal"
    exit 1
  fi
  if [ "$code" != 124 ] && [ "$code" != 137 ]; then
    echo "[错误] 应用未存活满 8 秒,exit=$code"
    exit 1
  fi

  # 3) 用户级开始菜单集成应已写入(零 root)
  test -f "$HOME/.local/share/applications/mygo-snake.desktop"
  grep -q "StartupWMClass=mygo-snake" "$HOME/.local/share/applications/mygo-snake.desktop"
  test -f "$HOME/.local/share/icons/hicolor/1024x1024/apps/mygo-snake.png"

  # 4) --uninstall 应清除上述文件
  /tmp/mygo-snake --uninstall
  test ! -e "$HOME/.local/share/applications/mygo-snake.desktop"
  echo "[OK] 麒麟同基座(ubuntu:20.04)冒烟全部通过"
'
