import { PageLayout } from "@/components/page-layout"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Building2Icon, MenuIcon, ShieldIcon, ServerIcon } from "lucide-react"
import { useNavigate } from "react-router-dom"
import { useAuthStore } from "@/stores/authStore"

export function SysDashboardPage() {
  const navigate = useNavigate()
  const user = useAuthStore((s) => s.user)

  // 默认登录后看到简化的 sys 统计
  const cards = [
    {
      title: "Sys 租户",
      desc: "管理所有租户（创建 / 启用 / 硬删）",
      icon: Building2Icon,
      to: "/sys/tenants",
    },
    {
      title: "Sys 菜单",
      desc: "管理 tenant_id=0 的全局菜单（所有租户共享）",
      icon: MenuIcon,
      to: "/sys/menus",
    },
    {
      title: "Sys 配置",
      desc: "管理全局 sys 级配置项（Scope=sys）",
      icon: ServerIcon,
      to: "/sys/configs",
    },
    {
      title: "Sys 字典",
      desc: "管理 sys 级数据字典（Visibility=public/all）",
      icon: ShieldIcon,
      to: "/sys/dicts",
    },
  ]

  return (
    <PageLayout>
      <div className="space-y-6 px-4 lg:px-6">
        <div>
          <h1 className="text-2xl font-bold">Sys 后台</h1>
          <p className="text-sm text-muted-foreground">
            欢迎，{user?.real_name || user?.code || "Sys 管理员"}。当前会话作用域：
            <span className="ml-1 rounded bg-primary/10 px-2 py-0.5 text-primary">sys</span>
          </p>
        </div>
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
          {cards.map((c) => (
            <Card
              key={c.title}
              className="cursor-pointer hover:border-primary/40 hover:shadow-md transition"
              onClick={() => navigate(c.to)}
            >
              <CardHeader className="pb-2">
                <CardTitle className="flex items-center gap-2 text-sm font-medium">
                  <c.icon className="size-4 text-primary" />
                  {c.title}
                </CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-xs text-muted-foreground">{c.desc}</p>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    </PageLayout>
  )
}
