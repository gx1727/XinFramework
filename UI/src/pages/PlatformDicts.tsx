import { useEffect, useState, useCallback, useMemo } from "react"
import { toast } from "sonner"
import { PageLayout } from "@/components/page-layout"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
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
import {
  PlusIcon,
  SearchIcon,
  EditIcon,
  TrashIcon,
  BookIcon,
  RefreshCw,
  HashIcon,
  ListIcon,
  GlobeIcon,
  CheckIcon,
  AlertCircleIcon,
} from "lucide-react"
import { t } from "@/locales"
import { dictApi, type DictItem as Dict, type DictValueItem } from "@/api"
import { FormDialog } from "@/components/schema/DynamicForm"
import type { FormSchema } from "@/types/schema"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { cn } from "@/lib/utils"

type Visibility = "all" | "whitelist" | "blacklist"

export function PlatformDictsPage() {
  const [dicts, setDicts] = useState<Dict[]>([])
  const [searchTerm, setSearchTerm] = useState("")
  const [selectedId, setSelectedId] = useState<number | null>(null)
  const [items, setItems] = useState<DictValueItem[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [isLoadingItems, setIsLoadingItems] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const [dictDialogOpen, setDictDialogOpen] = useState(false)
  const [dictDialogMode, setDictDialogMode] = useState<"add" | "edit">("add")
  const [currentDict, setCurrentDict] = useState<Dict | null>(null)

  const [deleteDictOpen, setDeleteDictOpen] = useState(false)
  const [dictToDelete, setDictToDelete] = useState<Dict | null>(null)

  const [itemDialogOpen, setItemDialogOpen] = useState(false)
  const [itemDialogMode, setItemDialogMode] = useState<"add" | "edit">("add")
  const [currentItem, setCurrentItem] = useState<DictValueItem | null>(null)

  const [deleteItemOpen, setDeleteItemOpen] = useState(false)
  const [itemToDelete, setItemToDelete] = useState<DictValueItem | null>(null)

  const filteredDicts = useMemo(() => {
    if (!searchTerm.trim()) return dicts
    const kw = searchTerm.toLowerCase()
    return dicts.filter(
      (d) => d.code.toLowerCase().includes(kw) || d.name.toLowerCase().includes(kw)
    )
  }, [dicts, searchTerm])

  const selectedDict = useMemo(
    () => dicts.find((d) => d.id === selectedId) ?? null,
    [dicts, selectedId]
  )

  const fetchDicts = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const res = await dictApi.listPlatformDicts({ page: 1, size: 200 })
      const list = res?.list ?? []
      setDicts(list)
      if (list.length > 0 && selectedId == null) {
        setSelectedId(list[0].id)
      }
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e)
      setError(`加载平台字典失败：${msg}`)
      setDicts([])
      setSelectedId(null)
    } finally {
      setIsLoading(false)
    }
  }, [selectedId])

  const fetchItems = useCallback(async (dictId: number) => {
    setIsLoadingItems(true)
    setError(null)
    try {
      const res = await dictApi.listPlatformItems(dictId)
      setItems(res?.list ?? [])
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e)
      setError(`加载字典项失败：${msg}`)
      setItems([])
    } finally {
      setIsLoadingItems(false)
    }
  }, [])

  useEffect(() => {
    fetchDicts()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    if (selectedId != null) {
      fetchItems(selectedId)
    } else {
      setItems([])
    }
  }, [selectedId, fetchItems])

  // ========================================================================
  // 字典 CRUD
  // ========================================================================
  const handleAddDict = () => {
    setDictDialogMode("add")
    setCurrentDict(null)
    setDictDialogOpen(true)
  }

  const handleEditDict = (d: Dict) => {
    setDictDialogMode("edit")
    setCurrentDict(d)
    setDictDialogOpen(true)
  }

  const handleDeleteDictConfirm = (d: Dict) => {
    setDictToDelete(d)
    setDeleteDictOpen(true)
  }

  const handleDeleteDict = async () => {
    if (!dictToDelete) return
    setIsSubmitting(true)
    const deletedName = dictToDelete.name
    const deletedId = dictToDelete.id
    try {
      await dictApi.deletePlatformDict(deletedId)
      toast.success(`字典「${deletedName}」已删除`)
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e)
      console.warn("[delete platform dict] failed:", e)
      setError(`删除字典失败：${msg}`)
      toast.error(`删除字典失败：${msg}`)
    } finally {
      const remaining = dicts.filter((d) => d.id !== deletedId)
      setDicts(remaining)
      if (selectedId === deletedId) {
        setSelectedId(
          remaining.length > 0 ? remaining[0].id : null
        )
      }
      setIsSubmitting(false)
      setDeleteDictOpen(false)
      setDictToDelete(null)
    }
  }

  const dictFormSchema: FormSchema = useMemo(
    () => ({
      items: [
        {
          field: "code",
          label: t.pages.platformDicts?.code || "字典编码",
          type: "text",
          required: true,
          disabled: dictDialogMode === "edit",
          placeholder: "如：gender",
          rules: [{ required: true, message: "请输入字典编码" }, { maxLength: 32 }],
        },
        {
          field: "name",
          label: t.pages.platformDicts?.name || "字典名称",
          type: "text",
          required: true,
          placeholder: "如：性别",
          rules: [{ required: true, message: "请输入字典名称" }, { maxLength: 64 }],
        },
        {
          field: "sort",
          label: t.pages.platformDicts?.sort || "排序",
          type: "number",
          defaultValue: 0,
        },
        {
          field: "visibility",
          label: t.pages.platformDicts?.visibility || "可见性",
          type: "select",
          defaultValue: "all",
          options: [
            { label: t.pages.platformDicts?.visibilityAll || "全部", value: "all" },
            { label: t.pages.platformDicts?.visibilityWhitelist || "白名单", value: "whitelist" },
            { label: t.pages.platformDicts?.visibilityBlacklist || "黑名单", value: "blacklist" },
          ],
        },
      ],
    }),
    [dictDialogMode]
  )

  const getDictInitialValues = (): Record<string, unknown> => {
    if (dictDialogMode === "edit" && currentDict) {
      return {
        code: currentDict.code,
        name: currentDict.name,
        sort: currentDict.sort ?? 0,
        visibility:
          ((currentDict as Dict & { visibility?: Visibility }).visibility) ||
          "all",
      }
    }
    return { sort: 0, visibility: "all" }
  }

  const handleDictSubmit = async (values: Record<string, unknown>) => {
    setIsSubmitting(true)
    try {
      if (dictDialogMode === "add") {
        const payload = {
          code: String(values.code),
          name: String(values.name),
          sort: Number(values.sort ?? 0),
          visibility: (values.visibility ?? "all") as Visibility,
        }
        const created = await dictApi.createPlatformDict(payload)
        setDicts((prev) => [...prev, created])
        setSelectedId(created.id)
        toast.success(`字典「${created.name}」已创建`)
      } else if (currentDict) {
        const payload = {
          name: String(values.name),
          sort: Number(values.sort ?? 0),
          visibility: (values.visibility ?? "all") as Visibility,
        }
        const updated = await dictApi.updatePlatformDict(currentDict.id, payload)
        setDicts((prev) =>
          prev.map((d) => (d.id === currentDict.id ? { ...d, ...updated } : d))
        )
        toast.success(`字典「${updated.name}」已更新`)
      }
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e)
      setError(`保存字典失败：${msg}`)
      toast.error(`保存字典失败：${msg}`)
      throw e
    } finally {
      setIsSubmitting(false)
    }
  }

  // ========================================================================
  // 字典项 CRUD
  // ========================================================================
  const handleAddItem = () => {
    if (!selectedId) return
    setItemDialogMode("add")
    setCurrentItem(null)
    setItemDialogOpen(true)
  }

  const handleEditItem = (it: DictValueItem) => {
    setItemDialogMode("edit")
    setCurrentItem(it)
    setItemDialogOpen(true)
  }

  const handleDeleteItemConfirm = (it: DictValueItem) => {
    setItemToDelete(it)
    setDeleteItemOpen(true)
  }

  const handleDeleteItem = async () => {
    if (!itemToDelete || !selectedId) return
    setIsSubmitting(true)
    const deletedId = itemToDelete.id
    const deletedName = itemToDelete.name
    try {
      await dictApi.deletePlatformItem(selectedId, deletedId)
      toast.success(`字典项「${deletedName}」已删除`)
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e)
      console.warn("[delete platform item] failed:", e)
      setError(`删除字典项失败：${msg}`)
      toast.error(`删除字典项失败：${msg}`)
    } finally {
      setItems((prev) => prev.filter((it) => it.id !== deletedId))
      setDicts((prev) =>
        prev.map((d) =>
          d.id === selectedId
            ? { ...d, item_count: Math.max(0, (d.item_count ?? 0) - 1) }
            : d
        )
      )
      setIsSubmitting(false)
      setDeleteItemOpen(false)
      setItemToDelete(null)
    }
  }

  const itemFormSchema: FormSchema = useMemo(
    () => ({
      items: [
        {
          field: "code",
          label: t.pages.platformDicts?.itemCode || "字典项编码",
          type: "text",
          required: true,
          disabled: itemDialogMode === "edit",
          placeholder: "如：male",
          rules: [{ required: true, message: "请输入字典项编码" }, { maxLength: 64 }],
        },
        {
          field: "name",
          label: t.pages.platformDicts?.itemName || "字典项名称",
          type: "text",
          required: true,
          placeholder: "如：男",
          rules: [{ required: true, message: "请输入字典项名称" }, { maxLength: 128 }],
        },
        {
          field: "sort",
          label: t.pages.platformDicts?.sort || "排序",
          type: "number",
          defaultValue: 0,
        },
      ],
    }),
    [itemDialogMode]
  )

  const getItemInitialValues = (): Record<string, unknown> => {
    if (itemDialogMode === "edit" && currentItem) {
      return {
        code: currentItem.code,
        name: currentItem.name,
        sort: currentItem.sort ?? 0,
      }
    }
    return { sort: items.length + 1 }
  }

  const handleItemSubmit = async (values: Record<string, unknown>) => {
    if (!selectedId) return
    setIsSubmitting(true)
    try {
      if (itemDialogMode === "add") {
        const payload = {
          code: String(values.code),
          name: String(values.name),
          sort: Number(values.sort ?? 0),
        }
        const created = await dictApi.createPlatformItem(selectedId, payload)
        setItems((prev) => [...prev, created])
        setDicts((prev) =>
          prev.map((d) =>
            d.id === selectedId
              ? { ...d, item_count: (d.item_count ?? 0) + 1 }
              : d
          )
        )
        toast.success(`字典项「${created.name}」已创建`)
        setItemDialogOpen(false)
      } else if (currentItem) {
        const payload = {
          name: String(values.name),
          sort: Number(values.sort ?? 0),
        }
        await dictApi.updatePlatformItem(selectedId, currentItem.id, payload)
        setItems((prev) =>
          prev.map((it) =>
            it.id === currentItem.id ? { ...it, ...payload } : it
          )
        )
        toast.success(`字典项「${payload.name}」已更新`)
        setItemDialogOpen(false)
      }
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e)
      setError(`保存字典项失败：${msg}`)
      toast.error(`保存字典项失败：${msg}`)
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <PageLayout>
      <div className="px-4 lg:px-6 space-y-4">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-2">
            <GlobeIcon className="size-5 text-primary" />
            <div>
              <h1 className="text-2xl font-semibold tracking-tight">
                {t.pages.platformDicts?.title || "平台字典"}
              </h1>
              <p className="text-sm text-muted-foreground">
                {t.pages.platformDicts?.subtitle ||
                  "维护 scope=platform 的数据字典与字典项，供所有租户消费"}
              </p>
            </div>
          </div>
        </div>

        {error && (
          <div className="flex items-start gap-2 p-3 rounded-md border border-destructive/40 bg-destructive/5 text-sm">
            <AlertCircleIcon className="h-4 w-4 text-destructive mt-0.5 shrink-0" />
            <div className="flex-1 min-w-0">
              <div className="font-medium text-destructive">接口调用失败</div>
              <div className="text-muted-foreground text-xs mt-0.5 break-all">{error}</div>
              <div className="text-muted-foreground text-xs mt-1">
                提示：请确认后端已启动（<code>http://localhost:8087</code>），
                且当前用户是 <code>super_admin</code>。
              </div>
            </div>
            <Button variant="ghost" size="sm" onClick={fetchDicts}>重试</Button>
          </div>
        )}

        <div className="grid grid-cols-1 lg:grid-cols-[320px_1fr] gap-4">
          <Card className="h-fit">
            <CardHeader className="pb-3">
              <div className="flex items-center justify-between">
                <CardTitle className="flex items-center gap-2 text-base">
                  <BookIcon className="h-4 w-4" />
                  {t.pages.platformDicts?.dictList || "字典列表"}
                </CardTitle>
                <div className="flex gap-1">
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7"
                    onClick={fetchDicts}
                    title="刷新"
                  >
                    <RefreshCw className="h-3.5 w-3.5" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7"
                    onClick={handleAddDict}
                    title="新建字典"
                  >
                    <PlusIcon className="h-3.5 w-3.5" />
                  </Button>
                </div>
              </div>
              <div className="relative pt-1">
                <SearchIcon className="absolute left-2 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
                <Input
                  placeholder={t.pages.platformDicts?.searchPlaceholder || "搜索字典编码或名称..."}
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-7 h-8 text-sm"
                />
              </div>
            </CardHeader>
            <CardContent className="pt-0">
              {isLoading ? (
                <div className="text-center text-sm text-muted-foreground py-8">
                  {t.common.loading}
                </div>
              ) : filteredDicts.length === 0 ? (
                <div className="text-center text-sm text-muted-foreground py-8">
                  {t.common.noData}
                </div>
              ) : (
                <ul className="space-y-1">
                  {filteredDicts.map((d) => {
                    const active = d.id === selectedId
                    return (
                      <li key={d.id}>
                        <button
                          onClick={() => setSelectedId(d.id)}
                          className={cn(
                            "w-full text-left px-3 py-2 rounded-md transition-colors flex items-center gap-2",
                            active
                              ? "bg-primary/10 text-primary"
                              : "hover:bg-accent"
                          )}
                        >
                          <div className="flex-1 min-w-0">
                            <div className="flex items-center gap-2">
                              <span className="font-medium truncate">{d.name}</span>
                              {d.status !== 1 && (
                                <Badge variant="secondary" className="text-[10px] h-4 px-1">
                                  停用
                                </Badge>
                              )}
                            </div>
                            <div className="flex items-center gap-1 text-xs text-muted-foreground mt-0.5">
                              <HashIcon className="h-3 w-3" />
                              <span className="font-mono">{d.code}</span>
                            </div>
                          </div>
                          <Badge variant="outline" className="text-[10px]">
                            {d.item_count ?? 0}
                          </Badge>
                        </button>
                      </li>
                    )
                  })}
                </ul>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle className="flex items-center gap-2 text-base">
                    {selectedDict ? (
                      <>
                        <ListIcon className="h-4 w-4" />
                        {selectedDict.name}
                        <Badge variant="outline" className="ml-1 font-mono text-[10px]">
                          {selectedDict.code}
                        </Badge>
                      </>
                    ) : (
                      <>
                        <ListIcon className="h-4 w-4" />
                        字典项
                      </>
                    )}
                  </CardTitle>
                  <CardDescription>
                    {selectedDict
                      ? `共 ${items.length} 项`
                      : (t.pages.platformDicts?.selectDictFirst || "请先在左侧选择一个字典")}
                  </CardDescription>
                </div>
                <div className="flex gap-2">
                  {selectedDict && (
                    <>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => selectedDict && handleEditDict(selectedDict)}
                      >
                        <EditIcon className="h-3.5 w-3.5 mr-1" />
                        {t.common.edit}
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => selectedDict && handleDeleteDictConfirm(selectedDict)}
                      >
                        <TrashIcon className="h-3.5 w-3.5 mr-1 text-destructive" />
                        {t.common.delete}
                      </Button>
                    </>
                  )}
                  <Button size="sm" onClick={handleAddItem} disabled={!selectedId}>
                    <PlusIcon className="h-3.5 w-3.5 mr-1" />
                    {t.pages.platformDicts?.addItem || "新增字典项"}
                  </Button>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              {!selectedDict ? (
                <div className="text-center text-sm text-muted-foreground py-12">
                  {t.pages.platformDicts?.selectDictFirst || "请先在左侧选择一个字典"}
                </div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>编码</TableHead>
                      <TableHead>名称</TableHead>
                      <TableHead className="w-20">排序</TableHead>
                      <TableHead className="w-20">状态</TableHead>
                      <TableHead className="text-right w-32">操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {items.map((it) => (
                      <TableRow key={it.id}>
                        <TableCell className="font-mono text-sm">
                          <div className="flex items-center gap-1">
                            <HashIcon className="h-3 w-3 text-muted-foreground" />
                            {it.code}
                          </div>
                        </TableCell>
                        <TableCell>{it.name}</TableCell>
                        <TableCell>{it.sort}</TableCell>
                        <TableCell>
                          {it.status === 1 ? (
                            <Badge variant="default" className="text-[10px]">
                              <CheckIcon className="h-3 w-3 mr-0.5" />
                              {t.pages.platformDicts?.enabled || "启用"}
                            </Badge>
                          ) : (
                            <Badge variant="secondary" className="text-[10px]">
                              {t.pages.platformDicts?.disabled || "停用"}
                            </Badge>
                          )}
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
                        <TableCell colSpan={5} className="text-center py-8 text-muted-foreground">
                          {t.common.noData}
                        </TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
              )}
              {isLoadingItems && (
                <div className="flex items-center justify-center py-8">
                  <div className="text-sm text-muted-foreground">{t.common.loading}</div>
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>

      <FormDialog
        open={dictDialogOpen}
        onOpenChange={setDictDialogOpen}
        title={
          dictDialogMode === "add"
            ? t.pages.platformDicts?.addDict || "新建字典"
            : t.pages.platformDicts?.editDict || "编辑字典"
        }
        width={520}
        schema={dictFormSchema}
        initialValues={getDictInitialValues()}
        onSubmit={handleDictSubmit}
        loading={isSubmitting}
      />

      <FormDialog
        open={itemDialogOpen}
        onOpenChange={setItemDialogOpen}
        title={
          itemDialogMode === "add"
            ? t.pages.platformDicts?.addItem || "新增字典项"
            : t.pages.platformDicts?.editItem || "编辑字典项"
        }
        width={480}
        schema={itemFormSchema}
        initialValues={getItemInitialValues()}
        onSubmit={handleItemSubmit}
        loading={isSubmitting}
      />

      <Dialog open={deleteDictOpen} onOpenChange={setDeleteDictOpen}>
        <DialogContent className="sm:max-w-[420px]">
          <DialogHeader>
            <DialogTitle>删除字典</DialogTitle>
            <DialogDescription>
              {`确定要删除字典「${dictToDelete?.name}」吗？字典下还有字典项将无法删除。`}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteDictOpen(false)}>
              {t.common.cancel}
            </Button>
            <Button variant="destructive" onClick={handleDeleteDict} disabled={isSubmitting}>
              {t.common.delete}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={deleteItemOpen} onOpenChange={setDeleteItemOpen}>
        <DialogContent className="sm:max-w-[420px]">
          <DialogHeader>
            <DialogTitle>删除字典项</DialogTitle>
            <DialogDescription>
              {`确定要删除字典项「${itemToDelete?.name}」吗？`}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteItemOpen(false)}>
              {t.common.cancel}
            </Button>
            <Button variant="destructive" onClick={handleDeleteItem} disabled={isSubmitting}>
              {t.common.delete}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </PageLayout>
  )
}
