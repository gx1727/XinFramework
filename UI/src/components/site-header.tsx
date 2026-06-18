import { Separator } from "@/components/ui/separator"
import { SidebarTrigger } from "@/components/ui/sidebar"
import { useTranslation } from "@/locales"
import { LanguageSwitcher } from "@/components/language-switcher"
import { Switch } from "@/components/ui/switch"
import { useTheme } from "@/components/theme-provider"
import { useConfigItem } from "@/stores/configStore"

export function SiteHeader() {
  const t = useTranslation()
  const { theme, setTheme } = useTheme()
  const siteName = useConfigItem("site", "site_name") as string | undefined
  const headerTitle = siteName || t.header.documents

  const isDark = theme === "dark" || (theme === "system" && window.matchMedia("(prefers-color-scheme: dark)").matches)

  const toggleTheme = () => {
    setTheme(isDark ? "light" : "dark")
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
        <div className="ml-auto flex items-center gap-4">
          <div className="flex items-center gap-2">
            <span className="text-sm text-muted-foreground">深色模式</span>
            <Switch
              checked={isDark}
              onCheckedChange={toggleTheme}
            />
          </div>
          <LanguageSwitcher />
        </div>
      </div>
    </header>
  )
}