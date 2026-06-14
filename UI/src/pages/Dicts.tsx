import { useEffect, useState, useCallback, useMemo } from "react"
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
  CheckIcon,
  AlertCircleIcon,
  DatabaseIcon,
} from "lucide-react"
import { useTranslation } from "@/locales"
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
import { Checkbox } from "@/components/ui/checkbox"

// Mock fallback data
const mockDicts: Dict[] = [
  { id: 1, code: "gender", name: "性别", sort: 1, status: 1, item_count: 2, created_at: "2026-04-26 10:00:00" },
  { id: 2, code: "user_status", name: "用户状态", sort: 2, status: 1, item_count: 2, created_at: "2026-04-26 10:00:00" },
  { id: 3, code: "education", name: "学历", sort: 3, status: 1, item_count: 3, created_at: "2026-04-26 10:00:00" },
]

const mockItems: Record<number, DictValueItem[]> = {
  1: [
    { id: 11, dict_id: 1, code: "male", name: "男", sort: 1, status: 1 },
    { id: 12, dict_id: 1, code: "female", name: "女", sort: 2, status: 1 },
  ],
  2: [
    { id: 21, dict_id: 2, code: "active", name: "启用", sort: 1, status: 1 },
    { id: 22, dict_id: 2, code: "disabled", name: "停用", sort: 2, status: 1 },
  ],
  3: [
    { id: 31, dict_id: 3, code: "bachelor", name: "本科", sort: 1, status: 1 },
    { id: 32, dict_id: 3, code: "master", name: "硕士", sort: 2, status: 1 },
    { id: 33, dict_id: 3, code: "doctor", name: "博士", sort: 3, status: 1 },
  ],
}

export function DictsPage() {
  const t = useTranslation()
  const [dicts, setDicts] = useState<Dict[]>([])
  const [searchTerm, setSearchTerm] = useState("")
  const [selectedId, setSelectedId] = useState<number | null>(null)
  const [items, setItems] = useState<DictValueItem[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [isLoadingItems, setIsLoadingItems] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [dataSource, setDataSource] = useState<"api" | "mock" | null>(null)
  const [useMockFallback, setUseMockFallback] = useState(() => {
    if (typeof window === "undefined") return false
    return localStorage.getItem("dict_use_mock") === "true"
  })

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
      const res = await dictApi.list({ page: 1, size: 200 })
      const list = res?.list ?? []
      setDicts(list)
      setDataSource("api")
      if (list.length > 0 && selectedId == null) {
        setSelectedId(list[0].id)
      }
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e)
      setError(`加载字典失败：${msg}`)
      if (useMockFallback) {
        setDicts(mockDicts)
        setDataSource("mock")
        if (mockDicts.length > 0 && selectedId == null) {
          setSelectedId(mockDicts[0].id)
        }
      } else {
        setDicts([])
        setSelectedId(null)
      }
    } finally {
      setIsLoading(false)
    }
  }, [selectedId, useMockFallback])

  // dev 开关同步到 localStorage
  useEffect(() => {
    if (typeof window !== "undefined") {
      localStorage.setItem("dict_use_mock", useMockFallback ? "true" : "false")
      if (useMockFallback) {
        // 切到 mock 时立即加载
        fetchDicts()
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [useMockFallback])

  const fetchItems = useCallback(async (dictId: number) => {
    setIsLoadingItems(true)
    setError(null)
    try {
      const res = await dictApi.listItems(dictId)
      setItems(res?.list ?? [])
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e)
      setError(`加载字典项失败：${msg}`)
      if (useMockFallback) {
        setItems(mockItems[dictId] ?? [])
      } else {
        setItems([])
      }
    } finally {
      setIsLoadingItems(false)
    }
  }, [useMockFallback])

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
    try {
      await dictApi.delete(dictToDelete.id)
    } catch (e) {
      console.warn("[delete dict] failed:", e)
    } finally {
      setDicts((prev) => prev.filter((d) => d.id !== dictToDelete.id))
      if (selectedId === dictToDelete.id) {
        setSelectedId(() => {
          const remaining = dicts.filter((d) => d.id !== dictToDelete.id)
          return remaining.length > 0 ? remaining[0].id : null
        })
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
          label: t.pages.dicts?.code || "字典编码",
          type: "text",
          required: true,
          disabled: dictDialogMode === "edit",
          placeholder: "如：gender",
          rules: [{ required: true, message: "请输入字典编码" }, { maxLength: 32 }],
        },
        {
          field: "name",
          label: t.pages.dicts?.name || "字典名称",
          type: "text",
          required: true,
          placeholder: "如：性别",
          rules: [{ required: true, message: "请输入字典名称" }, { maxLength: 64 }],
        },
        {
          field: "sort",
          label: t.pages.dicts?.sort || "排序",
          type: "number",
          defaultValue: 0,
        },
      ],
    }),
    [dictDialogMode, t]
  )

  const getDictInitialValues = (): Record<string, unknown> => {
    if (dictDialogMode === "edit" && currentDict) {
      return {
        code: currentDict.code,
        name: currentDict.name,
        sort: currentDict.sort ?? 0,
      }
    }
    return { sort: 0 }
  }

  const handleDictSubmit = async (values: Record<string, unknown>) => {
    setIsSubmitting(true)
    try {
      if (dictDialogMode === "add") {
        const payload = {
          code: String(values.code),
          name: String(values.name),
          sort: Number(values.sort ?? 0),
        }
        try {
          const created = await dictApi.create(payload)
          setDicts((prev) => [...prev, created])
          setSelectedId(created.id)
        } catch {
          const fake: Dict = {
            id: Date.now(),
            code: payload.code,
            name: payload.name,
            sort: payload.sort,
            status: 1,
            item_count: 0,
          }
          setDicts((prev) => [...prev, fake])
          setSelectedId(fake.id)
        }
      } else if (currentDict) {
        const payload = {
          name: String(values.name),
          sort: Number(values.sort ?? 0),
        }
        try {
          await dictApi.update(currentDict.id, payload)
        } catch {
          // no-op
        }
        setDicts((prev) =>
          prev.map((d) => (d.id === currentDict.id ? { ...d, ...payload } : d))
        )
      }
    } finally {
      setIsSubmitting(false)
    }
  }

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
    try {
      await dictApi.deleteItem(selectedId, itemToDelete.id)
    } catch (e) {
      console.warn("[delete item] failed:", e)
    } finally {
      setItems((prev) => prev.filter((it) => it.id !== itemToDelete.id))
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
          label: t.pages.dicts?.itemCode || "字典项编码",
          type: "text",
          required: true,
          disabled: itemDialogMode === "edit",
          placeholder: "如：male",
          rules: [{ required: true, message: "请输入字典项编码" }, { maxLength: 64 }],
        },
        {
          field: "name",
          label: t.pages.dicts?.itemName || "字典项名称",
          type: "text",
          required: true,
          placeholder: "如：男",
          rules: [{ required: true, message: "请输入字典项名称" }, { maxLength: 128 }],
        },
        {
          field: "sort",
          label: t.pages.dicts?.sort || "排序",
          type: "number",
          defaultValue: 0,
        },
      ],
    }),
    [itemDialogMode, t]
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
        try {
          const created = await dictApi.createItem(selectedId, payload)
          setItems((prev) => [...prev, created])
        } catch {
          const fake: DictValueItem = {
            id: Date.now(),
            dict_id: selectedId,
            code: payload.code,
            name: payload.name,
            sort: payload.sort,
            status: 1,
          }
          setItems((prev) => [...prev, fake])
        }
        setDicts((prev) =>
          prev.map((d) =>
            d.id === selectedId
              ? { ...d, item_count: (d.item_count ?? 0) + 1 }
              : d
          )
        )
      } else if (currentItem) {
        const payload = {
          name: String(values.name),
          sort: Number(values.sort ?? 0),
        }
        try {
          await dictApi.updateItem(selectedId, currentItem.id, payload)
        } catch {
          // no-op
        }
        setItems((prev) =>
          prev.map((it) => (it.id === currentItem.id ? { ...it, ...payload } : it))
        )
      }
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <PageLayout>
      <div className="px-4 lg:px-6 space-y-4">
        <div className="flex items-start justify-between">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">
              {t.pages.dicts?.title || "数据字典"}
            </h1>
            <p className="text-sm text-muted-foreground">
              {t.pages.dicts?.subtitle || "维护系统下拉/枚举等数据字典及其字典项"}
            </p>
          </div>
          <div className="flex items-center gap-3">
            {dataSource && (
              <Badge variant={dataSource === "api" ? "default" : "secondary"} className="text-[10px]">
                <DatabaseIcon className="h-3 w-3 mr-0.5" />
                {dataSource === "api" ? "实时数据" : "Mock 数据（开发模式）"}
              </Badge>
            )}
            <label className="flex items-center gap-1.5 text-xs text-muted-foreground cursor-pointer select-none">
              <Checkbox
                checked={useMockFallback}
                onCheckedChange={(v) => setUseMockFallback(Boolean(v))}
              />
              失败时使用 mock
            </label>
          </div>
        </div>
        {error && (
          <div className="flex items-start gap-2 p-3 rounded-md border border-destructive/40 bg-destructive/5 text-sm">
            <AlertCircleIcon className="h-4 w-4 text-destructive mt-0.5 shrink-0" />
            <div className="flex-1 min-w-0">
              <div className="font-medium text-destructive">接口调用失败</div>
              <div className="text-muted-foreground text-xs mt-0.5 break-all">{error}</div>
              <div className="text-muted-foreground text-xs mt-1">
                提示：请确认后端服务已启动（<code className="text-xs">http://localhost:8087</code>），
                且当前用户拥有 <code className="text-xs">dict:list</code> 权限。
                开启「失败时使用 mock」可在后端未启时继续演示。
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
                  {t.pages.dicts?.dictList || "字典列表"}
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
                  placeholder={t.pages.dicts?.searchPlaceholder || "搜索字典编码或名称..."}
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
                      : "请先在左侧选择一个字典"}
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
                    {t.pages.dicts?.addItem || "新增字典项"}
                  </Button>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              {!selectedDict ? (
                <div className="text-center text-sm text-muted-foreground py-12">
                  请先在左侧选择一个字典
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
                              启用
                            </Badge>
                          ) : (
                            <Badge variant="secondary" className="text-[10px]">
                              停用
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
        title={dictDialogMode === "add" ? "新建字典" : "编辑字典"}
        width={480}
        schema={dictFormSchema}
        initialValues={getDictInitialValues()}
        onSubmit={handleDictSubmit}
        loading={isSubmitting}
      />

      <FormDialog
        open={itemDialogOpen}
        onOpenChange={setItemDialogOpen}
        title={itemDialogMode === "add" ? "新增字典项" : "编辑字典项"}
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