#!/usr/bin/env bash
# 组装麒麟 V10 SP1 绿色免安装包:
#   mygo-snake_<版本>_linux_amd64_portable/{mygo-snake, mygo-snake.png, check-deps.sh, 启动说明.txt}
# 并产出 <包名>.tar.gz 与 .sha256(供离线摆渡校验)。
#
# 用法: scripts/package-portable.sh <linux/amd64 二进制> <图标 png> <版本号> <输出目录>

set -euo pipefail

APP=mygo-snake
BIN=${1:?用法: package-portable.sh <二进制> <图标> <版本> <输出目录>}
ICON=${2:?缺少图标参数}
VERSION=${3:?缺少版本参数}
OUT=${4:?缺少输出目录参数}
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

NAME="${APP}_${VERSION}_linux_amd64_portable"
DIR="$OUT/$NAME"
rm -rf "$DIR"
mkdir -p "$DIR"

cp "$BIN" "$DIR/$APP"
cp "$ICON" "$DIR/$APP.png"
cp "$ROOT/scripts/check-deps.sh" "$DIR/check-deps.sh"
cp "$ROOT/scripts/启动说明.txt" "$DIR/启动说明.txt"
chmod +x "$DIR/$APP" "$DIR/check-deps.sh"

tar -C "$OUT" -czf "$OUT/$NAME.tar.gz" "$NAME"
if command -v sha256sum >/dev/null 2>&1; then
  (cd "$OUT" && sha256sum "$NAME.tar.gz" > "$NAME.tar.gz.sha256")
else
  (cd "$OUT" && shasum -a 256 "$NAME.tar.gz" > "$NAME.tar.gz.sha256")
fi

echo "已生成:"
ls -l "$OUT/$NAME.tar.gz" "$OUT/$NAME.tar.gz.sha256"
