import { useEffect, useState } from "react"
import { PageLayout } from "@/components/page-layout"
import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import { Label } from "@/components/ui/label"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetDescription, SheetFooter } from "@/components/ui/sheet"
import { AlertDialog, AlertDialogContent, AlertDialogHeader, AlertDialogTitle, AlertDialogDescription, AlertDialogFooter, AlertDialogAction, AlertDialogCancel } from "@/components/ui/alert-dialog"
import { PlusIcon, EditIcon, TrashIcon } from "lucide-react"
import { toast } from "sonner"
import { useTranslation } from "@/locales"
import { frameCategoryApi, type FrameCategoryItem, ApiError } from "@/api"

const mockCategories: FrameCategoryItem[] = []

interface CategoryFormData {
  id: number | null
  code: string
  name: string
  type: string
  sort: string
  status: string
}

const defaultFormData: CategoryFormData = {
  id: null,
  code: "",
  name: "",
  type: "public",
  sort: "0",
  status: "1",
}

export function FrameCategoriesPage() {
  const t = useTranslation()
  const [categories, setCategories] = useState<FrameCategoryItem[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [sheetOpen, setSheetOpen] = useState(false)
  const [formData, setFormData] = useState<CategoryFormData>(defaultFormData)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<FrameCategoryItem | null>(null)

  useEffect(() => {
    fetchCategories()
  }, [])

  const fetchCategories = async () => {
    setIsLoading(true)
    try {
      const res = await frameCategoryApi.list()
      const list = (res as FrameCategoryItem[]) || []
      setCategories(list.length ? list : mockCategories)
    } catch {
      setCategories(mockCategories)
    } finally {
      setIsLoading(false)
    }
  }

  const handleOpenCreate = () => {
    setFormData(defaultFormData)
    setSheetOpen(true)
  }

  const handleOpenEdit = (cat: FrameCategoryItem) => {
    setFormData({
      id: cat.id,
      code: cat.code,
      name: cat.name,
      type: cat.type || "public",
      sort: String(cat.sort || 0),
      status: String(cat.status ?? 1),
    })
    setSheetOpen(true)
  }

  const handleDelete = (cat: FrameCategoryItem) => {
    setDeleteTarget(cat)
  }

  const confirmDelete = async () => {
    if (!deleteTarget) return
    try {
      await frameCategoryApi.delete(deleteTarget.id)
      setCategories((prev) => prev.filter((c) => c.id !== deleteTarget.id))
    } catch {
      setCategories((prev) => prev.filter((c) => c.id !== deleteTarget.id))
    } finally {
      setDeleteTarget(null)
    }
  }

  const handleSubmit = async () => {
    if (!formData.code.trim() || !formData.name.trim()) return
    setIsSubmitting(true)
    try {
      if (formData.id) {
        const payload = {
          id: formData.id,
          code: formData.code,
          name: formData.name,
          type: formData.type,
          sort: Number(formData.sort),
          status: Number(formData.status),
        }
        await frameCategoryApi.update(formData.id, payload)
        setCategories((prev) =>
          prev.map((c) => (c.id === formData.id ? { ...c, ...payload } : c))
        )
      } else {
        const payload = {
          code: formData.code,
          name: formData.name,
          type: formData.type || "public",
          sort: Number(formData.sort),
        }
        const res = await frameCategoryApi.create(payload)
        const created = res as FrameCategoryItem
        if (created?.id) {
          setCategories((prev) => [...prev, created])
        } else {
          setCategories((prev) => [...prev, { ...payload, id: Date.now(), status: 1 } as FrameCategoryItem])
        }
      }
      setSheetOpen(false)
      setFormData(defaultFormData)
      toast.success(isEditing ? "修改成功" : "创建成功")
    } catch (err) {
      const message = err instanceof ApiError ? err.message : "操作失败，请重试"
      toast.error(message)
    } finally {
      setIsSubmitting(false)
    }
  }

  const treeCategories = [...categories].sort((a, b) => a.sort - b.sort)

  const isEditing = formData.id !== null

  return (
    <PageLayout>
      <div className="px-4 lg:px-6">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-2xl font-bold">相框分类</h1>
            <p className="text-sm text-muted-foreground">管理相框分类，支持树形结构</p>
          </div>
          <Button onClick={handleOpenCreate}>
            <PlusIcon className="mr-2 h-4 w-4" />
            {t.common.add}
          </Button>
        </div>

        <Card>
          <CardHeader>
            <Badge variant="secondary">共 {treeCategories.length} 个分类</Badge>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <div className="flex items-center justify-center py-8">
                <div className="text-sm text-muted-foreground">{t.common.loading}</div>
              </div>
            ) : treeCategories.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">{t.common.noData}</div>
            ) : (
              <div className="border rounded-lg overflow-hidden">
                <div className="grid grid-cols-[1fr_120px_120px_100px_100px] gap-4 px-4 py-3 bg-muted/50 text-sm font-medium text-muted-foreground">
                  <span>分类名称</span>
                  <span>编码</span>
                  <span>类型</span>
                  <span>排序</span>
                  <span className="text-right">操作</span>
                </div>
                {treeCategories.map((cat) => (
                  <div key={cat.id} className="border-t">
                    <div className="grid grid-cols-[1fr_120px_120px_100px_100px] gap-4 px-4 py-3 items-center hover:bg-muted/30 transition-colors">
                      <div className="flex items-center gap-2">
                        <span className="font-medium">{cat.name}</span>
                        <Badge variant={cat.status === 1 ? "default" : "secondary"} className="text-xs">
                          {cat.status === 1 ? "启用" : "停用"}
                        </Badge>
                      </div>
                      <span className="text-sm text-muted-foreground font-mono">{cat.code}</span>
                      <span className="text-sm text-muted-foreground">{cat.type}</span>
                      <span className="text-sm text-muted-foreground">{cat.sort}</span>
                      <div className="flex gap-1 justify-end">
                        <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => handleOpenEdit(cat)}>
                          <EditIcon className="h-4 w-4" />
                        </Button>
                        <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => handleDelete(cat)}>
                          <TrashIcon className="h-4 w-4 text-destructive" />
                        </Button>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      <Sheet open={sheetOpen} onOpenChange={setSheetOpen}>
        <SheetContent side="right" className="w-full sm:max-w-lg overflow-y-auto">
          <SheetHeader>
            <SheetTitle>{isEditing ? (t.common.edit || "编辑") : (t.common.add || "添加")}分类</SheetTitle>
            <SheetDescription>管理相框分类信息</SheetDescription>
          </SheetHeader>
          <div className="flex flex-col gap-4 px-4">
            <div className="flex flex-col gap-2">
              <Label htmlFor="cat-code">编码</Label>
              <Input
                id="cat-code"
                placeholder="请输入分类编码"
                value={formData.code}
                onChange={(e) => setFormData((prev) => ({ ...prev, code: e.target.value }))}
              />
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="cat-name">名称</Label>
              <Input
                id="cat-name"
                placeholder="请输入分类名称"
                value={formData.name}
                onChange={(e) => setFormData((prev) => ({ ...prev, name: e.target.value }))}
              />
            </div>
            <div className="flex flex-col gap-2">
              <Label>类型</Label>
              <select
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                value={formData.type}
                onChange={(e) => setFormData((prev) => ({ ...prev, type: e.target.value }))}
              >
                <option value="public">公开</option>
                <option value="theme">主题</option>
              </select>
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="cat-sort">排序</Label>
              <Input
                id="cat-sort"
                type="number"
                placeholder="0"
                value={formData.sort}
                onChange={(e) => setFormData((prev) => ({ ...prev, sort: e.target.value }))}
              />
            </div>
            <div className="flex flex-col gap-2">
              <Label>状态</Label>
              <select
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                value={formData.status}
                onChange={(e) => setFormData((prev) => ({ ...prev, status: e.target.value }))}
              >
                <option value="1">启用</option>
                <option value="0">停用</option>
              </select>
            </div>
          </div>
          <SheetFooter>
            <div className="flex gap-2 w-full">
              <Button variant="outline" className="flex-1" onClick={() => setSheetOpen(false)}>
                {t.common.cancel || "取消"}
              </Button>
              <Button
                className="flex-1"
                onClick={handleSubmit}
                disabled={isSubmitting || !formData.code.trim() || !formData.name.trim()}
              >
                {isSubmitting ? (t.common.saving || "保存中...") : (t.common.save || "保存")}
              </Button>
            </div>
          </SheetFooter>
        </SheetContent>
      </Sheet>

      <AlertDialog open={deleteTarget !== null} onOpenChange={(open) => { if (!open) setDeleteTarget(null) }}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除</AlertDialogTitle>
            <AlertDialogDescription>
              确定要删除「{deleteTarget?.name}」吗？此操作不可撤销。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction variant="destructive" onClick={confirmDelete}>删除</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </PageLayout>
  )
}
