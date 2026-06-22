import { Separator } from "@/components/ui/separator"
import { SidebarTrigger } from "@/components/ui/sidebar"
import { t } from "@/locales"
import { Switch } from "@/components/ui/switch"
import { useTheme } from "@/components/theme-provider"
import { useConfigItem } from "@/stores/configStore"
import { useAuthStore } from "@/stores/authStore"
import { Badge } from "@/components/ui/badge"
import { GlobeIcon, BuildingIcon, LogOutIcon } from "lucide-react"
import { Button } from "@/components/ui/button"
import { useNavigate } from "react-router-dom"
import { TenantSwitcher } from "@/components/tenant-switcher"

export function SiteHeader() {
  const { theme, setTheme } = useTheme()
  const siteName = useConfigItem("site", "site_name") as string | undefined
  const headerTitle = siteName || t.header.documents
  const scope = useAuthStore((s) => s.scope)
  const logout = useAuthStore((s) => s.logout)
  const navigate = useNavigate()

  const isDark =
    theme === "dark" ||
    (theme === "system" && window.matchMedia("(prefers-color-scheme: dark)").matches)

  const toggleTheme = () => {
    setTheme(isDark ? "light" : "dark")
  }

  const handleLogout = () => {
    logout()
    navigate(scope === "platform" ? "/platform/login" : "/login", { replace: true })
  }

  return (
    <header className="flex h-(--header-height) shrink-0 items-center gap-2 border-b transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-(--header-height)">
      <div className="flex w-full items-center gap-1 px-4 lg:gap-2 lg:px-6">
        <SidebarTrigger className="-ml-1" />
        <Separator
          orientation="vertical"
          className="mx-2 data-[orientation=vertical]:h-4"
        />
        <h1 className="text-base font-medium">{headerTitle}</h1>

        {scope && (
          <Badge
            variant={scope === "platform" ? "default" : "secondary"}
            className="ml-2 gap-1"
          >
            {scope === "platform" ? (
              <>
                <GlobeIcon className="size-3" /> 平台域
              </>
            ) : (
              <>
                <BuildingIcon className="size-3" /> 租户域
              </>
            )}
          </Badge>
        )}

        <div className="ml-auto flex items-center gap-4">
          <TenantSwitcher />
          <div className="flex items-center gap-2">
            <span className="text-sm text-muted-foreground">深色模式</span>
            <Switch checked={isDark} onCheckedChange={toggleTheme} />
          </div>
          <Button
            size="sm"
            variant="outline"
            onClick={handleLogout}
            className="gap-1"
          >
            <LogOutIcon className="size-3" />
            退出登录
          </Button>
        </div>
      </div>
    </header>
  )
}