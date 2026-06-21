import { ShieldIcon, Link2Icon } from "lucide-react"
import { LoginForm } from "@/components/login-form"
import { useConfigItem } from "@/stores/configStore"
import { Link } from "react-router-dom"

export function PlatformLoginPage() {
  const siteName = useConfigItem("site", "site_name") as string | undefined
  const siteLogo = useConfigItem("site", "site_logo") as string | undefined

  const title = siteName || "XinFramework"

  if (typeof document !== "undefined" && document.title !== `${title} - 平台后台`) {
    document.title = `${title} - 平台后台`
  }

  return (
    <div className="grid min-h-svh place-items-center bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 p-6">
      <div className="w-full max-w-md">
        <div className="mb-6 flex items-center justify-center gap-2 text-slate-100">
          {siteLogo ? (
            <img src={siteLogo} alt={title} className="size-6 object-contain" />
          ) : (
            <div className="flex size-6 items-center justify-center rounded-md bg-primary text-primary-foreground">
              <ShieldIcon className="size-4" />
            </div>
          )}
          <span className="text-base font-semibold">{title} · 平台后台</span>
        </div>
        <div className="rounded-lg border border-slate-700 bg-slate-900/70 p-6 shadow-2xl backdrop-blur">
          <LoginForm mode="platform" />
        </div>
        <div className="mt-4 text-center text-xs text-slate-400">
          <Link
            to="/login"
            className="inline-flex items-center gap-1 hover:text-slate-100"
          >
            <Link2Icon className="size-3" /> 业务用户登录入口
          </Link>
        </div>
      </div>
    </div>
  )
}