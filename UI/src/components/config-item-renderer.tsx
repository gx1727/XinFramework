import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { Switch } from "@/components/ui/switch"
import { Button } from "@/components/ui/button"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { ImageUpload } from "@/components/image-upload"
import { RotateCcwIcon } from "lucide-react"
import type { ConfigItem } from "@/api"

interface ConfigItemRendererProps {
  item: ConfigItem
  value: unknown
  onChange: (value: unknown) => void
  onReset?: () => void
  disabled?: boolean
}

/**
 * 配置项动态渲染器
 * 根据 item.type 渲染不同的输入控件
 */
export function ConfigItemRenderer({ item, value, onChange, onReset, disabled }: ConfigItemRendererProps) {
  const v = value as never
  const isReadonly = disabled || item.is_readonly
  const description = item.description
  const placeholder =
    (item.validation as { placeholder?: string } | undefined)?.placeholder ||
    (typeof item.default_value === "string" ? item.default_value : "") ||
    ""

  const renderControl = () => {
    switch (item.type) {
      case "string":
      case "password":
        return (
          <Input
            type={item.type === "password" ? "password" : "text"}
            value={(v as string) ?? ""}
            onChange={(e) => onChange(e.target.value)}
            placeholder={placeholder}
            disabled={isReadonly}
            maxLength={(item.validation as { maxLength?: number } | undefined)?.maxLength}
          />
        )

      case "text":
        return (
          <Textarea
            value={(v as string) ?? ""}
            onChange={(e) => onChange(e.target.value)}
            placeholder={placeholder}
            disabled={isReadonly}
            rows={3}
          />
        )

      case "number":
        return (
          <Input
            type="number"
            value={v !== undefined && v !== null ? String(v) : ""}
            onChange={(e) => {
              const s = e.target.value
              if (s === "") {
                onChange(null)
                return
              }
              const n = Number(s)
              if (!isNaN(n)) onChange(n)
            }}
            placeholder={placeholder}
            disabled={isReadonly}
            min={(item.validation as { min?: number } | undefined)?.min}
            max={(item.validation as { max?: number } | undefined)?.max}
          />
        )

      case "boolean":
        return (
          <div className="flex h-10 items-center">
            <Switch
              checked={Boolean(v)}
              onCheckedChange={(c) => onChange(c)}
              disabled={isReadonly}
            />
          </div>
        )

      case "json":
        return (
          <Textarea
            value={typeof v === "string" ? v : v ? JSON.stringify(v, null, 2) : ""}
            onChange={(e) => {
              const s = e.target.value
              // 尝试解析为 JSON；解析失败时存为字符串
              if (s.trim() === "") {
                onChange(null)
                return
              }
              try {
                onChange(JSON.parse(s))
              } catch {
                onChange(s)
              }
            }}
            placeholder='{"key": "value"}'
            disabled={isReadonly}
            rows={5}
            className="font-mono text-sm"
          />
        )

      case "image":
        return (
          <ImageUpload
            value={(v as string) ?? ""}
            onChange={(url) => onChange(url)}
            placeholder={placeholder || "https://example.com/image.png"}
          />
        )

      case "color":
        return (
          <div className="flex items-center gap-2">
            <Input
              type="color"
              value={(v as string) ?? "#000000"}
              onChange={(e) => onChange(e.target.value)}
              disabled={isReadonly}
              className="h-10 w-20 cursor-pointer p-1"
            />
            <Input
              type="text"
              value={(v as string) ?? ""}
              onChange={(e) => onChange(e.target.value)}
              placeholder="#1677ff"
              disabled={isReadonly}
              className="flex-1"
            />
          </div>
        )

      case "select": {
        const options = item.options || []
        const cur = v === undefined || v === null ? "" : String(v)
        return (
          <Select value={cur} onValueChange={(s) => onChange(s)} disabled={isReadonly}>
            <SelectTrigger className="w-full">
              <SelectValue placeholder="请选择" />
            </SelectTrigger>
            <SelectContent>
              {options.map((o) => (
                <SelectItem key={String(o.value)} value={String(o.value)}>
                  {o.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        )
      }

      case "multiselect": {
        const options = item.options || []
        const selected = Array.isArray(v) ? (v as unknown[]).map(String) : []
        return (
          <div className="flex flex-wrap gap-2">
            {options.map((o) => {
              const val = String(o.value)
              const isSelected = selected.includes(val)
              return (
                <Button
                  key={val}
                  type="button"
                  variant={isSelected ? "default" : "outline"}
                  size="sm"
                  disabled={isReadonly}
                  onClick={() => {
                    if (isSelected) {
                      onChange(selected.filter((s) => s !== val))
                    } else {
                      onChange([...selected, val])
                    }
                  }}
                >
                  {o.label}
                </Button>
              )
            })}
          </div>
        )
      }

      default:
        return (
          <div className="text-muted-foreground text-sm">
            未知类型: {item.type}（value: {JSON.stringify(v)}）
          </div>
        )
    }
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Label className="text-sm font-medium">
            {item.label || item.key}
            {item.is_system && (
              <span className="text-muted-foreground ml-2 text-xs">(系统预置)</span>
            )}
          </Label>
          <code className="text-muted-foreground rounded bg-muted px-1.5 py-0.5 text-xs">
            {item.key}
          </code>
        </div>
        {onReset && item.default_value !== undefined && !item.is_readonly && (
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={onReset}
            className="h-7 px-2 text-xs"
            title="恢复默认"
          >
            <RotateCcwIcon className="mr-1 size-3" />
            默认
          </Button>
        )}
      </div>
      {renderControl()}
      {description && (
        <p className="text-muted-foreground text-xs">{description}</p>
      )}
    </div>
  )
}
