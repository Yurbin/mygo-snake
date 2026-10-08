#!/usr/bin/env bash
# ELF 兼容性自检:确保 linux 产物兼容银河麒麟 V10 SP1(glibc 2.31)。
#   1) 动态依赖仅限 glibc 基础三件套(libdl/libpthread/libc);
#   2) 引用的 GLIBC 符号版本 ≤ 2.31。
# 用法: scripts/check-elf.sh <二进制>

set -euo pipefail
BIN=${1:?用法: check-elf.sh <二进制>}

echo "文件类型: $(file -b "$BIN" 2>/dev/null || echo 未知)"

NEEDED="$(objdump -p "$BIN" | awk '/NEEDED/{print $2}' | sort -u)"
if [ -z "$NEEDED" ]; then
  echo "[OK] 静态二进制,无动态依赖"
  exit 0
fi
echo "动态依赖: $(echo "$NEEDED" | tr '\n' ' ')"
for lib in $NEEDED; do
  case "$lib" in
    libdl.so.2|libpthread.so.0|libc.so.0|libc.so.6|ld-linux-x86-64.so.2)
      ;;
    *)
      echo "[错误] 意外的动态依赖 $lib —— 运行期将要求目标机自带此库"
      exit 1
      ;;
  esac
done

MAX_GLIBC="$(objdump -T "$BIN" 2>/dev/null | grep -o "GLIBC_[0-9.]*" | sort -uV | tail -1 || true)"
echo "最高 GLIBC 符号版本: ${MAX_GLIBC:-无(未引用版本化符号)}"
if [ -n "$MAX_GLIBC" ]; then
  VER="${MAX_GLIBC#GLIBC_}"
  if [ "$(printf "%s\n" "2.31" "$VER" | sort -V | tail -1)" != "2.31" ]; then
    echo "[错误] GLIBC 符号 ${VER} 超过麒麟 V10 SP1 的 2.31 上限"
    exit 1
  fi
fi
echo "[OK] glibc 兼容麒麟 V10 SP1(2.31)"
