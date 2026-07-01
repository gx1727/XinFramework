// 通用分页条 — 替代各页面手写的"上一页/下一页/页码/每页"组合。
//
// 行为约定：
// - 受控组件：page/size 由调用方持有，组件只通过 onPageChange / onSizeChange 回调
// - onSizeChange 不传则不显示每页选择器
// - hideWhenEmpty=true（默认）时 total=0 整个组件不渲染
// - infoMode:
//   - "range"     → "第 a-b 条 / 共 N 条"（默认，向 Cache / Tenants 等看齐）
//   - "pageInfo"  → "第 X / Y 页"
//   - "total"     → "共 N 条"
// - currentSize 用于"后端分页 + 前端二次过滤"场景：实际渲染条数与 size 不一致时传入
//   （例如 Cache 隐藏系统键后 visibleKeys.length < size）

import * as React from "react"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import { t } from "@/locales"

const DEFAULT_SIZE_OPTIONS = [10, 20, 50, 100] as const

export type PaginationInfoMode = "range" | "pageInfo" | "total"

export interface DataTablePaginationProps {
  page: number
  size: number
  total: number
  onPageChange: (page: number) => void
  onSizeChange?: (size: number) => void
  isLoading?: boolean
  sizeOptions?: readonly number[]
  currentSize?: number
  showFirstLast?: boolean
  hideWhenEmpty?: boolean
  infoMode?: PaginationInfoMode
  extra?: React.ReactNode
  className?: string
}

function getPageNumbers(
  current: number,
  totalPages: number
): (number | "ellipsis")[] {
  if (totalPages <= 7) {
    return Array.from({ length: totalPages }, (_, i) => i + 1)
  }
  const set = new Set<number>([
    1,
    totalPages,
    current,
    current - 1,
    current + 1,
    current - 2,
    current + 2,
  ])
  const sorted = [...set]
    .filter((n) => n >= 1 && n <= totalPages)
    .sort((a, b) => a - b)
  const result: (number | "ellipsis")[] = []
  let prev = 0
  for (const n of sorted) {
    if (prev && n - prev > 1) result.push("ellipsis")
    result.push(n)
    prev = n
  }
  return result
}

export function DataTablePagination({
  page,
  size,
  total,
  onPageChange,
  onSizeChange,
  isLoading = false,
  sizeOptions = DEFAULT_SIZE_OPTIONS,
  currentSize,
  showFirstLast = true,
  hideWhenEmpty = true,
  infoMode = "range",
  extra,
  className,
}: DataTablePaginationProps) {
  if (hideWhenEmpty && total <= 0) return null

  const totalPages = Math.max(1, Math.ceil(total / size))
  const actualSize = currentSize ?? size
  const from = total > 0 ? (page - 1) * size + 1 : 0
  const to = total > 0 ? Math.min((page - 1) * size + actualSize, total) : 0

  const labels = t.common.pagination
  const leftInfo = (() => {
    switch (infoMode) {
      case "pageInfo":
        return labels.pageInfo
          .replace("{page}", String(page))
          .replace("{totalPages}", String(totalPages))
      case "total":
        return labels.total.replace("{total}", String(total))
      case "range":
      default:
        return labels.range
          .replace("{from}", String(from))
          .replace("{to}", String(to))
          .replace("{total}", String(total))
    }
  })()

  const pageNumbers = getPageNumbers(page, totalPages)

  return (
    <div
      className={cn(
        "mt-4 flex flex-wrap items-center justify-between gap-3",
        className
      )}
    >
      <div className="flex items-center gap-2 text-sm text-muted-foreground">
        <span>{leftInfo}</span>
        {extra}
      </div>

      <div className="flex items-center gap-1">
        {showFirstLast && (
          <Button
            variant="outline"
            size="sm"
            disabled={page <= 1 || isLoading}
            onClick={() => onPageChange(1)}
          >
            {labels.first}
          </Button>
        )}
        <Button
          variant="outline"
          size="sm"
          disabled={page <= 1 || isLoading}
          onClick={() => onPageChange(Math.max(1, page - 1))}
        >
          {labels.prev}
        </Button>
        {showFirstLast &&
          pageNumbers.map((n, idx) =>
            n === "ellipsis" ? (
              <span
                key={`e-${idx}`}
                className="px-2 text-sm text-muted-foreground"
              >
                …
              </span>
            ) : (
              <Button
                key={n}
                variant={n === page ? "default" : "outline"}
                size="sm"
                disabled={isLoading}
                onClick={() => onPageChange(n)}
                className="min-w-9"
              >
                {n}
              </Button>
            )
          )}
        <Button
          variant="outline"
          size="sm"
          disabled={page >= totalPages || isLoading}
          onClick={() => onPageChange(Math.min(totalPages, page + 1))}
        >
          {labels.next}
        </Button>
        {showFirstLast && (
          <Button
            variant="outline"
            size="sm"
            disabled={page >= totalPages || isLoading}
            onClick={() => onPageChange(totalPages)}
          >
            {labels.last}
          </Button>
        )}
      </div>

      {onSizeChange && (
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <span>{labels.pageSize}</span>
          <select
            className="h-8 rounded-md border border-input bg-background px-2 text-sm disabled:opacity-50"
            value={size}
            disabled={isLoading}
            onChange={(e) => onSizeChange(Number(e.target.value))}
          >
            {sizeOptions.map((s) => (
              <option key={s} value={s}>
                {s}
              </option>
            ))}
          </select>
        </div>
      )}
    </div>
  )
}
