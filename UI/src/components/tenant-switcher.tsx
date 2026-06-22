// TenantSwitcher 顶部租户切换下拉（路径 B 多身份支持）。
//
// 仅当满足以下条件才渲染：
//   1. scope === "tenant"
//   2. availableIdentities.length >= 2（有多个租户身份）
//
// 行为：选目标 tenant → 调 switchTenant → 刷新页面（侧边栏 / 权限缓存）。

import { ChevronsUpDownIcon, CheckIcon } from "lucide-react"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Button } from "@/components/ui/button"
import { useAuthStore } from "@/stores/authStore"

export function TenantSwitcher() {
  const scope = useAuthStore((s) => s.scope)
  const user = useAuthStore((s) => s.user)
  const availableIdentities = useAuthStore((s) => s.availableIdentities)
  const switchTenant = useAuthStore((s) => s.switchTenant)

  // 平台域 / 单身份账号 / 未登录 → 不显示
  if (scope !== "tenant" || !user) return null
  if (availableIdentities.length < 2) return null

  const current = availableIdentities.find((i) => i.tenant_id === user.tenant_id)

  const handleSwitch = async (tenantId: number) => {
    if (tenantId === user.tenant_id) return
    const success = await switchTenant(tenantId)
    if (success) {
      // 切换后刷新页面：侧边栏菜单 / 权限缓存 / dataScope 都跟 tenant 绑，
      // 简单 reload 让所有 zustand store 重新加载
      window.location.reload()
    }
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="outline"
          className="h-9 gap-2 px-3 font-normal"
          aria-label="切换租户"
        >
          <span className="truncate max-w-[180px]">
            {current?.tenant_name ?? `租户 #${user.tenant_id}`}
          </span>
          <ChevronsUpDownIcon className="size-3.5 shrink-0 text-muted-foreground" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-64">
        <DropdownMenuLabel className="text-xs text-muted-foreground">
          切换租户身份
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        {availableIdentities.map((identity) => {
          const isCurrent = identity.tenant_id === user.tenant_id
          return (
            <DropdownMenuItem
              key={`${identity.tenant_id}-${identity.user_id}`}
              onSelect={() => handleSwitch(identity.tenant_id)}
              className="flex items-start gap-2 py-2"
            >
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 text-sm font-medium">
                  {identity.tenant_name}
                  {isCurrent && (
                    <CheckIcon className="size-3.5 shrink-0 text-primary" />
                  )}
                </div>
                <div className="mt-0.5 flex items-center gap-2 text-xs text-muted-foreground">
                  <span>{identity.tenant_code}</span>
                  <span>·</span>
                  <span>{identity.user_code}</span>
                  <span>·</span>
                  <span className="rounded bg-muted px-1 py-0.5 text-foreground">
                    {identity.role}
                  </span>
                </div>
              </div>
            </DropdownMenuItem>
          )
        })}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}