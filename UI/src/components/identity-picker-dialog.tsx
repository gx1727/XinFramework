// IdentityPickerDialog 多身份账号登录时弹出"选择身份"对话框（路径 B 多身份支持）。
//
// 流程：用户在 LoginForm 提交账号密码 → loginPrecheck 返回 N 个 tenant 身份
// （+ 可选 platform 角色） → LoginForm 弹此 Dialog → 用户点选 → 调 selectTenant
// 或 platformLogin 完成登录。

import { useNavigate } from "react-router-dom"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { BuildingIcon, ShieldCheckIcon, UserIcon } from "lucide-react"
import type { TenantIdentity } from "@/api"
import { cn } from "@/lib/utils"

interface IdentityPickerDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  identities: TenantIdentity[]
  platformAvailable: boolean
  platformRoles: string[]
  onSelectTenant: (tenantId: number) => void
  onSelectPlatform: () => void
  onCancel: () => void
}

export function IdentityPickerDialog({
  open,
  onOpenChange,
  identities,
  platformAvailable,
  platformRoles,
  onSelectTenant,
  onSelectPlatform,
  onCancel,
}: IdentityPickerDialogProps) {
  const handleOpenChange = (next: boolean) => {
    if (!next) onCancel()
    onOpenChange(next)
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>选择登录身份</DialogTitle>
          <DialogDescription>
            该账号关联了多个身份，请选择一个进入
          </DialogDescription>
        </DialogHeader>

        <div className="flex flex-col gap-2 max-h-[60vh] overflow-y-auto py-2">
          {identities.map((identity) => (
            <button
              key={`${identity.tenant_id}-${identity.user_id}`}
              type="button"
              onClick={() => onSelectTenant(identity.tenant_id)}
              className={cn(
                "flex items-start gap-3 rounded-md border border-input bg-background p-3 text-left transition-colors",
                "hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              )}
            >
              <BuildingIcon className="size-5 mt-0.5 shrink-0 text-muted-foreground" />
              <div className="flex-1 min-w-0">
                <div className="font-medium text-sm">
                  {identity.tenant_name}
                  <span className="ml-2 text-xs text-muted-foreground">
                    ({identity.tenant_code})
                  </span>
                </div>
                <div className="mt-1 flex items-center gap-3 text-xs text-muted-foreground">
                  <span className="flex items-center gap-1">
                    <UserIcon className="size-3" />
                    {identity.user_code}
                  </span>
                  <span>·</span>
                  <span className="rounded bg-muted px-1.5 py-0.5 text-foreground">
                    {identity.role}
                  </span>
                  {identity.real_name && (
                    <>
                      <span>·</span>
                      <span>{identity.real_name}</span>
                    </>
                  )}
                </div>
              </div>
            </button>
          ))}

          {platformAvailable && (
            <button
              type="button"
              onClick={onSelectPlatform}
              className={cn(
                "flex items-start gap-3 rounded-md border border-primary/30 bg-primary/5 p-3 text-left transition-colors",
                "hover:bg-primary/10 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              )}
            >
              <ShieldCheckIcon className="size-5 mt-0.5 shrink-0 text-primary" />
              <div className="flex-1 min-w-0">
                <div className="font-medium text-sm">进入平台后台</div>
                <div className="mt-1 flex items-center gap-2 text-xs text-muted-foreground">
                  <span>平台角色：</span>
                  {platformRoles.map((r) => (
                    <span
                      key={r}
                      className="rounded bg-primary/10 px-1.5 py-0.5 text-primary"
                    >
                      {r}
                    </span>
                  ))}
                </div>
              </div>
            </button>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={onCancel}>
            取消
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}