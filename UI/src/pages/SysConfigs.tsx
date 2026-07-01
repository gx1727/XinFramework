import { useEffect, useState, useCallback, useMemo } from "react"
import { toast } from "sonner"
import { PageLayout } from "@/components/page-layout"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Switch } from "@/components/ui/switch"
import { Textarea } from "@/components/ui/textarea"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { FormDialog } from "@/components/schema/DynamicForm"
import type { FormSchema } from "@/types/schema"
import {
  PlusIcon,
  SearchIcon,
  EditIcon,
  TrashIcon,
  RefreshCw,
  GlobeIcon,
  HashIcon,
  ListIcon,
  SettingsIcon,
  AlertCircleIcon,
} from "lucide-react"
import { t } from "@/locales"
import {
  configApi,
  type ConfigCategory,
  type ConfigItem,
  type ConfigItemType,
} from "@/api"
import { cn } from "@/lib/utils"

// ============================================================================
// 类型选项（用于 type 下拉）
// ============================================================================
const ITEM_TYPE_OPTIONS: { value: ConfigItemType; label: string }[] = [
  { value: "string", label: t.pages.sysConfigs?.typeString || "字符串" },
  { value: "number", label: t.pages.sysConfigs?.typeNumber || "数字" },
  { value: "boolean", label: t.pages.sysConfigs?.typeBoolean || "布尔" },
  { value: "json", label: t.pages.sysConfigs?.typeJson || "JSON" },
  { value: "image", label: t.pages.sysConfigs?.typeImage || "图片" },
  { value: "color", label: t.pages.sysConfigs?.typeColor || "颜色" },
  { value: "select", label: t.pages.sysConfigs?.typeSelect || "单选" },
  {
    value: "multiselect",
    label: t.pages.sysConfigs?.typeMultiselect || "多选",
  },
  { value: "text", label: t.pages.sysConfigs?.typeText || "长文本" },
  { value: "password", label: t.pages.sysConfigs?.typePassword || "密码" },
]

// ============================================================================
// 工具：把 unknown 值渲染成预览文本
// ============================================================================
function renderValuePreview(v: unknown): string {
  if (v === undefined || v === null) return "-"
  if (typeof v === "string") return v
  try {
    return JSON.stringify(v)
  } catch {
    return String(v)
  }
}

// ============================================================================
// 主组件
// ============================================================================
export function SysConfigsPage() {
  const [groups, setGroups] = useState<ConfigCategory[]>([])
  const [searchTerm, setSearchTerm] = useState("")
  const [selectedGroupId, setSelectedGroupId] = useState<number | null>(null)
  const [items, setItems] = useState<ConfigItem[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [isLoadingItems, setIsLoadingItems] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const [groupDialogOpen, setGroupDialogOpen] = useState(false)
  const [groupDialogMode, setGroupDialogMode] = useState<"add" | "edit">("add")
  const [currentGroup, setCurrentGroup] = useState<ConfigCategory | null>(null)

  const [deleteGroupOpen, setDeleteGroupOpen] = useState(false)
  const [groupToDelete, setGroupToDelete] = useState<ConfigCategory | null>(
    null
  )

  const [itemDialogOpen, setItemDialogOpen] = useState(false)
  const [itemDialogMode, setItemDialogMode] = useState<"add" | "edit">("add")
  const [currentItem, setCurrentItem] = useState<ConfigItem | null>(null)

  const [deleteItemOpen, setDeleteItemOpen] = useState(false)
  const [itemToDelete, setItemToDelete] = useState<ConfigItem | null>(null)

  // ---- fetch ----
  const fetchGroups = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const list = await configApi.listPlatformGroups()
      const arr = Array.isArray(list) ? list : []
      setGroups(arr)
      if (arr.length > 0 && selectedGroupId == null) {
        setSelectedGroupId(arr[0].id)
      }
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e)
      setError(`加载分组失败：${msg}`)
      setGroups([])
    } finally {
      setIsLoading(false)
    }
  }, [selectedGroupId])

  const fetchItems = useCallback(async (groupId: number) => {
    setIsLoadingItems(true)
    try {
      const list = await configApi.listPlatformItems(groupId)
      setItems(Array.isArray(list) ? list : [])
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e)
      setError(`加载配置项失败：${msg}`)
      setItems([])
    } finally {
      setIsLoadingItems(false)
    }
  }, [])

  useEffect(() => {
    fetchGroups()
  }, [fetchGroups])

  useEffect(() => {
    if (selectedGroupId != null) {
      fetchItems(selectedGroupId)
    } else {
      setItems([])
    }
  }, [selectedGroupId, fetchItems])

  const filteredGroups = useMemo(() => {
    if (!searchTerm.trim()) return groups
    const kw = searchTerm.toLowerCase()
    return groups.filter(
      (g) =>
        g.code.toLowerCase().includes(kw) ||
        g.name.toLowerCase().includes(kw) ||
        (g.description ?? "").toLowerCase().includes(kw)
    )
  }, [groups, searchTerm])

  const selectedGroup = useMemo(
    () => groups.find((g) => g.id === selectedGroupId) ?? null,
    [groups, selectedGroupId]
  )

  // ========================================================================
  // 分组 (Group) CRUD
  // ========================================================================
  const handleAddGroup = () => {
    setGroupDialogMode("add")
    setCurrentGroup(null)
    setGroupDialogOpen(true)
  }

  const handleEditGroup = (g: ConfigCategory) => {
    setGroupDialogMode("edit")
    setCurrentGroup(g)
    setGroupDialogOpen(true)
  }

  const handleDeleteGroupConfirm = (g: ConfigCategory) => {
    setGroupToDelete(g)
    setDeleteGroupOpen(true)
  }

  const handleDeleteGroup = async () => {
    if (!groupToDelete) return
    setIsSubmitting(true)
    const deletedName = groupToDelete.name
    try {
      await configApi.deletePlatformGroup(groupToDelete.id)
      toast.success(`分组「${deletedName}」已删除`)
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e)
      console.warn("[delete sys group] failed:", e)
      setError(`删除分组失败：${msg}`)
      toast.error(`删除分组失败：${msg}`)
    } finally {
      const deletedId = groupToDelete.id
      const remaining = groups.filter((g) => g.id !== deletedId)
      setGroups(remaining)
      if (selectedGroupId === deletedId) {
        setSelectedGroupId(remaining.length > 0 ? remaining[0].id : null)
      }
      setIsSubmitting(false)
      setDeleteGroupOpen(false)
      setGroupToDelete(null)
    }
  }

  const groupFormSchema: FormSchema = useMemo(
    () => ({
      items: [
        {
          field: "code",
          label: t.pages.sysConfigs?.code || "分组编码",
          type: "text",
          required: true,
          disabled: groupDialogMode === "edit",
          placeholder: "如：system",
          rules: [
            { required: true, message: "请输入分组编码" },
            { maxLength: 64 },
          ],
        },
        {
          field: "name",
          label: t.pages.sysConfigs?.name || "分组名称",
          type: "text",
          required: true,
          placeholder: "如：系统配置",
          rules: [
            { required: true, message: "请输入分组名称" },
            { maxLength: 64 },
          ],
        },
        {
          field: "description",
          label: t.pages.sysConfigs?.description || "描述",
          type: "textarea",
          placeholder: "可选，分组用途说明",
          props: { rows: 2 },
        },
        {
          field: "sort",
          label: t.pages.sysConfigs?.sort || "排序",
          type: "number",
          defaultValue: 0,
        },
        {
          field: "is_public",
          label: t.pages.sysConfigs?.isPublic || "公开",
          type: "switch",
          defaultValue: false,
        },
      ],
    }),
    [groupDialogMode]
  )

  const getGroupInitialValues = (): Record<string, unknown> => {
    if (groupDialogMode === "edit" && currentGroup) {
      return {
        code: currentGroup.code,
        name: currentGroup.name,
        description: currentGroup.description ?? "",
        sort: currentGroup.sort ?? 0,
        is_public: currentGroup.is_public ?? false,
      }
    }
    return { sort: 0, is_public: false }
  }

  const handleGroupSubmit = async (values: Record<string, unknown>) => {
    setIsSubmitting(true)
    try {
      if (groupDialogMode === "add") {
        const payload = {
          code: String(values.code),
          name: String(values.name),
          description: values.description
            ? String(values.description)
            : undefined,
          sort: Number(values.sort ?? 0),
          is_public: Boolean(values.is_public),
        }
        const created = await configApi.createPlatformGroup(payload)
        setGroups((prev) => [...prev, created])
        setSelectedGroupId(created.id)
        toast.success(`分组「${created.name}」已创建`)
      } else if (currentGroup) {
        const payload = {
          name: String(values.name),
          description: values.description
            ? String(values.description)
            : undefined,
          sort: Number(values.sort ?? 0),
          is_public: Boolean(values.is_public),
        }
        const updated = await configApi.updatePlatformGroup(
          currentGroup.id,
          payload
        )
        setGroups((prev) =>
          prev.map((g) => (g.id === currentGroup.id ? { ...g, ...updated } : g))
        )
        toast.success(`分组「${updated.name}」已更新`)
      }
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e)
      setError(`保存分组失败：${msg}`)
      toast.error(`保存分组失败：${msg}`)
      throw e
    } finally {
      setIsSubmitting(false)
    }
  }

  // ========================================================================
  // 配置项 (Item) CRUD
  // ========================================================================
  const handleAddItem = () => {
    if (!selectedGroupId) return
    setItemDialogMode("add")
    setCurrentItem(null)
    setItemDialogOpen(true)
  }

  const handleEditItem = (it: ConfigItem) => {
    setItemDialogMode("edit")
    setCurrentItem(it)
    setItemDialogOpen(true)
  }

  const handleDeleteItemConfirm = (it: ConfigItem) => {
    setItemToDelete(it)
    setDeleteItemOpen(true)
  }

  const handleDeleteItem = async () => {
    if (!itemToDelete || !selectedGroupId) return
    setIsSubmitting(true)
    const deletedId = itemToDelete.id
    const deletedLabel = itemToDelete.label || itemToDelete.key
    try {
      await configApi.deletePlatformItem(selectedGroupId, deletedId)
      toast.success(`配置项「${deletedLabel}」已删除`)
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e)
      console.warn("[delete sys item] failed:", e)
      setError(`删除配置项失败：${msg}`)
      toast.error(`删除配置项失败：${msg}`)
    } finally {
      setItems((prev) => prev.filter((it) => it.id !== deletedId))
      setIsSubmitting(false)
      setDeleteItemOpen(false)
      setItemToDelete(null)
    }
  }

  const handleItemSubmit = async (values: Record<string, unknown>) => {
    if (!selectedGroupId) return
    setIsSubmitting(true)
    try {
      const type = String(values.type) as ConfigItemType
      const value = coerceValueByType(values.value, type)
      const defaultValue = coerceValueByType(values.defaultValue, type)

      const baseOptions =
        values.options && String(values.options).trim()
          ? safeJsonParse(String(values.options))
          : undefined

      if (itemDialogMode === "add") {
        const payload: Parameters<typeof configApi.createPlatformItem>[1] = {
          key: String(values.key),
          type,
          value,
          default_value: defaultValue,
          label: values.label ? String(values.label) : undefined,
          description: values.description
            ? String(values.description)
            : undefined,
          options: baseOptions,
          sort: Number(values.sort ?? 0),
          is_public: Boolean(values.is_public),
          is_readonly: Boolean(values.is_readonly),
          is_system: Boolean(values.is_system),
        }
        const created = await configApi.createPlatformItem(
          selectedGroupId,
          payload
        )
        setItems((prev) => [...prev, created])
        const label = created.label || created.key
        toast.success(`配置项「${label}」已创建`)
        setItemDialogOpen(false)
      } else if (currentItem) {
        const updatePayload = {
          value,
          label: values.label ? String(values.label) : undefined,
          description: values.description
            ? String(values.description)
            : undefined,
          sort: Number(values.sort ?? 0),
          is_public: Boolean(values.is_public),
          is_readonly: Boolean(values.is_readonly),
        }
        const updated = await configApi.updatePlatformItem(
          selectedGroupId,
          currentItem.id,
          updatePayload
        )
        setItems((prev) =>
          prev.map((it) =>
            it.id === currentItem.id ? { ...it, ...updated } : it
          )
        )
        const label = updated.label || updated.key
        toast.success(`配置项「${label}」已更新`)
        setItemDialogOpen(false)
      }
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e)
      setError(`保存配置项失败：${msg}`)
      toast.error(`保存配置项失败：${msg}`)
    } finally {
      setIsSubmitting(false)
    }
  }

  // ========================================================================
  // 渲染
  // ========================================================================
  return (
    <PageLayout>
      <div className="space-y-4 px-4 lg:px-6">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-2">
            <GlobeIcon className="size-5 text-primary" />
            <div>
              <h1 className="text-2xl font-semibold tracking-tight">
                {t.pages.sysConfigs?.title || "Sys 配置"}
              </h1>
              <p className="text-sm text-muted-foreground">
                {t.pages.sysConfigs?.subtitle ||
                  "维护 scope=sys 的配置分组与配置项，供所有租户消费"}
              </p>
            </div>
          </div>
        </div>

        {error && (
          <div className="flex items-start gap-2 rounded-md border border-destructive/40 bg-destructive/5 p-3 text-sm">
            <AlertCircleIcon className="mt-0.5 h-4 w-4 shrink-0 text-destructive" />
            <div className="min-w-0 flex-1">
              <div className="font-medium text-destructive">接口调用失败</div>
              <div className="mt-0.5 text-xs break-all text-muted-foreground">
                {error}
              </div>
            </div>
            <Button variant="ghost" size="sm" onClick={fetchGroups}>
              重试
            </Button>
          </div>
        )}

        <div className="grid grid-cols-1 gap-4 lg:grid-cols-[320px_1fr]">
          {/* 左：分组列表 */}
          <Card className="h-fit">
            <CardHeader className="pb-3">
              <div className="flex items-center justify-between">
                <CardTitle className="flex items-center gap-2 text-base">
                  <SettingsIcon className="h-4 w-4" />
                  {t.pages.sysConfigs?.groupList || "分组列表"}
                </CardTitle>
                <div className="flex gap-1">
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7"
                    onClick={fetchGroups}
                    title="刷新"
                  >
                    <RefreshCw className="h-3.5 w-3.5" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7"
                    onClick={handleAddGroup}
                    title={t.pages.sysConfigs?.addGroup || "新建分组"}
                  >
                    <PlusIcon className="h-3.5 w-3.5" />
                  </Button>
                </div>
              </div>
              <div className="relative pt-1">
                <SearchIcon className="absolute top-1/2 left-2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground" />
                <Input
                  placeholder={
                    t.pages.sysConfigs?.searchPlaceholder ||
                    "搜索分组编码或名称..."
                  }
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="h-8 pl-7 text-sm"
                />
              </div>
            </CardHeader>
            <CardContent className="pt-0">
              {isLoading ? (
                <div className="py-8 text-center text-sm text-muted-foreground">
                  {t.common.loading}
                </div>
              ) : filteredGroups.length === 0 ? (
                <div className="py-8 text-center text-sm text-muted-foreground">
                  {t.common.noData}
                </div>
              ) : (
                <ul className="space-y-1">
                  {filteredGroups.map((g) => {
                    const active = g.id === selectedGroupId
                    return (
                      <li key={g.id}>
                        <button
                          onClick={() => setSelectedGroupId(g.id)}
                          className={cn(
                            "flex w-full items-center gap-2 rounded-md px-3 py-2 text-left transition-colors",
                            active
                              ? "bg-primary/10 text-primary"
                              : "hover:bg-accent"
                          )}
                        >
                          <div className="min-w-0 flex-1">
                            <div className="flex items-center gap-2">
                              <span className="truncate font-medium">
                                {g.name}
                              </span>
                              {g.is_system && (
                                <Badge
                                  variant="secondary"
                                  className="h-4 px-1 text-[10px]"
                                >
                                  系统
                                </Badge>
                              )}
                              {g.is_public && (
                                <Badge
                                  variant="outline"
                                  className="h-4 px-1 text-[10px]"
                                >
                                  公开
                                </Badge>
                              )}
                              {g.status !== 1 && (
                                <Badge
                                  variant="secondary"
                                  className="h-4 px-1 text-[10px]"
                                >
                                  停用
                                </Badge>
                              )}
                            </div>
                            <div className="mt-0.5 flex items-center gap-1 text-xs text-muted-foreground">
                              <HashIcon className="h-3 w-3" />
                              <span className="font-mono">{g.code}</span>
                            </div>
                          </div>
                        </button>
                      </li>
                    )
                  })}
                </ul>
              )}
            </CardContent>
          </Card>

          {/* 右：选中分组的配置项 */}
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle className="flex items-center gap-2 text-base">
                    {selectedGroup ? (
                      <>
                        <ListIcon className="h-4 w-4" />
                        {selectedGroup.name}
                        <Badge
                          variant="outline"
                          className="ml-1 font-mono text-[10px]"
                        >
                          {selectedGroup.code}
                        </Badge>
                      </>
                    ) : (
                      <>
                        <ListIcon className="h-4 w-4" />
                        {t.pages.sysConfigs?.items || "配置项"}
                      </>
                    )}
                  </CardTitle>
                  <CardDescription>
                    {selectedGroup
                      ? `${t.pages.sysConfigs?.items || "配置项"}（共 ${items.length} 项）`
                      : t.pages.sysConfigs?.selectGroupFirst ||
                        "请先在左侧选择一个分组"}
                  </CardDescription>
                </div>
                <div className="flex gap-2">
                  {selectedGroup && (
                    <>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() =>
                          selectedGroup && handleEditGroup(selectedGroup)
                        }
                      >
                        <EditIcon className="mr-1 h-3.5 w-3.5" />
                        {t.pages.sysConfigs?.editGroup || "编辑分组"}
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() =>
                          selectedGroup &&
                          handleDeleteGroupConfirm(selectedGroup)
                        }
                      >
                        <TrashIcon className="mr-1 h-3.5 w-3.5 text-destructive" />
                        {t.pages.sysConfigs?.deleteGroup || "删除分组"}
                      </Button>
                    </>
                  )}
                  <Button
                    size="sm"
                    onClick={handleAddItem}
                    disabled={!selectedGroupId}
                  >
                    <PlusIcon className="mr-1 h-3.5 w-3.5" />
                    {t.pages.sysConfigs?.addItem || "新增配置项"}
                  </Button>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              {!selectedGroup ? (
                <div className="py-12 text-center text-sm text-muted-foreground">
                  {t.pages.sysConfigs?.selectGroupFirst ||
                    "请先在左侧选择一个分组"}
                </div>
              ) : (
                <>
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Key</TableHead>
                        <TableHead>
                          {t.pages.sysConfigs?.label || "名称"}
                        </TableHead>
                        <TableHead className="w-20">
                          {t.pages.sysConfigs?.type || "类型"}
                        </TableHead>
                        <TableHead>
                          {t.pages.sysConfigs?.value || "当前值"}
                        </TableHead>
                        <TableHead className="w-20">
                          {t.pages.sysConfigs?.sort || "排序"}
                        </TableHead>
                        <TableHead className="w-32">属性</TableHead>
                        <TableHead className="w-32 text-right">
                          {t.common.edit ? "" : ""}
                          操作
                        </TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {items.map((it) => (
                        <TableRow key={it.id}>
                          <TableCell className="font-mono text-sm">
                            <div className="flex items-center gap-1">
                              <HashIcon className="h-3 w-3 text-muted-foreground" />
                              {it.key}
                            </div>
                          </TableCell>
                          <TableCell>{it.label ?? "-"}</TableCell>
                          <TableCell>
                            <Badge variant="outline" className="text-[10px]">
                              {it.type}
                            </Badge>
                          </TableCell>
                          <TableCell className="max-w-[260px]">
                            <span className="block truncate font-mono text-xs text-muted-foreground">
                              {renderValuePreview(it.value)}
                            </span>
                          </TableCell>
                          <TableCell>{it.sort}</TableCell>
                          <TableCell>
                            <div className="flex flex-wrap gap-1">
                              {it.is_public && (
                                <Badge
                                  variant="secondary"
                                  className="text-[10px]"
                                >
                                  公开
                                </Badge>
                              )}
                              {it.is_readonly && (
                                <Badge
                                  variant="secondary"
                                  className="text-[10px]"
                                >
                                  只读
                                </Badge>
                              )}
                              {it.is_system && (
                                <Badge
                                  variant="secondary"
                                  className="text-[10px]"
                                >
                                  系统
                                </Badge>
                              )}
                              {it.status !== 1 && (
                                <Badge
                                  variant="secondary"
                                  className="text-[10px]"
                                >
                                  停用
                                </Badge>
                              )}
                            </div>
                          </TableCell>
                          <TableCell className="text-right">
                            <div className="flex items-center justify-end gap-1">
                              <Button
                                variant="ghost"
                                size="icon"
                                className="h-7 w-7"
                                onClick={() => handleEditItem(it)}
                              >
                                <EditIcon className="h-3.5 w-3.5" />
                              </Button>
                              <Button
                                variant="ghost"
                                size="icon"
                                className="h-7 w-7"
                                onClick={() => handleDeleteItemConfirm(it)}
                              >
                                <TrashIcon className="h-3.5 w-3.5 text-destructive" />
                              </Button>
                            </div>
                          </TableCell>
                        </TableRow>
                      ))}
                      {items.length === 0 && !isLoadingItems && (
                        <TableRow>
                          <TableCell
                            colSpan={7}
                            className="py-8 text-center text-muted-foreground"
                          >
                            {t.common.noData}
                          </TableCell>
                        </TableRow>
                      )}
                    </TableBody>
                  </Table>
                  {isLoadingItems && (
                    <div className="flex items-center justify-center py-8">
                      <div className="text-sm text-muted-foreground">
                        {t.common.loading}
                      </div>
                    </div>
                  )}
                </>
              )}
            </CardContent>
          </Card>
        </div>
      </div>

      <FormDialog
        open={groupDialogOpen}
        onOpenChange={setGroupDialogOpen}
        title={
          groupDialogMode === "add"
            ? t.pages.sysConfigs?.addGroup || "新建分组"
            : t.pages.sysConfigs?.editGroup || "编辑分组"
        }
        width={520}
        schema={groupFormSchema}
        initialValues={getGroupInitialValues()}
        onSubmit={handleGroupSubmit}
        loading={isSubmitting}
      />

      <ItemFormDialog
        open={itemDialogOpen}
        onOpenChange={setItemDialogOpen}
        mode={itemDialogMode}
        item={currentItem}
        onSubmit={handleItemSubmit}
        loading={isSubmitting}
      />

      <Dialog open={deleteGroupOpen} onOpenChange={setDeleteGroupOpen}>
        <DialogContent className="sm:max-w-[420px]">
          <DialogHeader>
            <DialogTitle>
              {t.pages.sysConfigs?.deleteGroup || "删除分组"}
            </DialogTitle>
            <DialogDescription>
              {`确定要删除分组「${groupToDelete?.name}」吗？分组下还有配置项将无法删除。`}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteGroupOpen(false)}>
              {t.common.cancel}
            </Button>
            <Button
              variant="destructive"
              onClick={handleDeleteGroup}
              disabled={isSubmitting}
            >
              {t.common.delete}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={deleteItemOpen} onOpenChange={setDeleteItemOpen}>
        <DialogContent className="sm:max-w-[420px]">
          <DialogHeader>
            <DialogTitle>
              {t.pages.sysConfigs?.deleteItem || "删除配置项"}
            </DialogTitle>
            <DialogDescription>
              {`确定要删除配置项「${itemToDelete?.label ?? itemToDelete?.key}」吗？`}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteItemOpen(false)}>
              {t.common.cancel}
            </Button>
            <Button
              variant="destructive"
              onClick={handleDeleteItem}
              disabled={isSubmitting}
            >
              {t.common.delete}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </PageLayout>
  )
}

// ============================================================================
// 类型转换：根据 ConfigItemType 把 form 中的 unknown 值转成合适形态
// ============================================================================
function coerceValueByType(v: unknown, type: ConfigItemType): unknown {
  if (v === undefined || v === null || v === "") {
    return type === "boolean" ? false : undefined
  }
  switch (type) {
    case "number":
      return typeof v === "number" ? v : Number(v)
    case "boolean":
      if (typeof v === "boolean") return v
      return v === "true" || v === "1" || v === 1
    case "json":
      if (typeof v === "string") return safeJsonParse(v) ?? v
      return v
    case "multiselect":
      if (Array.isArray(v)) return v
      if (typeof v === "string") {
        return v
          .split(",")
          .map((s) => s.trim())
          .filter(Boolean)
      }
      return v
    default:
      return v
  }
}

function safeJsonParse(s: string): unknown {
  try {
    return JSON.parse(s)
  } catch {
    return undefined
  }
}

// ============================================================================
// 配置项编辑对话框（手写：type 决定 value 编辑控件）
// ============================================================================
interface ItemFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  mode: "add" | "edit"
  item: ConfigItem | null
  onSubmit: (values: Record<string, unknown>) => Promise<void>
  loading?: boolean
}

interface ItemFormState {
  key: string
  label: string
  description: string
  type: ConfigItemType
  sort: number
  is_public: boolean
  is_readonly: boolean
  is_system: boolean
  value: unknown
  defaultValue: unknown
  optionsText: string
}

function emptyItemState(): ItemFormState {
  return {
    key: "",
    label: "",
    description: "",
    type: "string",
    sort: 0,
    is_public: false,
    is_readonly: false,
    is_system: false,
    value: undefined,
    defaultValue: undefined,
    optionsText: "",
  }
}

function ItemFormDialog({
  open,
  onOpenChange,
  mode,
  item,
  onSubmit,
  loading,
}: ItemFormDialogProps) {
  const [form, setForm] = useState<ItemFormState>(emptyItemState())
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!open) return
    if (mode === "edit" && item) {
      setForm({
        key: item.key,
        label: item.label ?? "",
        description: item.description ?? "",
        type: (item.type as ConfigItemType) || "string",
        sort: item.sort ?? 0,
        is_public: !!item.is_public,
        is_readonly: !!item.is_readonly,
        is_system: !!item.is_system,
        value: item.value,
        defaultValue: item.default_value,
        optionsText: item.options ? safeJsonStringify(item.options) : "",
      })
    } else {
      setForm(emptyItemState())
    }
    setError(null)
  }, [open, mode, item])

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!form.key.trim()) {
      setError("Key 不能为空")
      return
    }
    setError(null)
    await onSubmit({
      key: form.key.trim(),
      label: form.label,
      description: form.description,
      type: form.type,
      sort: form.sort,
      is_public: form.is_public,
      is_readonly: form.is_readonly,
      is_system: form.is_system,
      value: form.value,
      defaultValue: form.defaultValue,
      options: form.optionsText,
    })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[88vh] overflow-y-auto sm:max-w-[640px]">
        <DialogHeader>
          <DialogTitle>
            {mode === "add"
              ? t.pages.sysConfigs?.addItem || "新增配置项"
              : t.pages.sysConfigs?.editItem || "编辑配置项"}
          </DialogTitle>
          <DialogDescription>
            Key 在编辑时不可改；类型决定了「值」字段的输入控件与后端存储形态。
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={submit} className="space-y-4">
          {error && (
            <div className="rounded-md border border-destructive/40 bg-destructive/5 p-2 text-sm text-destructive">
              {error}
            </div>
          )}

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <label className="text-sm font-medium">Key *</label>
              <Input
                value={form.key}
                onChange={(e) =>
                  setForm((f) => ({ ...f, key: e.target.value }))
                }
                placeholder="如：site.title"
                disabled={mode === "edit"}
                required
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-sm font-medium">
                {t.pages.sysConfigs?.type || "类型"}
              </label>
              <Select
                value={form.type}
                onValueChange={(v) =>
                  setForm((f) => ({
                    ...f,
                    type: v as ConfigItemType,
                    value: undefined,
                    defaultValue: undefined,
                  }))
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {ITEM_TYPE_OPTIONS.map((opt) => (
                    <SelectItem key={opt.value} value={opt.value}>
                      {opt.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <label className="text-sm font-medium">
                {t.pages.sysConfigs?.label || "名称"}
              </label>
              <Input
                value={form.label}
                onChange={(e) =>
                  setForm((f) => ({ ...f, label: e.target.value }))
                }
                placeholder="如：站点名称"
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-sm font-medium">
                {t.pages.sysConfigs?.sort || "排序"}
              </label>
              <Input
                type="number"
                value={String(form.sort)}
                onChange={(e) =>
                  setForm((f) => ({
                    ...f,
                    sort: e.target.value ? Number(e.target.value) : 0,
                  }))
                }
              />
            </div>
          </div>

          <div className="space-y-1.5">
            <label className="text-sm font-medium">
              {t.pages.sysConfigs?.description || "描述"}
            </label>
            <Textarea
              value={form.description}
              onChange={(e) =>
                setForm((f) => ({ ...f, description: e.target.value }))
              }
              placeholder="可选，用途说明"
              rows={2}
            />
          </div>

          {/* value 编辑控件（根据 type 动态渲染）*/}
          <ValueEditor
            label={t.pages.sysConfigs?.value || "当前值"}
            type={form.type}
            value={form.value}
            onChange={(v) => setForm((f) => ({ ...f, value: v }))}
          />
          <ValueEditor
            label={t.pages.sysConfigs?.defaultValue || "默认值"}
            type={form.type}
            value={form.defaultValue}
            onChange={(v) => setForm((f) => ({ ...f, defaultValue: v }))}
          />

          {(form.type === "select" || form.type === "multiselect") && (
            <div className="space-y-1.5">
              <label className="text-sm font-medium">
                {t.pages.sysConfigs?.options || "可选值（JSON 数组）"}
              </label>
              <Textarea
                value={form.optionsText}
                onChange={(e) =>
                  setForm((f) => ({
                    ...f,
                    optionsText: e.target.value,
                  }))
                }
                placeholder='[{"label":"选项1","value":1}]'
                rows={3}
              />
              <p className="text-xs text-muted-foreground">
                格式：[{"{"}label, value{"}"}]，value 为 string | number |
                boolean
              </p>
            </div>
          )}

          <div className="flex flex-wrap gap-6 pt-1">
            <label className="flex items-center gap-2 text-sm">
              <Switch
                checked={form.is_public}
                onCheckedChange={(v) =>
                  setForm((f) => ({ ...f, is_public: v }))
                }
              />
              {t.pages.sysConfigs?.isPublic || "公开"}
            </label>
            <label className="flex items-center gap-2 text-sm">
              <Switch
                checked={form.is_readonly}
                onCheckedChange={(v) =>
                  setForm((f) => ({ ...f, is_readonly: v }))
                }
              />
              {t.pages.sysConfigs?.isReadonly || "只读"}
            </label>
            <label className="flex items-center gap-2 text-sm">
              <Switch
                checked={form.is_system}
                onCheckedChange={(v) =>
                  setForm((f) => ({ ...f, is_system: v }))
                }
                disabled={mode === "edit"}
              />
              {t.pages.sysConfigs?.isSystem || "系统"}
            </label>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              {t.common.cancel}
            </Button>
            <Button type="submit" disabled={loading}>
              {loading ? t.common.saving : t.common.save}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

// ============================================================================
// Value 编辑器：根据 type 渲染不同控件
// ============================================================================
interface ValueEditorProps {
  label: string
  type: ConfigItemType
  value: unknown
  onChange: (v: unknown) => void
}

function ValueEditor({ label, type, value, onChange }: ValueEditorProps) {
  const str = value === undefined || value === null ? "" : String(value)
  const arr = Array.isArray(value) ? value.map(String) : []

  return (
    <div className="space-y-1.5">
      <label className="text-sm font-medium">{label}</label>
      {(type === "string" ||
        type === "image" ||
        type === "text" ||
        type === "password") && (
        <Input
          type={type === "password" ? "password" : "text"}
          value={str}
          onChange={(e) => onChange(e.target.value)}
        />
      )}
      {type === "number" && (
        <Input
          type="number"
          value={str}
          onChange={(e) =>
            onChange(e.target.value ? Number(e.target.value) : undefined)
          }
        />
      )}
      {type === "boolean" && (
        <Switch checked={value === true} onCheckedChange={(v) => onChange(v)} />
      )}
      {type === "color" && (
        <div className="flex items-center gap-2">
          <Input
            type="color"
            value={str || "#000000"}
            onChange={(e) => onChange(e.target.value)}
            className="h-9 w-16 p-1"
          />
          <Input
            value={str}
            onChange={(e) => onChange(e.target.value)}
            placeholder="#RRGGBB"
          />
        </div>
      )}
      {type === "select" && (
        <Select value={str} onValueChange={(v) => onChange(v)}>
          <SelectTrigger>
            <SelectValue placeholder="选择值" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="__null__">（空）</SelectItem>
          </SelectContent>
        </Select>
      )}
      {type === "multiselect" && (
        <Input
          value={arr.join(", ")}
          onChange={(e) =>
            onChange(
              e.target.value
                .split(",")
                .map((s) => s.trim())
                .filter(Boolean)
            )
          }
          placeholder="多个值用逗号分隔"
        />
      )}
      {type === "json" && (
        <Textarea
          value={typeof value === "string" ? value : safeJsonStringify(value)}
          onChange={(e) => onChange(e.target.value)}
          placeholder='{"k":"v"} 或 ["a","b"]'
          rows={4}
        />
      )}
    </div>
  )
}

function safeJsonStringify(v: unknown): string {
  try {
    return JSON.stringify(v, null, 2) ?? ""
  } catch {
    return ""
  }
}
