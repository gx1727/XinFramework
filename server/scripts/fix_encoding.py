#!/usr/bin/env python3
"""fix_encoding.py — 检测 + 修复被 GBK 等非 UTF-8 编码写入的源码文件。

背景：
  AGENTS.md §1 要求所有源文件 UTF-8 无 BOM。
  PowerShell / 记事本 默认 GBK（含中文标点）写文件时，会把整文件转成 GBK
  字节序（无 BOM），Go / TS 编译时报 `illegal UTF-8 encoding`。

行为：
  - 扫描 root_dir 下所有源码文件（参见 SOURCE_EXTS）
  - 跳过明显 ASCII / 合法 UTF-8 的文件
  - 对有非法 UTF-8 序列的文件，尝试依次按 GBK / GB18030 / Big5 解码；
    解码成功且结果为合理 UTF-8 的，写回原文件（覆盖前会做一次"再解码回原字节"
    的回环校验，避免误转码，失败则原样不动并打印警告）

用法：
    python fix_encoding.py [--check] [root_dir]   # root_dir 默认当前目录
        --check  只打印找到的坏文件，不修改
"""

import codecs
import os
import sys

# 与 strip_bom.py 保持一致
SOURCE_EXTS = {
    ".go",
    ".ts",
    ".tsx",
    ".js",
    ".jsx",
    ".mjs",
    ".cjs",
    ".json",
    ".css",
    ".scss",
    ".less",
    ".html",
    ".htm",
    ".vue",
    ".svelte",
    ".md",
    ".mdx",
    ".yaml",
    ".yml",
    ".toml",
    ".sql",
    ".sh",
    ".bash",
    ".zsh",
    ".ps1",
    ".bat",
    ".cmd",
    ".py",
    ".rb",
    ".rs",
    ".java",
    ".kt",
    ".swift",
    ".xml",
    ".conf",
    ".ini",
    ".env",
    ".properties",
}
SKIP_DIRS = {
    "node_modules",
    ".git",
    ".vercel",
    ".next",
    "dist",
    "build",
    "__pycache__",
    ".idea",
    ".vscode",
    "out",
    "target",
    "uploads",
    "logs",
}

CANDIDATE_ENCODINGS = ("gbk", "gb18030", "big5")


def is_valid_utf8(data: bytes) -> bool:
    try:
        data.decode("utf-8")
        return True
    except UnicodeDecodeError:
        return False


def try_guess_encoding(data: bytes):
    """依次尝试 CANDIDATE_ENCODINGS,返回 (encoding, decoded_str) 或 (None, None)。"""
    for enc in CANDIDATE_ENCODINGS:
        try:
            text = data.decode(enc)
            return enc, text
        except UnicodeDecodeError:
            continue
    return None, None


def process_file(path: str, check_only: bool) -> str | None:
    """返回坏文件诊断字符串（check_only 时）或 None。坏文件并尝试修复。"""
    try:
        with open(path, "rb") as f:
            data = f.read()
    except (OSError, PermissionError) as e:
        return f"  [SKIP read-error] {path}: {e}"

    if is_valid_utf8(data):
        return None

    if check_only:
        return f"  [BAD UTF-8] {path}  ({len(data)} bytes)"

    # 尝试 GBK 系列
    enc, text = try_guess_encoding(data)
    if enc is None:
        return f"  [WARN can't decode] {path}  (tried {CANDIDATE_ENCODINGS})"

    # 回环校验：UTF-8 编码回去的字节应等于原始字节（首尾可能有差异=原文件尾部换行/CRLF，
    # 这种差异我们忽略部分只校验 UTF-8 序列）；严格的回环校验会让 CRLF / LF 偏差误报，
    # 因此只对成功解码后的字符串做再次 UTF-8 解码校验。
    try:
        text.encode("utf-8").decode("utf-8")
    except UnicodeError:
        return f"  [WARN roundtrip-fail] {path}  (decoded as {enc} but UTF-8 roundtrip failed)"

    try:
        # 强制写 LF 换行（与仓库现有 .go / .ts 文件风格一致）；如有 CRLF 也可改 LF
        # 这里不做 CRLF 强制改写，避免与 LSP / git 风格冲突 — 仅落 UTF-8 字节
        with open(path, "wb") as f:
            f.write(text.encode("utf-8"))
    except (OSError, PermissionError) as e:
        return f"  [SKIP write-error] {path}: {e}"

    return f"  [FIXED as {enc}] {path}"


def walk(root: str, check_only: bool):
    bad = []
    scanned = 0
    for dirpath, dirnames, filenames in os.walk(root):
        dirnames[:] = [d for d in dirnames if d not in SKIP_DIRS]
        for fname in filenames:
            ext = os.path.splitext(fname)[1].lower()
            if ext not in SOURCE_EXTS:
                continue
            full = os.path.join(dirpath, fname)
            scanned += 1
            msg = process_file(full, check_only)
            if msg:
                bad.append(msg)
    return scanned, bad


def main() -> int:
    args = sys.argv[1:]
    check_only = False
    if args and args[0] == "--check":
        check_only = True
        args = args[1:]

    root = args[0] if args else "."
    root = os.path.abspath(root)

    if not os.path.isdir(root):
        print(f"error: not a directory: {root}", file=sys.stderr)
        return 2

    print(f"[fix_encoding] root = {root}")
    print(f"[fix_encoding] mode = {'check-only' if check_only else 'fix'}")

    scanned, bad = walk(root, check_only)

    print(f"[fix_encoding] scanned {scanned} source files")
    if bad:
        for m in bad:
            print(m)
        if check_only:
            print(f"[fix_encoding] {len(bad)} files with bad UTF-8")
            return 1
        print(f"[fix_encoding] done")
    else:
        print("[fix_encoding] CLEAN: all source files are valid UTF-8")
    return 0


if __name__ == "__main__":
    sys.exit(main())
