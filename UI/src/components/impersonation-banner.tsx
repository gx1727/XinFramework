// ImpersonationBanner — super_admin 模拟登录租户时的顶部醒目横幅。
//
// 显示在 PageLayout 内容顶部（page header 之上），包含：
//   - 醒目红色背景 + AlertTriangleIcon 图标
//   - 当前模拟的租户名称 + 审计提示
//   - "退出模拟" 按钮，调用 authStore.stopImpersonation()，成功后跳回 /platform/tenants
//
// 当 authStore.impersonation 为 null 时不渲染。
import { useNavigate } from "react-router-dom"
import { AlertTriangleIcon, LogOutIcon } from "lucide-react"
import { Button } from "@/components/ui/button"
import { useAuthStore } from "@/stores/authStore"
import { t } from "@/locales"

export function ImpersonationBanner() {
  const navigate = useNavigate()
  const impersonation = useAuthStore((s) => s.impersonation)
  const stopImpersonation = useAuthStore((s) => s.stopImpersonation)

  if (!impersonation) return null

  const handleExit = async () => {
    const ok = await stopImpersonation()
    if (ok) {
      navigate("/platform/tenants", { replace: true })
    }
  }

  return (
    <div
      role="alert"
      className="flex items-center justify-between gap-3 border-b-2 border-amber-300 bg-amber-50 px-4 py-2 text-amber-900"
    >
      <div className="flex items-center gap-3 min-w-0">
        <AlertTriangleIcon className="size-5 shrink-0" />
        <div className="min-w-0">
          <div className="text-sm font-semibold">
            {t.impersonation.banner.replace("{name}", impersonation.tenantName)}
          </div>
          <div className="text-xs text-amber-800/80">
            {t.impersonation.bannerHint}
          </div>
        </div>
      </div>
      <Button
        variant="outline"
        size="sm"
        onClick={handleExit}
        className="border-amber-400 bg-white hover:bg-amber-100"
      >
        <LogOutIcon className="size-4 mr-1" />
        {t.impersonation.exit}
      </Button>
    </div>
  )
}