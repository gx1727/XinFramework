import { useRef, useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { UploadIcon, XIcon, LinkIcon } from "lucide-react"
import { assetApi } from "@/api/client"
import { toast } from "sonner"

interface ImageUploadProps {
  value?: string
  onChange: (value: string) => void
  placeholder?: string
  accept?: string
}

/**
 * 通用图片上传组件
 * - 支持点击上传（调 asset API）
 * - 支持外链 URL 直接输入
 * - 支持清空
 */
export function ImageUpload({ value, onChange, placeholder, accept = "image/*" }: ImageUploadProps) {
  const inputRef = useRef<HTMLInputElement>(null)
  const [uploading, setUploading] = useState(false)
  const [showUrlInput, setShowUrlInput] = useState(false)
  const [urlDraft, setUrlDraft] = useState(value || "")

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    setUploading(true)
    try {
      const res = await assetApi.upload(file)
      onChange(res.url)
      toast.success("上传成功")
    } catch (err) {
      console.error("[image-upload] upload failed", err)
      toast.error(err instanceof Error ? err.message : "上传失败")
    } finally {
      setUploading(false)
      if (inputRef.current) inputRef.current.value = ""
    }
  }

  const handleClear = () => {
    onChange("")
    setUrlDraft("")
  }

  const handleApplyUrl = () => {
    onChange(urlDraft)
    setShowUrlInput(false)
  }

  return (
    <div className="flex flex-col gap-2">
      {value ? (
        <div className="flex items-start gap-3">
          <div className="bg-muted flex h-20 w-20 shrink-0 items-center justify-center overflow-hidden rounded-md border">
            <img
              src={value}
              alt="preview"
              className="h-full w-full object-contain"
              onError={(e) => {
                ;(e.target as HTMLImageElement).style.display = "none"
              }}
            />
          </div>
          <div className="flex flex-1 flex-col gap-2">
            <div className="text-muted-foreground break-all text-xs">{value}</div>
            <div className="flex gap-2">
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => inputRef.current?.click()}
                disabled={uploading}
              >
                <UploadIcon className="mr-1 size-3.5" />
                {uploading ? "上传中..." : "重新上传"}
              </Button>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => setShowUrlInput((v) => !v)}
              >
                <LinkIcon className="mr-1 size-3.5" />
                外链 URL
              </Button>
              <Button type="button" variant="ghost" size="sm" onClick={handleClear}>
                <XIcon className="mr-1 size-3.5" />
                清空
              </Button>
            </div>
          </div>
        </div>
      ) : (
        <div className="flex items-center gap-2">
          <Button
            type="button"
            variant="outline"
            onClick={() => inputRef.current?.click()}
            disabled={uploading}
          >
            <UploadIcon className="mr-1 size-3.5" />
            {uploading ? "上传中..." : "点击上传"}
          </Button>
          <Button
            type="button"
            variant="ghost"
            onClick={() => {
              setUrlDraft("")
              setShowUrlInput((v) => !v)
            }}
          >
            <LinkIcon className="mr-1 size-3.5" />
            外链 URL
          </Button>
        </div>
      )}

      {showUrlInput && (
        <div className="flex items-center gap-2">
          <Input
            value={urlDraft}
            onChange={(e) => setUrlDraft(e.target.value)}
            placeholder={placeholder || "https://..."}
            className="flex-1"
          />
          <Button type="button" size="sm" onClick={handleApplyUrl}>
            应用
          </Button>
          <Button type="button" size="sm" variant="ghost" onClick={() => setShowUrlInput(false)}>
            取消
          </Button>
        </div>
      )}

      <input
        ref={inputRef}
        type="file"
        accept={accept}
        className="hidden"
        onChange={handleFileChange}
      />
    </div>
  )
}
