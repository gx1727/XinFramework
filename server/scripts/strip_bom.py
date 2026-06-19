#!/usr/bin/env python3
"""strip_bom.py - 去除项目源码文件的 UTF-8 BOM (EF BB BF)。

项目约定（AGENTS.md 5.1）：所有源码文件使用 UTF-8 无 BOM。
PowerShell 默认 GBK 写入时容易 mangle 成 ? 或重复 prepend BOM，
导致 Go 报 `invalid BOM in the middle of the file` 等诡异错误。

用法：
    python strip_bom.py [root_dir]            # 默认扫当前目录
    python strip_bom.py --check [root_dir]    # 只检查不修改
"""

import os
import sys

# 常见源码扩展名（按需扩展）
SOURCE_EXTS = {
    ".go", ".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs",
    ".json", ".css", ".scss", ".less", ".html", ".htm", ".vue", ".svelte",
    ".md", ".mdx", ".yaml", ".yml", ".toml",
    ".sql", ".sh", ".bash", ".zsh", ".ps1", ".bat", ".cmd",
    ".py", ".rb", ".rs", ".java", ".kt", ".swift",
    ".xml", ".conf", ".ini", ".env", ".properties",
}

# 跳过这些目录
SKIP_DIRS = {
    "node_modules", ".git", ".vercel", ".next", "dist", "build",
    "__pycache__", ".idea", ".vscode", "out", "target",
}

UTF8_BOM = b"\xef\xbb\xbf"


def count_leading_boms(data: bytes) -> int:
    """统计 data 开头连续的 UTF-8 BOM 个数（0/1/2/...）。"""
    n = 0
    while data[n * 3 : n * 3 + 3] == UTF8_BOM:
        n += 1
    return n


def process_file(path: str, check_only: bool) -> int:
    """处理一个文件，返回剥离的 BOM 数量。check_only=True 时不写盘。"""
    try:
        with open(path, "rb") as f:
            data = f.read()
    except (OSError, PermissionError) as e:
        print(f"  [SKIP read-error] {path}: {e}", file=sys.stderr)
        return 0

    n = count_leading_boms(data)
    if n == 0:
        return 0

    if check_only:
        return n

    # 跳过 n 个 BOM（每个 3 字节）
    stripped = data[n * 3 :]
    try:
        with open(path, "wb") as f:
            f.write(stripped)
    except (OSError, PermissionError) as e:
        print(f"  [SKIP write-error] {path}: {e}", file=sys.stderr)
        return 0
    return n


def walk(root: str, check_only: bool):
    scanned = 0
    fixed = []  # (path, bom_count)

    for dirpath, dirnames, filenames in os.walk(root):
        # in-place 过滤，跳过不需要的目录
        dirnames[:] = [d for d in dirnames if d not in SKIP_DIRS]

        for fname in filenames:
            ext = os.path.splitext(fname)[1].lower()
            if ext not in SOURCE_EXTS:
                continue
            full = os.path.join(dirpath, fname)
            scanned += 1
            n = process_file(full, check_only)
            if n > 0:
                fixed.append((full, n))

    return scanned, fixed


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

    print(f"[strip_bom] root = {root}")
    print(f"[strip_bom] mode = {'check-only' if check_only else 'fix'}")

    scanned, fixed = walk(root, check_only)

    print(f"[strip_bom] scanned {scanned} source files")
    if fixed:
        # 按 BOM 数降序，再按路径排序
        fixed.sort(key=lambda x: (-x[1], x[0]))
        for path, n in fixed:
            print(f"  [{n}x BOM] {path}")
        if check_only:
            print(f"[strip_bom] {len(fixed)} files still have BOM")
            return 1
        print(f"[strip_bom] stripped BOM from {len(fixed)} files")
    else:
        print("[strip_bom] CLEAN: no source files with leading BOM")
    return 0


if __name__ == "__main__":
    sys.exit(main())
