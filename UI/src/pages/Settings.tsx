import { useEffect, useState, useMemo, useCallback } from "react"
import { PageLayout } from "@/components/page-layout"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Badge } from "@/components/ui/badge"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Switch } from "@/components/ui/switch"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { useConfigStore } from "@/stores/configStore"
import { ConfigItemRenderer } from "@/components/config-item-renderer"
import type { ConfigGroup, ConfigItem, ConfigItemType } from "@/api/client"
import { PlusIcon, Trash2Icon, SaveIcon, RefreshCwIcon, SettingsIcon } from "lucide-react"
import { toast } from "sonner"
import { cn } from "@/lib/utils"

// 平台支持的 item type 列表（用于"新增项"对话框）
const ITEM_TYPES: { value: ConfigItemType; label: string }[] = [
  { value: "string", label: "字符串" },
  { value: "text", label: "多行文本" },
  { value: "number", label: "数字" },
  { value: "boolean", label: "布尔" },
  { value: "json", label: "JSON" },
  { value: "image", label: "图片" },
  { value: "color", label: "颜色" },
  { value: "select", label: "下拉单选" },
  { value: "multiselect", label: "下拉多选" },
  { value: "password", label: "密码" },
]

export function SettingsPage() {
  const {
    groups,
    groupItems,
    isLoadingGroups,
    isLoadingItems,
    error,
    loadGroups,
    loadItems,
    createGroup,
    deleteGroup,
    createItem,
    updateItem,
    resetItem,
    deleteItem,
  } = useConfigStore()

  const [activeGroupId, setActiveGroupId] = useState<number | null>(null)
  const [dirtyValues, setDirtyValues] = useState<Record<number, unknown>>({})
  const [saving, setSaving] = useState(false)

  // 新增分组对话框
  const [groupDialogOpen, setGroupDialogOpen] = useState(false)
  const [groupDraft, setGroupDraft] = useState<{
    code: string
    name: string
    description: string
    icon: string
    is_public: boolean
  }>({ code: "", name: "", description: "", icon: "SettingsIcon", is_public: false })

  // 新增项对话框
  const [itemDialogOpen, setItemDialogOpen] = useState(false)
  const [itemDraft, setItemDraft] = useState<{
    key: string
    type: ConfigItemType
    label: string
    description: string
    default_value: string
    is_public: boolean
  }>({ key: "", type: "string", label: "", description: "", default_value: "", is_public: false })

  // 删除确认
  const [deleteGroupTarget, setDeleteGroupTarget] = useState<ConfigGroup | null>(null)
  const [deleteItemTarget, setDeleteItemTarget] = useState<ConfigItem | null>(null)

  // 初次加载
  useEffect(() => {
    void loadGroups().then((gs) => {
      if (gs.length > 0 && activeGroupId === null) {
        const first = gs[0]
        setActiveGroupId(first.id)
        void loadItems(first.id)
      }
    })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // 切换 tab
  const handleTabChange = useCallback(
    (id: string) => {
      const gid = Number(id)
      setActiveGroupId(gid)
      setDirtyValues({})
      if (!groupItems[gid]) {
        void loadItems(gid)
      }
    },
    [groupItems, loadItems]
  )

  // 改 item value（暂存 dirty）
  const handleItemChange = (itemId: number, value: unknown) => {
    setDirtyValues((prev) => ({ ...prev, [itemId]: value }))
  }

  // 单项保存
  const handleSaveItem = async (item: ConfigItem) => {
    if (!(item.id in dirtyValues)) return
    setSaving(true)
    try {
      await updateItem(item.id, { value: dirtyValues[item.id] })
      setDirtyValues((prev) => {
        const next = { ...prev }
        delete next[item.id]
        return next
      })
      toast.success(`${item.label || item.key} 已保存`)
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "保存失败")
    } finally {
      setSaving(false)
    }
  }

  // 整组保存
  const handleSaveAll = async () => {
    if (!activeGroupId) return
    const items = groupItems[activeGroupId] || []
    const dirtyItems = items.filter((it) => it.id in dirtyValues)
    if (dirtyItems.length === 0) {
      toast.info("没有未保存的修改")
      return
    }
    setSaving(true)
    try {
      for (const it of dirtyItems) {
        await updateItem(it.id, { value: dirtyValues[it.id] })
      }
      setDirtyValues({})
      toast.success(`已保存 ${dirtyItems.length} 项`)
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "保存失败")
    } finally {
      setSaving(false)
    }
  }

  // 单项 reset
  const handleResetItem = async (item: ConfigItem) => {
    setSaving(true)
    try {
      const updated = await resetItem(item.id)
      setDirtyValues((prev) => {
        const next = { ...prev }
        delete next[item.id]
        return next
      })
      toast.success(`已恢复 ${updated.label || updated.key} 默认值`)
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "重置失败")
    } finally {
      setSaving(false)
    }
  }

  // 新增分组
  const handleCreateGroup = async () => {
    if (!groupDraft.code || !groupDraft.name) {
      toast.error("编码和名称不能为空")
      return
    }
    try {
      const g = await createGroup(groupDraft)
      setGroupDialogOpen(false)
      setGroupDraft({ code: "", name: "", description: "", icon: "SettingsIcon", is_public: false })
      toast.success("分组已创建")
      setActiveGroupId(g.id)
      void loadItems(g.id)
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "创建失败")
    }
  }

  // 新增项
  const handleCreateItem = async () => {
    if (!activeGroupId) return
    if (!itemDraft.key) {
      toast.error("key 不能为空")
      return
    }
    try {
      let defVal: unknown = itemDraft.default_value
      if (itemDraft.type === "number") {
        defVal = itemDraft.default_value ? Number(itemDraft.default_value) : 0
      } else if (itemDraft.type === "boolean") {
        defVal = itemDraft.default_value === "true"
      } else if (itemDraft.type === "json") {
        try {
          defVal = itemDraft.default_value ? JSON.parse(itemDraft.default_value) : null
        } catch {
          toast.error("默认值 JSON 解析失败")
          return
        }
      }
      await createItem(activeGroupId, {
        key: itemDraft.key,
        type: itemDraft.type,
        label: itemDraft.label || undefined,
        description: itemDraft.description || undefined,
        default_value: defVal,
        value: defVal,
        is_public: itemDraft.is_public,
      })
      setItemDialogOpen(false)
      setItemDraft({ key: "", type: "string", label: "", description: "", default_value: "", is_public: false })
      toast.success("配置项已创建")
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "创建失败")
    }
  }

  // 删除分组
  const handleDeleteGroup = async () => {
    if (!deleteGroupTarget) return
    try {
      await deleteGroup(deleteGroupTarget.id)
      toast.success("分组已删除")
      setDeleteGroupTarget(null)
      // 切换到第一个分组
      const remaining = groups.filter((g) => g.id !== deleteGroupTarget.id)
      if (remaining.length > 0) {
        setActiveGroupId(remaining[0].id)
        void loadItems(remaining[0].id)
      } else {
        setActiveGroupId(null)
      }
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "删除失败")
    }
  }

  // 删除项
  const handleDeleteItem = async () => {
    if (!deleteItemTarget) return
    try {
      await deleteItem(deleteItemTarget.id)
      toast.success("配置项已删除")
      setDeleteItemTarget(null)
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "删除失败")
    }
  }

  const activeGroup = useMemo(
    () => groups.find((g) => g.id === activeGroupId),
    [groups, activeGroupId]
  )
  const activeItems = useMemo(() => {
    if (!activeGroupId) return []
    return groupItems[activeGroupId] || []
  }, [activeGroupId, groupItems])
  const dirtyCount = useMemo(() => {
    if (!activeGroupId) return 0
    return activeItems.filter((it) => it.id in dirtyValues).length
  }, [activeItems, dirtyValues, activeGroupId])

  return (
    <PageLayout>
      <div className="px-4 lg:px-6">
        {/* 顶部 header */}
        <div className="mb-6 flex items-start justify-between">
          <div>
            <h1 className="text-2xl font-bold">配置管理</h1>
            <p className="text-muted-foreground text-sm">维护系统级配置项，支持多分组、多种类型、租户可改值</p>
          </div>
          <div className="flex gap-2">
            <Button variant="outline" size="sm" onClick={() => void loadGroups()}>
              <RefreshCwIcon className="mr-1 size-4" />
              刷新
            </Button>
            <Button size="sm" onClick={() => setGroupDialogOpen(true)}>
              <PlusIcon className="mr-1 size-4" />
              新建分组
            </Button>
          </div>
        </div>

        {error && (
          <div className="bg-destructive/10 text-destructive mb-4 rounded-md p-3 text-sm">
            {error}
          </div>
        )}

        {isLoadingGroups && groups.length === 0 ? (
          <Card>
            <CardContent className="text-muted-foreground py-12 text-center text-sm">加载中...</CardContent>
          </Card>
        ) : groups.length === 0 ? (
          <Card>
            <CardContent className="text-muted-foreground py-12 text-center text-sm">
              暂无配置分组，点击右上角"新建分组"开始
            </CardContent>
          </Card>
        ) : (
          <Tabs
            value={activeGroupId !== null ? String(activeGroupId) : undefined}
            onValueChange={handleTabChange}
            className="space-y-4"
          >
            <TabsList className="flex-wrap">
              {groups.map((g) => (
                <TabsTrigger key={g.id} value={String(g.id)} className="gap-2">
                  <SettingsIcon className="size-3.5" />
                  {g.name}
                  {g.is_system && (
                    <span className="text-muted-foreground text-[10px]">系统</span>
                  )}
                </TabsTrigger>
              ))}
            </TabsList>

            {groups.map((g) => (
              <TabsContent key={g.id} value={String(g.id)} className="space-y-4">
                <Card>
                  <CardHeader className="flex flex-row items-start justify-between space-y-0">
                    <div>
                      <CardTitle className="flex items-center gap-2">
                        {g.name}
                        {g.is_system && <Badge variant="secondary">系统预置</Badge>}
                        {g.is_public && <Badge variant="outline">公共</Badge>}
                      </CardTitle>
                      <CardDescription>
                        {g.description || `分组编码：${g.code}`}
                      </CardDescription>
                    </div>
                    <div className="flex gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => {
                          setActiveGroupId(g.id)
                          setItemDialogOpen(true)
                        }}
                      >
                        <PlusIcon className="mr-1 size-4" />
                        新增项
                      </Button>
                      {!g.is_system && (
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setDeleteGroupTarget(g)}
                        >
                          <Trash2Icon className="text-destructive size-4" />
                        </Button>
                      )}
                    </div>
                  </CardHeader>
                  <CardContent className="space-y-6">
                    {isLoadingItems && activeItems.length === 0 ? (
                      <div className="text-muted-foreground py-8 text-center text-sm">加载配置项...</div>
                    ) : activeItems.length === 0 ? (
                      <div className="text-muted-foreground py-8 text-center text-sm">
                        该分组下暂无配置项
                      </div>
                    ) : (
                      <>
                        {activeItems.map((item) => {
                          const dirty = item.id in dirtyValues
                          const value = dirty ? dirtyValues[item.id] : item.value
                          return (
                            <div
                              key={item.id}
                              className={cn(
                                "rounded-lg border p-4 transition-colors",
                                dirty && "border-primary bg-primary/5"
                              )}
                            >
                              <div className="mb-3 flex items-start justify-between">
                                <div className="flex-1">
                                  <ConfigItemRenderer
                                    item={item}
                                    value={value}
                                    onChange={(v) => handleItemChange(item.id, v)}
                                    onReset={() => void handleResetItem(item)}
                                  />
                                </div>
                                <div className="ml-4 flex shrink-0 flex-col items-end gap-2">
                                  {item.is_readonly && (
                                    <Badge variant="secondary" className="text-xs">只读</Badge>
                                  )}
                                  {item.is_public && (
                                    <Badge variant="outline" className="text-xs">公共</Badge>
                                  )}
                                  <div className="flex gap-2">
                                    {dirty ? (
                                      <>
                                        <Button
                                          size="sm"
                                          variant="outline"
                                          onClick={() => {
                                            setDirtyValues((prev) => {
                                              const next = { ...prev }
                                              delete next[item.id]
                                              return next
                                            })
                                          }}
                                        >
                                          取消
                                        </Button>
                                        <Button
                                          size="sm"
                                          onClick={() => void handleSaveItem(item)}
                                          disabled={saving}
                                        >
                                          <SaveIcon className="mr-1 size-3.5" />
                                          保存
                                        </Button>
                                      </>
                                    ) : null}
                                    {!item.is_system && (
                                      <Button
                                        size="sm"
                                        variant="ghost"
                                        onClick={() => setDeleteItemTarget(item)}
                                      >
                                        <Trash2Icon className="text-destructive size-4" />
                                      </Button>
                                    )}
                                  </div>
                                </div>
                              </div>
                            </div>
                          )
                        })}

                        {/* 整组保存按钮 */}
                        {dirtyCount > 0 && (
                          <div className="bg-primary/10 sticky bottom-0 flex items-center justify-between rounded-md p-3">
                            <span className="text-sm font-medium">
                              {dirtyCount} 项未保存
                            </span>
                            <div className="flex gap-2">
                              <Button
                                variant="outline"
                                size="sm"
                                onClick={() => setDirtyValues({})}
                              >
                                全部取消
                              </Button>
                              <Button
                                size="sm"
                                onClick={() => void handleSaveAll()}
                                disabled={saving}
                              >
                                <SaveIcon className="mr-1 size-4" />
                                保存全部
                              </Button>
                            </div>
                          </div>
                        )}
                      </>
                    )}
                  </CardContent>
                </Card>
              </TabsContent>
            ))}
          </Tabs>
        )}
      </div>

      {/* 新建分组对话框 */}
      <Dialog open={groupDialogOpen} onOpenChange={setGroupDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>新建配置分组</DialogTitle>
            <DialogDescription>分组是配置的逻辑分类，例：site / email / security</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label>编码 (code)</Label>
              <Input
                value={groupDraft.code}
                onChange={(e) => setGroupDraft({ ...groupDraft, code: e.target.value })}
                placeholder="my_group"
                maxLength={64}
              />
              <p className="text-muted-foreground mt-1 text-xs">英文，唯一不可改</p>
            </div>
            <div>
              <Label>名称</Label>
              <Input
                value={groupDraft.name}
                onChange={(e) => setGroupDraft({ ...groupDraft, name: e.target.value })}
                placeholder="我的分组"
                maxLength={64}
              />
            </div>
            <div>
              <Label>描述</Label>
              <Input
                value={groupDraft.description}
                onChange={(e) => setGroupDraft({ ...groupDraft, description: e.target.value })}
                placeholder="可选"
                maxLength={255}
              />
            </div>
            <div>
              <Label>图标 (lucide 名称)</Label>
              <Input
                value={groupDraft.icon}
                onChange={(e) => setGroupDraft({ ...groupDraft, icon: e.target.value })}
                placeholder="SettingsIcon"
                maxLength={64}
              />
            </div>
            <div className="flex items-center justify-between">
              <div>
                <Label>公共读</Label>
                <p className="text-muted-foreground text-xs">开启后该分组下 is_public 项未登录可读</p>
              </div>
              <Switch
                checked={groupDraft.is_public}
                onCheckedChange={(c) => setGroupDraft({ ...groupDraft, is_public: c })}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setGroupDialogOpen(false)}>
              取消
            </Button>
            <Button onClick={() => void handleCreateGroup()}>创建</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 新增项对话框 */}
      <Dialog open={itemDialogOpen} onOpenChange={setItemDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>新增配置项</DialogTitle>
            <DialogDescription>
              {activeGroup ? `分组：${activeGroup.name}` : ""}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label>Key</Label>
              <Input
                value={itemDraft.key}
                onChange={(e) => setItemDraft({ ...itemDraft, key: e.target.value })}
                placeholder="my_setting"
                maxLength={128}
              />
              <p className="text-muted-foreground mt-1 text-xs">英文，唯一不可改</p>
            </div>
            <div>
              <Label>类型</Label>
              <Select
                value={itemDraft.type}
                onValueChange={(v) => setItemDraft({ ...itemDraft, type: v as ConfigItemType })}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {ITEM_TYPES.map((t) => (
                    <SelectItem key={t.value} value={t.value}>
                      {t.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div>
              <Label>显示名</Label>
              <Input
                value={itemDraft.label}
                onChange={(e) => setItemDraft({ ...itemDraft, label: e.target.value })}
                placeholder="我的设置"
                maxLength={128}
              />
            </div>
            <div>
              <Label>描述</Label>
              <Input
                value={itemDraft.description}
                onChange={(e) => setItemDraft({ ...itemDraft, description: e.target.value })}
                placeholder="可选"
                maxLength={512}
              />
            </div>
            <div>
              <Label>默认值</Label>
              <Input
                value={itemDraft.default_value}
                onChange={(e) => setItemDraft({ ...itemDraft, default_value: e.target.value })}
                placeholder={
                  itemDraft.type === "boolean"
                    ? "true / false"
                    : itemDraft.type === "json"
                      ? '{"key":"value"}'
                      : "可选"
                }
              />
            </div>
            <div className="flex items-center justify-between">
              <div>
                <Label>公共</Label>
                <p className="text-muted-foreground text-xs">未登录是否可读</p>
              </div>
              <Switch
                checked={itemDraft.is_public}
                onCheckedChange={(c) => setItemDraft({ ...itemDraft, is_public: c })}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setItemDialogOpen(false)}>
              取消
            </Button>
            <Button onClick={() => void handleCreateItem()}>创建</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 删除分组确认 */}
      <AlertDialog open={!!deleteGroupTarget} onOpenChange={(o) => !o && setDeleteGroupTarget(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>删除分组？</AlertDialogTitle>
            <AlertDialogDescription>
              将删除分组「{deleteGroupTarget?.name}」。该分组下若有配置项，删除将被拒绝。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={() => void handleDeleteGroup()}>
              确认删除
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* 删除项确认 */}
      <AlertDialog open={!!deleteItemTarget} onOpenChange={(o) => !o && setDeleteItemTarget(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>删除配置项？</AlertDialogTitle>
            <AlertDialogDescription>
              将删除配置项「{deleteItemTarget?.key}」。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={() => void handleDeleteItem()}>
              确认删除
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </PageLayout>
  )
}
