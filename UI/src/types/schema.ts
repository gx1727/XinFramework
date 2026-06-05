export type FieldType = 
  | "text"
  | "number"
  | "password"
  | "email"
  | "textarea"
  | "select"
  | "radio"
  | "checkbox"
  | "switch"
  | "date"
  | "datetime"
  | "time"
  | "daterange"
  | "datetimerange"
  | "upload"
  | "image"
  | "file"
  | "editor"
  | "divider"
  | "slot"
  | "icon"

export interface ShowIfCondition {
  dependsOn: string
  equals?: unknown
  notEquals?: unknown
  in?: unknown[]
  notIn?: unknown[]
  contains?: unknown
  greaterThan?: number
  lessThan?: number
  and?: ShowIfCondition[]
  or?: ShowIfCondition[]
}

export interface FormItemSchema {
  field: string
  label: string
  type: FieldType
  placeholder?: string
  defaultValue?: unknown
  required?: boolean
  disabled?: boolean
  readonly?: boolean
  rules?: ValidationRule[]
  showIf?: ShowIfCondition
  options?: SelectOption[]
  props?: Record<string, unknown>
  colSpan?: number
  tooltip?: string
}

export interface ValidationRule {
  required?: boolean
  min?: number
  max?: number
  minLength?: number
  maxLength?: number
  pattern?: string
  type?: "string" | "number" | "email" | "url"
  message?: string
}

export interface SelectOption {
  label: string
  value: string | number
  disabled?: boolean
  children?: SelectOption[]
}

export type ColumnType =
  | "text"
  | "number"
  | "date"
  | "datetime"
  | "status"
  | "tag"
  | "badge"
  | "image"
  | "avatar"
  | "switch"
  | "action"
  | "custom"

export interface ColumnSchema {
  key: string
  title: string
  type?: ColumnType
  width?: number | string
  minWidth?: number
  sortable?: boolean
  fixed?: "left" | "right"
  align?: "left" | "center" | "right"
  ellipsis?: boolean
  show?: boolean
  render?: (value: unknown, record: Record<string, unknown>, index: number) => React.ReactNode
  props?: Record<string, unknown>
}

export interface ActionButton {
  key: string
  label: string
  type?: "primary" | "default" | "danger" | "warning"
  size?: "small" | "middle" | "large"
  icon?: string
  permission?: string
  disabled?: boolean
  showIf?: ShowIfCondition
  onClick?: (record: Record<string, unknown>, index: number) => void
}

export interface SearchSchema {
  field: string
  label: string
  type: FieldType
  placeholder?: string
  defaultValue?: unknown
  showIf?: ShowIfCondition
  options?: SelectOption[]
  props?: Record<string, unknown>
}

export interface TableSchema {
  columns: ColumnSchema[]
  actions?: ActionButton[]
  search?: SearchSchema[]
  pagination?: {
    pageSize?: number
    pageSizes?: number[]
    showSizeChanger?: boolean
    showQuickJumper?: boolean
  }
}

export interface FormSchema {
  items: FormItemSchema[]
  layout?: "horizontal" | "vertical" | "inline"
  labelCol?: { span: number }
  wrapperCol?: { span: number }
}

export interface CrudSchema {
  table: TableSchema
  form: FormSchema
  api: {
    list: string
    create?: string
    update?: string
    delete?: string
    detail?: string
  }
}

export interface FormDialogSchema {
  title: string
  width?: number | string
  form: FormSchema
  onSubmit?: (values: Record<string, unknown>) => Promise<void>
}
