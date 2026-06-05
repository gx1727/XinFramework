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
import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog"
import { AlertDialog, AlertDialogContent, AlertDialogHeader, AlertDialogTitle, AlertDialogDescription, AlertDialogFooter, AlertDialogAction, AlertDialogCancel } from "@/components/ui/alert-dialog"
import { ImageIcon, PlusIcon, SearchIcon, EditIcon, TrashIcon, UploadIcon, ChevronLeftIcon, ChevronRightIcon } from "lucide-react"
import { toast } from "sonner"
import { useTranslation } from "@/locales"
import { avatarApi, avatarCategoryApi, assetApi, type AvatarItem, type AvatarCategoryItem, ApiError } from "@/api"

const mockAvatars: AvatarItem[] = []

const mockCategories: AvatarCategoryItem[] = []

interface AvatarFormData {
  id: number | null
  name: string
  category_id: string
  type: string
  is_public: string
  source_file: File | null
  thumbnail_file: File | null
  source_url: string
  thumbnail_url: string
  status: string
}

const defaultFormData: AvatarFormData = {
  id: null,
  name: "",
  category_id: "",
  type: "custom",
  is_public: "true",
  source_file: null,
  thumbnail_file: null,
  source_url: "",
  thumbnail_url: "",
  status: "1",
}

function ImageUpload({
  label,
  file,
  value,
  onFileChange,
  onUrlChange,
}: {
  label: string
  file: File | null
  value: string
  onFileChange: (file: File | null) => void
  onUrlChange: (url: string) => void
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
      onUrlChange("")
    },
    [onFileChange, onUrlChange]
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
    onUrlChange("")
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

export function AvatarsPage() {
  const t = useTranslation()
  const [avatars, setAvatars] = useState<AvatarItem[]>([])
  const [categories, setCategories] = useState<AvatarCategoryItem[]>([])
  const [selectedCategory, setSelectedCategory] = useState<number | null>(null)
  const [searchTerm, setSearchTerm] = useState("")
  const [isLoading, setIsLoading] = useState(true)
  const [sheetOpen, setSheetOpen] = useState(false)
  const [formData, setFormData] = useState<AvatarFormData>(defaultFormData)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<AvatarItem | null>(null)
  const [previewImage, setPreviewImage] = useState<string | null>(null)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const pageSize = 10

  const fetchData = useCallback(async () => {
    setIsLoading(true)
    try {
      const [avatarsRes, categoriesRes] = await Promise.all([
        avatarApi.list({ ...(selectedCategory ? { category_id: selectedCategory } : {}), page, size: pageSize }),
        avatarCategoryApi.list(),
      ])
      const avatarsData = avatarsRes as { list?: AvatarItem[]; total?: number } | AvatarItem[]
      if (Array.isArray(avatarsData)) {
        setAvatars(avatarsData)
        setTotal(avatarsData.length)
      } else {
        setAvatars(avatarsData?.list || [])
        setTotal(avatarsData?.total || 0)
      }
      const catList = (categoriesRes as AvatarCategoryItem[]) || []
      setCategories(catList.length ? catList : mockCategories)
    } catch {
      setAvatars(mockAvatars)
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

  const filteredAvatars = avatars.filter((avatar) =>
    avatar.name.toLowerCase().includes(searchTerm.toLowerCase())
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

  const handleOpenEdit = (avatar: AvatarItem) => {
    setFormData({
      id: avatar.id,
      name: avatar.name,
      category_id: avatar.category_id ? String(avatar.category_id) : "",
      type: avatar.type || "custom",
      is_public: avatar.is_public ? "true" : "false",
      source_file: null,
      thumbnail_file: null,
      source_url: avatar.source_url || "",
      thumbnail_url: avatar.thumbnail_url || "",
      status: String(avatar.status ?? 1),
    })
    setSheetOpen(true)
  }

  const handleDelete = (avatar: AvatarItem) => {
    setDeleteTarget(avatar)
  }

  const confirmDelete = async () => {
    if (!deleteTarget) return
    try {
      await avatarApi.delete(deleteTarget.id)
      setAvatars((prev) => prev.filter((a) => a.id !== deleteTarget.id))
    } catch {
      setAvatars((prev) => prev.filter((a) => a.id !== deleteTarget.id))
    } finally {
      setDeleteTarget(null)
    }
  }

  const handleSubmit = async () => {
    if (!formData.source_file && !formData.source_url) return

    setIsSubmitting(true)
    try {
      let sourceUrl = formData.source_url || undefined
      let thumbnailUrl = formData.thumbnail_url || undefined

      if (formData.source_file) {
        const uploadRes = await assetApi.upload(formData.source_file)
        sourceUrl = uploadRes.url
      }
      if (formData.thumbnail_file) {
        const uploadRes = await assetApi.upload(formData.thumbnail_file)
        thumbnailUrl = uploadRes.url
      }

      if (formData.id) {
        const payload = {
          name: formData.name || undefined,
          category_id: formData.category_id ? Number(formData.category_id) : undefined,
          source_url: sourceUrl!,
          thumbnail_url: thumbnailUrl,
          is_public: formData.is_public === "true",
          status: Number(formData.status),
        }
        await avatarApi.update(formData.id, payload)
        setAvatars((prev) =>
          prev.map((a) =>
            a.id === formData.id
              ? { ...a, source_url: sourceUrl!, category_id: payload.category_id ?? a.category_id, name: payload.name ?? a.name, thumbnail_url: payload.thumbnail_url, is_public: payload.is_public, status: payload.status }
              : a
          )
        )
      } else {
        const payload = {
          source_url: sourceUrl!,
          category_id: formData.category_id ? Number(formData.category_id) : undefined,
          name: formData.name || undefined,
          thumbnail_url: thumbnailUrl,
          file_size: formData.source_file?.size,
          is_public: formData.is_public === "true",
        }
        const res = await avatarApi.create(payload)
        const created = res as AvatarItem
        if (created?.id) {
          setAvatars((prev) => [...prev, created])
        } else {
          setAvatars((prev) => [...prev, {
            id: Date.now(),
            user_id: 1,
            name: formData.name || (formData.source_file?.name || ""),
            source_url: sourceUrl!,
            thumbnail_url: thumbnailUrl,
            file_size: formData.source_file?.size,
            type: formData.type,
            is_public: formData.is_public === "true",
            category_id: formData.category_id ? Number(formData.category_id) : 0,
            status: 1,
          }])
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
            <h1 className="text-2xl font-bold">头像管理</h1>
            <p className="text-sm text-muted-foreground">管理用户头像</p>
          </div>
          <Button onClick={handleOpenCreate}>
            <PlusIcon className="mr-2 h-4 w-4" />
            上传头像
          </Button>
        </div>

        <Card>
          <CardHeader>
            <div className="flex items-center gap-4 flex-wrap">
              <div className="relative flex-1 min-w-[200px] max-w-sm">
                <SearchIcon className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  placeholder="搜索头像..."
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
                共 {total} 个头像
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
              ) : filteredAvatars.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">{t.common.noData}</div>
              ) : (
                filteredAvatars.map((avatar) => {
                  const thumbUrl = resolveAssetUrl(avatar.thumbnail_url || avatar.source_url)
                  return (
                    <div key={avatar.id} className="grid grid-cols-[72px_1fr_100px_100px_80px_60px_100px] gap-4 px-4 py-3 items-center border-t hover:bg-muted/30 transition-colors">
                      <div
                        className={`h-14 w-[72px] rounded overflow-hidden flex items-center justify-center bg-muted ${thumbUrl ? "cursor-pointer" : ""}`}
                        onClick={() => thumbUrl && setPreviewImage(thumbUrl)}
                      >
                        {thumbUrl ? (
                          <img src={thumbUrl} alt={avatar.name} className="w-full h-full object-cover" />
                        ) : (
                          <ImageIcon className="h-6 w-6 text-muted-foreground" />
                        )}
                      </div>
                      <div className="min-w-0">
                        <div className="font-medium truncate">{avatar.name}</div>
                        <div className="flex items-center gap-3 mt-1 text-xs text-muted-foreground">
                          <span>❤ {avatar.like_count || 0}</span>
                          <span>👁 {avatar.view_count || 0}</span>
                        </div>
                      </div>
                      <span className="text-sm text-muted-foreground">{getCategoryName(avatar.category_id)}</span>
                      <Badge variant="outline" className="text-xs">{avatar.type}</Badge>
                      <span className="text-sm text-muted-foreground">{avatar.sort || 0}</span>
                      <Badge variant={avatar.status === 1 ? "default" : "secondary"} className="text-xs">
                        {avatar.status === 1 ? "启用" : "停用"}
                      </Badge>
                      <div className="flex gap-1 justify-end">
                        <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => handleOpenEdit(avatar)}>
                          <EditIcon className="h-4 w-4" />
                        </Button>
                        <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => handleDelete(avatar)}>
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

      <Dialog open={previewImage !== null} onOpenChange={(open) => { if (!open) setPreviewImage(null) }}>
        <DialogContent className="max-w-3xl p-0 overflow-hidden">
          <DialogTitle className="sr-only">图片预览</DialogTitle>
          {previewImage && (
            <img src={previewImage} alt="预览" className="w-full h-auto" />
          )}
        </DialogContent>
      </Dialog>

      <Sheet open={sheetOpen} onOpenChange={setSheetOpen}>
        <SheetContent side="right" className="w-full sm:max-w-lg overflow-y-auto">
          <SheetHeader>
            <SheetTitle>{isEditing ? "编辑头像" : "上传头像"}</SheetTitle>
            <SheetDescription>管理头像信息</SheetDescription>
          </SheetHeader>
          <div className="flex flex-col gap-4 px-4">
            <div className="flex flex-col gap-2">
              <Label htmlFor="avatar-name">名称</Label>
              <Input
                id="avatar-name"
                placeholder="请输入头像名称"
                value={formData.name}
                onChange={(e) => setFormData((prev) => ({ ...prev, name: e.target.value }))}
              />
            </div>
            <div className="flex flex-col gap-2">
              <Label>分类</Label>
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
            <ImageUpload
              label="头像图片"
              file={formData.source_file}
              value={formData.source_url}
              onFileChange={(file) => setFormData((prev) => ({ ...prev, source_file: file }))}
              onUrlChange={(url) => setFormData((prev) => ({ ...prev, source_url: url }))}
            />
            <ImageUpload
              label="缩略图（可选）"
              file={formData.thumbnail_file}
              value={formData.thumbnail_url}
              onFileChange={(file) => setFormData((prev) => ({ ...prev, thumbnail_file: file }))}
              onUrlChange={(url) => setFormData((prev) => ({ ...prev, thumbnail_url: url }))}
            />
            <div className="flex flex-col gap-2">
              <Label>类型</Label>
              <Select
                value={formData.type}
                onValueChange={(value) => setFormData((prev) => ({ ...prev, type: value }))}
              >
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="请选择类型" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="custom">自定义</SelectItem>
                  <SelectItem value="system">系统</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="flex flex-col gap-2">
              <Label>是否公开</Label>
              <Select
                value={formData.is_public}
                onValueChange={(value) => setFormData((prev) => ({ ...prev, is_public: value }))}
              >
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="请选择" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="true">公开</SelectItem>
                  <SelectItem value="false">私密</SelectItem>
                </SelectContent>
              </Select>
            </div>
            {isEditing && (
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
            )}
          </div>
          <SheetFooter>
            <div className="flex gap-2 w-full">
              <Button variant="outline" className="flex-1" onClick={() => setSheetOpen(false)}>
                {t.common.cancel || "取消"}
              </Button>
              <Button
                className="flex-1"
                onClick={handleSubmit}
                disabled={isSubmitting || (!formData.source_file && !formData.source_url)}
              >
                {isSubmitting ? "保存中..." : (t.common.save || "保存")}
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
