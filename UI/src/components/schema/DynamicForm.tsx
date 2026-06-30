import { useState, useEffect } from "react"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Checkbox } from "@/components/ui/checkbox"
import { Switch } from "@/components/ui/switch"
import { Label } from "@/components/ui/label"
import { Separator } from "@/components/ui/separator"
import { IconPicker } from "@/components/ui/icon-picker"
import type { FormSchema, FormItemSchema } from "@/types/schema"
import { evaluateShowIf } from "./showIfEvaluator"

interface DynamicFormProps {
  schema: FormSchema
  initialValues?: Record<string, unknown>
  onSubmit: (values: Record<string, unknown>) => Promise<void>
  onCancel?: () => void
  loading?: boolean
}

export function DynamicForm({
  schema,
  initialValues = {},
  onSubmit,
  onCancel,
  loading = false,
}: DynamicFormProps) {
  const [values, setValues] = useState<Record<string, unknown>>(initialValues)
  const [errors, setErrors] = useState<Record<string, string>>({})

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect -- 初始化/重置表单是约定写法，外部 initialValues 变化需要重新赋值
    setValues(initialValues)
  }, [initialValues])

  const handleChange = (field: string, value: unknown) => {
    setValues((prev) => ({ ...prev, [field]: value }))
    if (errors[field]) {
      setErrors((prev) => {
        const newErrors = { ...prev }
        delete newErrors[field]
        return newErrors
      })
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    const newErrors: Record<string, string> = {}
    for (const item of schema.items) {
      if (item.showIf && !evaluateShowIf(item.showIf, values)) {
        continue
      }

      const value = values[item.field]

      if (
        item.required &&
        (value === undefined || value === null || value === "")
      ) {
        newErrors[item.field] = `${item.label}不能为空`
        continue
      }

      if (item.rules) {
        for (const rule of item.rules) {
          if (
            rule.required &&
            (value === undefined || value === null || value === "")
          ) {
            newErrors[item.field] = rule.message || `${item.label}不能为空`
            break
          }
          if (
            rule.minLength &&
            typeof value === "string" &&
            value.length < rule.minLength
          ) {
            newErrors[item.field] =
              rule.message || `${item.label}长度不能少于${rule.minLength}`
            break
          }
          if (
            rule.maxLength &&
            typeof value === "string" &&
            value.length > rule.maxLength
          ) {
            newErrors[item.field] =
              rule.message || `${item.label}长度不能超过${rule.maxLength}`
            break
          }
          if (rule.pattern) {
            const regex = new RegExp(rule.pattern)
            if (!regex.test(String(value))) {
              newErrors[item.field] = rule.message || `${item.label}格式不正确`
              break
            }
          }
        }
      }
    }

    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors)
      return
    }

    await onSubmit(values)
  }

  const visibleItems = schema.items.filter(
    (item) => !item.showIf || evaluateShowIf(item.showIf, values)
  )

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="grid gap-4">
        {visibleItems.map((item) => (
          <FormField
            key={item.field}
            schema={item}
            value={values[item.field]}
            error={errors[item.field]}
            onChange={(value) => handleChange(item.field, value)}
          />
        ))}
      </div>

      <DialogFooter>
        {onCancel && (
          <Button type="button" variant="outline" onClick={onCancel}>
            取消
          </Button>
        )}
        <Button type="submit" disabled={loading}>
          {loading ? "提交中..." : "提交"}
        </Button>
      </DialogFooter>
    </form>
  )
}

interface FormFieldProps {
  schema: FormItemSchema
  value: unknown
  error?: string
  onChange: (value: unknown) => void
}

function FormField({ schema, value, error, onChange }: FormFieldProps) {
  const colSpan = schema.colSpan || 1

  if (schema.type === "divider") {
    return (
      <div className="col-span-full">
        <Separator />
      </div>
    )
  }

  const fieldContent = () => {
    switch (schema.type) {
      case "text":
      case "password":
      case "email":
        return (
          <Input
            type={schema.type}
            placeholder={schema.placeholder}
            value={String(value ?? "")}
            onChange={(e) => onChange(e.target.value)}
            disabled={schema.disabled || schema.readonly}
          />
        )

      case "number":
        return (
          <Input
            type="number"
            placeholder={schema.placeholder}
            value={value !== undefined ? String(value) : ""}
            onChange={(e) =>
              onChange(e.target.value ? Number(e.target.value) : undefined)
            }
            disabled={schema.disabled}
          />
        )

      case "textarea":
        return (
          <Textarea
            placeholder={schema.placeholder}
            value={String(value ?? "")}
            onChange={(e) => onChange(e.target.value)}
            disabled={schema.disabled}
            {...schema.props}
          />
        )

      case "select":
        return (
          <Select
            value={value !== undefined ? String(value) : ""}
            onValueChange={onChange}
            disabled={schema.disabled}
          >
            <SelectTrigger>
              <SelectValue
                placeholder={schema.placeholder || `请选择${schema.label}`}
              />
            </SelectTrigger>
            <SelectContent>
              {schema.options?.map((option) => (
                <SelectItem
                  key={option.value}
                  value={String(option.value)}
                  disabled={option.disabled}
                >
                  {option.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        )

      case "radio":
        return (
          <div className="flex flex-col gap-2">
            {schema.options?.map((option) => (
              <div
                key={String(option.value)}
                className="flex items-center gap-2"
              >
                <input
                  type="radio"
                  id={`${schema.field}-${option.value}`}
                  name={schema.field}
                  value={String(option.value)}
                  checked={String(value) === String(option.value)}
                  onChange={(e) => {
                    const strValue = e.target.value
                    if (strValue === "true") {
                      onChange(true)
                    } else if (strValue === "false") {
                      onChange(false)
                    } else if (!isNaN(Number(strValue))) {
                      onChange(Number(strValue))
                    } else {
                      onChange(strValue)
                    }
                  }}
                  disabled={schema.disabled || option.disabled}
                  className="accent-primary"
                />
                <Label
                  htmlFor={`${schema.field}-${option.value}`}
                  className="font-normal"
                >
                  {option.label}
                </Label>
              </div>
            ))}
          </div>
        )

      case "checkbox":
        return (
          <div className="flex flex-col gap-2">
            {schema.options?.map((option) => (
              <div key={option.value} className="flex items-center gap-2">
                <Checkbox
                  id={`${schema.field}-${option.value}`}
                  checked={Array.isArray(value) && value.includes(option.value)}
                  onCheckedChange={(checked) => {
                    const currentValue = Array.isArray(value) ? [...value] : []
                    if (checked) {
                      onChange([...currentValue, option.value])
                    } else {
                      onChange(currentValue.filter((v) => v !== option.value))
                    }
                  }}
                  disabled={schema.disabled || option.disabled}
                />
                <Label
                  htmlFor={`${schema.field}-${option.value}`}
                  className="font-normal"
                >
                  {option.label}
                </Label>
              </div>
            ))}
          </div>
        )

      case "switch":
        return (
          <Switch
            checked={Boolean(value)}
            onCheckedChange={onChange}
            disabled={schema.disabled}
          />
        )

      case "date":
      case "datetime":
      case "time":
        return (
          <Input
            type={schema.type === "datetime" ? "datetime-local" : schema.type}
            value={value ? String(value).slice(0, 16) : ""}
            onChange={(e) => onChange(e.target.value)}
            disabled={schema.disabled}
          />
        )

      case "icon":
        return (
          <IconPicker
            value={value as string | undefined}
            onChange={(val) => onChange(val)}
            placeholder={schema.placeholder || "选择图标"}
            disabled={schema.disabled}
          />
        )

      default:
        return (
          <Input
            type="text"
            placeholder={schema.placeholder}
            value={String(value ?? "")}
            onChange={(e) => onChange(e.target.value)}
            disabled={schema.disabled}
          />
        )
    }
  }

  return (
    <div className={`grid gap-2 ${colSpan > 1 ? `col-span-${colSpan}` : ""}`}>
      <div className="flex items-center gap-2">
        <Label htmlFor={schema.field}>
          {schema.label}
          {schema.required && <span className="ml-1 text-destructive">*</span>}
        </Label>
        {schema.tooltip && (
          <span className="text-xs text-muted-foreground">
            ({schema.tooltip})
          </span>
        )}
      </div>
      {fieldContent()}
      {error && <p className="text-sm text-destructive">{error}</p>}
    </div>
  )
}

interface FormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  /**
   * 位于 DialogHeader 内、DialogTitle 之下、表单之上的额外内容。
   * 用于放置只读信息块（例如系统自动生成的 account_id / code）。
   */
  headerExtra?: React.ReactNode
  width?: number | string
  schema: FormSchema
  initialValues?: Record<string, unknown>
  onSubmit: (values: Record<string, unknown>) => Promise<void>
  loading?: boolean
}

export function FormDialog({
  open,
  onOpenChange,
  title,
  headerExtra,
  width = 520,
  schema,
  initialValues,
  onSubmit,
  loading = false,
}: FormDialogProps) {
  const [formKey, setFormKey] = useState(0)

  useEffect(() => {
    if (open) {
      // eslint-disable-next-line react-hooks/set-state-in-effect -- 重置 formKey 使子组件 remount 是约定写法
      setFormKey((prev) => prev + 1)
    }
  }, [open])

  const handleSubmit = async (values: Record<string, unknown>) => {
    await onSubmit(values)
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        className="sm:max-w-[425px]"
        style={{ maxWidth: typeof width === "number" ? `${width}px` : width }}
      >
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          {headerExtra != null ? headerExtra : <DialogDescription />}
        </DialogHeader>
        <DynamicForm
          key={formKey}
          schema={schema}
          initialValues={initialValues}
          onSubmit={handleSubmit}
          onCancel={() => onOpenChange(false)}
          loading={loading}
        />
      </DialogContent>
    </Dialog>
  )
}
