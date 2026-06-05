import { useCallback, useEffect, useRef, useState } from "react"
import { PageLayout } from "@/components/page-layout"
import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { resolveAssetUrl } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetDescription, SheetFooter } from "@/components/ui/sheet"
import { AlertDialog, AlertDialogContent, AlertDialogHeader, AlertDialogTitle, AlertDialogDescription, AlertDialogFooter, AlertDialogAction, AlertDialogCancel } from "@/components/ui/alert-dialog"
import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog"
import { ImageIcon, PlusIcon, SearchIcon, EditIcon, TrashIcon, UploadIcon, ChevronLeftIcon, ChevronRightIcon } from "lucide-react"
import { toast } from "sonner"
import { useTranslation } from "@/locales"
import { frameApi, frameCategoryApi, assetApi, type FrameItem, type FrameCategoryItem, ApiError } from "@/api"

const mockFrames: FrameItem[] = [
  { id: 1, category_id: 1, name: "程序员专属", description: "适合程序员的简约头像框", preview_url: "/frames/preview/programmer.png", template_url: "/frames/template/programmer.png", type: "public", sort: 1, status: 1 },
  { id: 2, category_id: 1, name: "新年快乐", description: "新年主题头像框", preview_url: "/frames/preview/newyear.png", template_url: "/frames/template/newyear.png", type: "public", sort: 2, status: 1 },
  { id: 3, category_id: 2, name: "校庆100周年", description: "学校100周年庆头像框", preview_url: "/frames/preview/school100.png", template_url: "/frames/template/school100.png", type: "public", sort: 1, status: 1 },
  { id: 4, category_id: 2, name: "毕业季", description: "毕业生专属头像框", preview_url: "/frames/preview/graduation.png", template_url: "/frames/template/graduation.png", type: "public", sort: 2, status: 1 },
  { id: 5, category_id: 3, name: "企业员工", description: "企业定制头像框", preview_url: "/frames/preview/company.png", template_url: "/frames/template/company.png", type: "custom", sort: 1, status: 1 },
]

const mockCategories: FrameCategoryItem[] = [
  { id: 1, code: "emotion", name: "情绪类", type: "emotion", sort: 1, status: 1 },
  { id: 2, code: "school", name: "学校活动", type: "custom", sort: 2, status: 1 },
  { id: 3, code: "company", name: "企业定制", type: "custom", sort: 3, status: 1 },
]

interface FrameFormData {
  id: number | null
  name: string
  description: string
  category_id: string
  type: string
  sort: string
  status: string
  preview_url: string
  preview_file: File | null
  template_url: string
  template_file: File | null
}

const defaultFormData: FrameFormData = {
  id: null,
  name: "",
  description: "",
  category_id: "",
  type: "public",
  sort: "0",
  status: "1",
  preview_url: "",
  preview_file: null,
  template_url: "",
  template_file: null,
}

function ImageUpload({
  label,
  value,
  file,
  onChange,
  onFileChange,
}: {
  label: string
  value: string
  file: File | null
  onChange: (url: string) => void
  onFileChange: (file: File | null) => void
}) {
  const inputRef = useRef<HTMLInputElement>(null)
  const [dragOver, setDragOver] = useState(false)
  const [localPreview, setLocalPreview] = useState<string | null>(null)

  useEffect(() => {
    if (file) {
      const url = URL.createObjectURL(file)
      setLocalPreview(url)
      return () => URL.revokeObjectURL(url)
    } else {
      setLocalPreview(null)
    }
  }, [file])

  const displayUrl = localPreview || resolveAssetUrl(value)

  const handleFiles = useCallback(
    (files: FileList | null) => {
      if (!files || files.length === 0) return
      const f = files[0]
      if (!f.type.startsWith("image/")) return
      onFileChange(f)
      onChange("")
    },
    [onChange, onFileChange]
  )

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault()
      setDragOver(false)
      handleFiles(e.dataTransfer.files)
    },
    [handleFiles]
  )

  const handleClear = () => {
    onFileChange(null)
    onChange("")
  }

  return (
    <div className="flex flex-col gap-2">
      <Label>{label}</Label>
      {displayUrl ? (
        <div className="relative group rounded-lg border overflow-hidden w-32">
          <img src={displayUrl} alt={label} className="w-full aspect-square object-cover" />
          <div className="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-2">
            <Button type="button" variant="secondary" size="sm" onClick={() => inputRef.current?.click()}>
              更换
            </Button>
            <Button type="button" variant="destructive" size="sm" onClick={handleClear}>
              删除
            </Button>
          </div>
        </div>
      ) : (
        <div
          className={`flex flex-col items-center justify-center rounded-lg border-2 border-dashed aspect-square w-32 cursor-pointer transition-colors ${
            dragOver ? "border-primary bg-primary/5" : "border-muted-foreground/25 hover:border-primary/50"
          }`}
          onClick={() => inputRef.current?.click()}
          onDragOver={(e) => { e.preventDefault(); setDragOver(true) }}
          onDragLeave={() => setDragOver(false)}
          onDrop={handleDrop}
        >
          <UploadIcon className="h-8 w-8 text-muted-foreground mb-2" />
          <span className="text-sm text-muted-foreground">点击或拖拽上传图片</span>
        </div>
      )}
      <input ref={inputRef} type="file" accept="image/*" className="hidden" onChange={(e) => handleFiles(e.target.files)} />
    </div>
  )
}

export function FramesPage() {
  const t = useTranslation()
  const [frames, setFrames] = useState<FrameItem[]>([])
  const [categories, setCategories] = useState<FrameCategoryItem[]>([])
  const [selectedCategory, setSelectedCategory] = useState<number | null>(null)
  const [searchTerm, setSearchTerm] = useState("")
  const [isLoading, setIsLoading] = useState(true)
  const [sheetOpen, setSheetOpen] = useState(false)
  const [formData, setFormData] = useState<FrameFormData>(defaultFormData)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<FrameItem | null>(null)
  const [previewImage, setPreviewImage] = useState<string | null>(null)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const pageSize = 10

  const fetchData = useCallback(async () => {
    setIsLoading(true)
    try {
      const [framesRes, categoriesRes] = await Promise.all([
        frameApi.list({ ...(selectedCategory ? { category_id: selectedCategory } : {}), page, size: pageSize }),
        frameCategoryApi.list(),
      ])
      const framesData = framesRes as { list?: FrameItem[]; total?: number } | FrameItem[]
      if (Array.isArray(framesData)) {
        setFrames(framesData)
        setTotal(framesData.length)
      } else {
        setFrames(framesData?.list || [])
        setTotal(framesData?.total || 0)
      }
      const catList = (categoriesRes as FrameCategoryItem[]) || []
      setCategories(catList.length ? catList : mockCategories)
    } catch {
      setFrames(mockFrames)
      setCategories(mockCategories)
    } finally {
      setIsLoading(false)
    }
  }, [selectedCategory, page])

  useEffect(() => {
    setPage(1)
  }, [selectedCategory])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  const filteredFrames = frames.filter((frame) =>
    frame.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    frame.description?.toLowerCase().includes(searchTerm.toLowerCase())
  )

  const totalPages = Math.max(1, Math.ceil(total / pageSize))

  const getCategoryName = (categoryId: number) => {
    const category = categories.find((c) => c.id === categoryId)
    return category?.name || "-"
  }

  const handleOpenCreate = () => {
    setFormData(defaultFormData)
    setSheetOpen(true)
  }

  const handleOpenEdit = (frame: FrameItem) => {
    setFormData({
      id: frame.id,
      name: frame.name,
      description: frame.description || "",
      category_id: frame.category_id ? String(frame.category_id) : "",
      type: frame.type || "public",
      sort: String(frame.sort || 0),
      status: String(frame.status ?? 1),
      preview_url: frame.preview_url || "",
      preview_file: null,
      template_url: frame.template_url || "",
      template_file: null,
    })
    setSheetOpen(true)
  }

  const handleDelete = (frame: FrameItem) => {
    setDeleteTarget(frame)
  }

  const confirmDelete = async () => {
    if (!deleteTarget) return
    try {
      await frameApi.delete(deleteTarget.id)
      setFrames((prev) => prev.filter((f) => f.id !== deleteTarget.id))
    } catch {
      setFrames((prev) => prev.filter((f) => f.id !== deleteTarget.id))
    } finally {
      setDeleteTarget(null)
    }
  }

  const handleSubmit = async () => {
    if (!formData.name.trim() || !formData.category_id) return

    setIsSubmitting(true)
    try {
      let previewUrl = formData.preview_url || undefined
      let templateUrl = formData.template_url || undefined

      if (formData.preview_file) {
        const uploadRes = await assetApi.upload(formData.preview_file)
        previewUrl = uploadRes.url
      }
      if (formData.template_file) {
        const uploadRes = await assetApi.upload(formData.template_file)
        templateUrl = uploadRes.url
      }

      if (formData.id) {
        const payload = {
          category_id: Number(formData.category_id),
          name: formData.name,
          description: formData.description || undefined,
          preview_url: previewUrl,
          template_url: templateUrl,
          type: formData.type,
          sort: Number(formData.sort),
          status: Number(formData.status),
        }
        await frameApi.update(formData.id, payload)
        setFrames((prev) =>
          prev.map((f) => (f.id === formData.id ? { ...f, ...payload, id: f.id } : f))
        )
      } else {
        const payload = {
          name: formData.name,
          category_id: Number(formData.category_id),
          description: formData.description || undefined,
          preview_url: previewUrl,
          template_url: templateUrl,
          type: formData.type || "public",
          sort: Number(formData.sort),
        }
        const res = await frameApi.create(payload as Parameters<typeof frameApi.create>[0])
        const created = res as FrameItem
        if (created?.id) {
          setFrames((prev) => [...prev, created])
        } else {
          setFrames((prev) => [...prev, { ...payload, id: Date.now(), status: 1 } as FrameItem])
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

  const isEditing = formData.id !== null

  return (
    <PageLayout>
      <div className="px-4 lg:px-6">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-2xl font-bold">{t.pages.frames?.title || "相框管理"}</h1>
            <p className="text-sm text-muted-foreground">{t.pages.frames?.subtitle || "管理头像相框模板"}</p>
          </div>
          <Button onClick={handleOpenCreate}>
            <PlusIcon className="mr-2 h-4 w-4" />
            {t.common.add}
          </Button>
        </div>

        <Card>
          <CardHeader>
            <div className="flex items-center gap-4 flex-wrap">
              <div className="relative flex-1 min-w-[200px] max-w-sm">
                <SearchIcon className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  placeholder={t.pages.frames?.searchPlaceholder || "搜索相框..."}
                  className="pl-9"
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                />
              </div>
              <div className="flex items-center gap-2">
                <Button
                  variant={selectedCategory === null ? "default" : "outline"}
                  size="sm"
                  onClick={() => setSelectedCategory(null)}
                >
                  全部
                </Button>
                {categories.map((category) => (
                  <Button
                    key={category.id}
                    variant={selectedCategory === category.id ? "default" : "outline"}
                    size="sm"
                    onClick={() => setSelectedCategory(category.id)}
                  >
                    {category.name}
                  </Button>
                ))}
              </div>
              <Badge variant="secondary">
                共 {total} 个相框
              </Badge>
            </div>
          </CardHeader>
          <CardContent>
            <div className="border rounded-lg overflow-hidden">
              <div className="grid grid-cols-[72px_1fr_100px_100px_80px_60px_100px] gap-4 px-4 py-3 bg-muted/50 text-sm font-medium text-muted-foreground">
                <span>缩略图</span>
                <span>名称</span>
                <span>分类</span>
                <span>类型</span>
                <span>排序</span>
                <span>状态</span>
                <span className="text-right">操作</span>
              </div>
              {isLoading ? (
                <div className="flex items-center justify-center py-8">
                  <div className="text-sm text-muted-foreground">{t.common.loading}</div>
                </div>
              ) : filteredFrames.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">{t.common.noData}</div>
              ) : (
                filteredFrames.map((frame) => {
                  const thumbUrl = resolveAssetUrl(frame.preview_url || frame.template_url)
                  return (
                    <div key={frame.id} className="grid grid-cols-[72px_1fr_100px_100px_80px_60px_100px] gap-4 px-4 py-3 items-center border-t hover:bg-muted/30 transition-colors">
                      <div
                        className={`h-14 w-[72px] rounded overflow-hidden flex items-center justify-center bg-muted ${thumbUrl ? "cursor-pointer" : ""}`}
                        onClick={() => thumbUrl && setPreviewImage(thumbUrl)}
                      >
                        {thumbUrl ? (
                          <img src={thumbUrl} alt={frame.name} className="w-full h-full object-contain" />
                        ) : (
                          <ImageIcon className="h-6 w-6 text-muted-foreground" />
                        )}
                      </div>
                      <div className="min-w-0">
                        <div className="font-medium truncate">{frame.name}</div>
                        {frame.description && (
                          <div className="text-sm text-muted-foreground truncate">{frame.description}</div>
                        )}
                      </div>
                      <span className="text-sm text-muted-foreground">{getCategoryName(frame.category_id)}</span>
                      <Badge variant="outline" className="text-xs">{frame.type}</Badge>
                      <span className="text-sm text-muted-foreground">{frame.sort}</span>
                      <Badge variant={frame.status === 1 ? "default" : "secondary"} className="text-xs">
                        {frame.status === 1 ? "启用" : "停用"}
                      </Badge>
                      <div className="flex gap-1 justify-end">
                        <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => handleOpenEdit(frame)}>
                          <EditIcon className="h-4 w-4" />
                        </Button>
                        <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => handleDelete(frame)}>
                          <TrashIcon className="h-4 w-4 text-destructive" />
                        </Button>
                      </div>
                    </div>
                  )
                })
              )}
            </div>
            {total > 0 && (
              <div className="flex items-center justify-between mt-4">
                <span className="text-sm text-muted-foreground">共 {total} 条</span>
                <div className="flex items-center gap-2">
                  <Button
                    variant="outline"
                    size="icon"
                    className="h-8 w-8"
                    disabled={page <= 1}
                    onClick={() => setPage((p) => Math.max(1, p - 1))}
                  >
                    <ChevronLeftIcon className="h-4 w-4" />
                  </Button>
                  <span className="text-sm">{page} / {totalPages}</span>
                  <Button
                    variant="outline"
                    size="icon"
                    className="h-8 w-8"
                    disabled={page >= totalPages}
                    onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                  >
                    <ChevronRightIcon className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      <Sheet open={sheetOpen} onOpenChange={setSheetOpen}>
        <SheetContent side="right" className="w-full sm:max-w-lg overflow-y-auto">
          <SheetHeader>
            <SheetTitle>{isEditing ? (t.common.edit || "编辑") : (t.common.add || "添加")}</SheetTitle>
            <SheetDescription>{t.pages.frames?.subtitle || "管理头像相框模板"}</SheetDescription>
          </SheetHeader>
          <div className="flex flex-col gap-4 px-4">
            <div className="flex flex-col gap-2">
              <Label htmlFor="frame-name">{t.pages.frames?.nameLabel || "名称"}</Label>
              <Input
                id="frame-name"
                placeholder="请输入相框名称"
                value={formData.name}
                onChange={(e) => setFormData((prev) => ({ ...prev, name: e.target.value }))}
              />
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="frame-description">{t.pages.frames?.descriptionLabel || "描述"}</Label>
              <Input
                id="frame-description"
                placeholder="请输入描述"
                value={formData.description}
                onChange={(e) => setFormData((prev) => ({ ...prev, description: e.target.value }))}
              />
            </div>
            <div className="flex flex-col gap-2">
              <Label>{t.pages.frames?.categoryLabel || "分类"}</Label>
              <Select
                value={formData.category_id}
                onValueChange={(value) => setFormData((prev) => ({ ...prev, category_id: value }))}
              >
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="请选择分类" />
                </SelectTrigger>
                <SelectContent>
                  {categories.map((category) => (
                    <SelectItem key={category.id} value={String(category.id)}>
                      {category.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="flex flex-col gap-2">
              <Label>{t.pages.frames?.typeLabel || "类型"}</Label>
              <Select
                value={formData.type}
                onValueChange={(value) => setFormData((prev) => ({ ...prev, type: value }))}
              >
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="请选择类型" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="public">公开</SelectItem>
                  <SelectItem value="private">私密</SelectItem>
                  <SelectItem value="space">活动空间</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="frame-sort">排序</Label>
              <Input
                id="frame-sort"
                type="number"
                placeholder="0"
                value={formData.sort}
                onChange={(e) => setFormData((prev) => ({ ...prev, sort: e.target.value }))}
              />
            </div>
             <ImageUpload
              label={t.pages.frames?.templateLabel || "模板图片"}
              value={formData.template_url}
              file={formData.template_file}
              onChange={(url) => setFormData((prev) => ({ ...prev, template_url: url }))}
              onFileChange={(file) => setFormData((prev) => ({ ...prev, template_file: file }))}
            />
            <ImageUpload
              label={t.pages.frames?.previewLabel || "预览图"}
              value={formData.preview_url}
              file={formData.preview_file}
              onChange={(url) => setFormData((prev) => ({ ...prev, preview_url: url }))}
              onFileChange={(file) => setFormData((prev) => ({ ...prev, preview_file: file }))}
            />
            <div className="flex flex-col gap-2">
              <Label>{t.pages.frames?.statusLabel || "状态"}</Label>
              <Select
                value={formData.status}
                onValueChange={(value) => setFormData((prev) => ({ ...prev, status: value }))}
              >
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="请选择状态" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="1">启用</SelectItem>
                  <SelectItem value="0">停用</SelectItem>
                </SelectContent>
              </Select>
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
                disabled={isSubmitting || !formData.name.trim() || !formData.category_id}
              >
                {isSubmitting ? (t.common.saving || "保存中...") : (t.common.save || "保存")}
              </Button>
            </div>
          </SheetFooter>
        </SheetContent>
      </Sheet>

      <Dialog open={previewImage !== null} onOpenChange={(open) => { if (!open) setPreviewImage(null) }}>
        <DialogContent className="max-w-3xl p-0 overflow-hidden">
          <DialogTitle className="sr-only">图片预览</DialogTitle>
          {previewImage && (
            <img src={previewImage} alt="预览" className="w-full h-auto" />
          )}
        </DialogContent>
      </Dialog>

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
