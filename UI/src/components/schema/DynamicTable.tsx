import { useState, useCallback } from "react"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Input } from "@/components/ui/input"
import { Skeleton } from "@/components/ui/skeleton"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  ChevronLeftIcon,
  ChevronRightIcon,
} from "lucide-react"
import type { TableSchema, ColumnSchema, SearchSchema } from "@/types/schema"
import type { ShowIfCondition } from "@/types/schema"
import { evaluateShowIf } from "./showIfEvaluator"

interface DynamicTableProps {
  schema: TableSchema
  data: Record<string, unknown>[]
  loading?: boolean
  pagination?: {
    current: number
    pageSize: number
    total: number
    onChange: (page: number, pageSize: number) => void
  }
  rowSelection?: {
    selectedRowKeys: string[]
    onChange: (keys: string[]) => void
  }
  onAction?: (key: string, record: Record<string, unknown>) => void
  searchValues?: Record<string, unknown>
  onSearch?: (values: Record<string, unknown>) => void
}

export function DynamicTable({
  schema,
  data,
  loading = false,
  pagination,
  rowSelection,
  onAction,
  searchValues = {},
  onSearch,
}: DynamicTableProps) {
  const [localSearchValues, setLocalSearchValues] = useState<Record<string, unknown>>(searchValues)

  const handleSearch = useCallback(() => {
    onSearch?.(localSearchValues)
  }, [localSearchValues, onSearch])

  const handleReset = useCallback(() => {
    const resetValues: Record<string, unknown> = {}
    schema.search?.forEach((item) => {
      if (item.defaultValue !== undefined) {
        resetValues[item.field] = item.defaultValue
      }
    })
    setLocalSearchValues(resetValues)
    onSearch?.(resetValues)
  }, [schema.search, onSearch])

  const visibleColumns = schema.columns.filter((col) => col.show !== false)

  return (
    <div className="space-y-4">
      {schema.search && schema.search.length > 0 && (
        <div className="flex flex-wrap gap-4 rounded-lg border p-4">
          {schema.search.map((item) => (
            <SearchField
              key={item.field}
              schema={item}
              value={localSearchValues[item.field]}
              onChange={(value) =>
                setLocalSearchValues((prev) => ({ ...prev, [item.field]: value }))
              }
            />
          ))}
          <div className="flex items-end gap-2">
            <Button onClick={handleSearch}>搜索</Button>
            <Button variant="outline" onClick={handleReset}>
              重置
            </Button>
          </div>
        </div>
      )}

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              {rowSelection && (
                <TableHead className="w-[50px]">
                  <Checkbox
                    checked={
                      rowSelection.selectedRowKeys.length > 0 &&
                      rowSelection.selectedRowKeys.length === data.length
                    }
                    onCheckedChange={(checked) => {
                      rowSelection.onChange(checked ? data.map((_, i) => String(i)) : [])
                    }}
                  />
                </TableHead>
              )}
              {visibleColumns.map((column) => (
                <TableHead
                  key={column.key}
                  style={{ width: column.width, textAlign: column.align }}
                >
                  {column.title}
                </TableHead>
              ))}
              {schema.actions && schema.actions.length > 0 && (
                <TableHead className="w-[200px]">操作</TableHead>
              )}
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              Array.from({ length: 5 }).map((_, i) => (
                <TableRow key={i}>
                  {rowSelection && (
                    <TableCell>
                      <Skeleton className="h-4 w-4" />
                    </TableCell>
                  )}
                  {visibleColumns.map((col) => (
                    <TableCell key={col.key}>
                      <Skeleton className="h-4 w-[100px]" />
                    </TableCell>
                  ))}
                  {schema.actions && (
                    <TableCell>
                      <Skeleton className="h-8 w-[120px]" />
                    </TableCell>
                  )}
                </TableRow>
              ))
            ) : data.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={
                    visibleColumns.length +
                    (rowSelection ? 1 : 0) +
                    (schema.actions ? 1 : 0)
                  }
                  className="h-24 text-center"
                >
                  暂无数据
                </TableCell>
              </TableRow>
            ) : (
              data.map((record, index) => (
                <TableRow key={index}>
                  {rowSelection && (
                    <TableCell>
                      <Checkbox
                        checked={rowSelection.selectedRowKeys.includes(String(index))}
                        onCheckedChange={(checked) => {
                          const keys = checked
                            ? [...rowSelection.selectedRowKeys, String(index)]
                            : rowSelection.selectedRowKeys.filter((k) => k !== String(index))
                          rowSelection.onChange(keys)
                        }}
                      />
                    </TableCell>
                  )}
                  {visibleColumns.map((column) => (
                    <TableCell
                      key={column.key}
                      style={{ textAlign: column.align }}
                    >
                      <ColumnRenderer
                        column={column}
                        value={record[column.key]}
                        record={record}
                        index={index}
                      />
                    </TableCell>
                  ))}
                  {schema.actions && (
                    <TableCell>
                      <div className="flex gap-2">
                        {schema.actions.map((action) => (
                          <ActionButton
                            key={action.key}
                            action={action}
                            record={record}
                            onClick={() => onAction?.(action.key, record)}
                          />
                        ))}
                      </div>
                    </TableCell>
                  )}
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {pagination && (
        <div className="flex items-center justify-between">
          <div className="text-sm text-muted-foreground">
            共 {pagination.total} 条
          </div>
          <div className="flex items-center gap-2">
            <Select
              value={String(pagination.pageSize)}
              onValueChange={(value) =>
                pagination.onChange(1, Number(value))
              }
            >
              <SelectTrigger className="w-[100px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {[10, 20, 50, 100].map((size) => (
                  <SelectItem key={size} value={String(size)}>
                    {size} 条/页
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Button
              variant="outline"
              size="icon"
              disabled={pagination.current <= 1}
              onClick={() =>
                pagination.onChange(pagination.current - 1, pagination.pageSize)
              }
            >
              <ChevronLeftIcon className="h-4 w-4" />
            </Button>
            <span className="text-sm">
              {pagination.current} / {Math.ceil(pagination.total / pagination.pageSize)}
            </span>
            <Button
              variant="outline"
              size="icon"
              disabled={
                pagination.current >=
                Math.ceil(pagination.total / pagination.pageSize)
              }
              onClick={() =>
                pagination.onChange(pagination.current + 1, pagination.pageSize)
              }
            >
              <ChevronRightIcon className="h-4 w-4" />
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}

function SearchField({
  schema,
  value,
  onChange,
}: {
  schema: SearchSchema
  value: unknown
  onChange: (value: unknown) => void
}) {
  if (schema.showIf && !evaluateShowIf(schema.showIf, {})) {
    return null
  }

  switch (schema.type) {
    case "select":
      return (
        <div className="flex flex-col gap-1">
          <label className="text-sm font-medium">{schema.label}</label>
          <Select value={String(value || "")} onValueChange={onChange}>
            <SelectTrigger className="w-[200px]">
              <SelectValue placeholder={schema.placeholder} />
            </SelectTrigger>
            <SelectContent>
              {schema.options?.map((option) => (
                <SelectItem key={option.value} value={String(option.value)}>
                  {option.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      )
    default:
      return (
        <div className="flex flex-col gap-1">
          <label className="text-sm font-medium">{schema.label}</label>
          <Input
            type={schema.type === "number" ? "number" : "text"}
            placeholder={schema.placeholder}
            value={String(value || "")}
            onChange={(e) => onChange(e.target.value)}
            className="w-[200px]"
          />
        </div>
      )
  }
}

function ColumnRenderer({
  column,
  value,
  record,
  index,
}: {
  column: ColumnSchema
  value: unknown
  record: Record<string, unknown>
  index: number
}) {
  if (column.render) {
    return column.render(value, record, index) as React.ReactElement
  }

  switch (column.type) {
    case "tag":
    case "badge":
      return <Badge variant="secondary">{String(value)}</Badge>
    case "status":
      return (
        <Badge variant={value === "active" ? "default" : "outline"}>
          {value === "active" ? "启用" : "禁用"}
        </Badge>
      )
    case "image":
      return value ? (
        <img
          src={String(value)}
          alt=""
          className="h-8 w-8 rounded object-cover"
        />
      ) : null
    case "avatar":
      return value ? (
        <div className="flex h-8 w-8 items-center justify-center rounded-full bg-muted">
          {String(value).charAt(0).toUpperCase()}
        </div>
      ) : null
    case "switch":
      return (
        <Checkbox checked={Boolean(value)} disabled />
      )
    default:
      return <span className={column.ellipsis ? "truncate" : ""}>{String(value ?? "-")}</span>
  }
}

function ActionButton({
  action,
  record,
  onClick,
}: {
  action: {
    key: string
    label: string
    type?: "primary" | "default" | "danger" | "warning"
    disabled?: boolean
    showIf?: ShowIfCondition
  }
  record: Record<string, unknown>
  onClick: () => void
}) {
  if (action.showIf && !evaluateShowIf(action.showIf, record)) {
    return null
  }

  return (
    <Button
      variant={action.type === "danger" ? "destructive" : "outline"}
      size="sm"
      disabled={action.disabled}
      onClick={onClick}
    >
      {action.label}
    </Button>
  )
}
