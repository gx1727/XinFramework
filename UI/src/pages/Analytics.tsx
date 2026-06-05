import { PageLayout } from "@/components/page-layout"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { ChartAreaInteractive } from "@/components/chart-area-interactive"
import { useTranslation } from "@/locales"

export function AnalyticsPage() {
  const t = useTranslation()

  return (
    <PageLayout>
      <div className="px-4 lg:px-6">
        <div className="mb-6">
          <h1 className="text-2xl font-bold">{t.pages.analytics?.title || "数据分析"}</h1>
          <p className="text-sm text-muted-foreground">{t.pages.analytics?.subtitle || "查看系统运行数据和趋势分析"}</p>
        </div>

        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4 mb-6">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">总访问量</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">128,459</div>
              <p className="text-xs text-muted-foreground">+12.5% 较上月</p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">活跃用户</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">3,847</div>
              <p className="text-xs text-muted-foreground">+8.2% 较上月</p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">转化率</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">24.5%</div>
              <p className="text-xs text-muted-foreground">+3.1% 较上月</p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">平均响应时间</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">245ms</div>
              <p className="text-xs text-muted-foreground">-15% 较上月</p>
            </CardContent>
          </Card>
        </div>

        <Card className="mb-6">
          <CardHeader>
            <CardTitle>{t.pages.analytics?.trendChart || "趋势图表"}</CardTitle>
            <CardDescription>{t.pages.analytics?.trendDesc || "展示近30天的数据变化"}</CardDescription>
          </CardHeader>
          <CardContent>
            <ChartAreaInteractive />
          </CardContent>
        </Card>

        <div className="grid gap-4 md:grid-cols-2">
          <Card>
            <CardHeader>
              <CardTitle>{t.pages.analytics?.topPages || "热门页面"}</CardTitle>
              <CardDescription>{t.pages.analytics?.topPagesDesc || "访问量最高的页面排名"}</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <span className="text-sm">/dashboard</span>
                  <span className="font-mono text-sm">28,459</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm">/users</span>
                  <span className="font-mono text-sm">15,234</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm">/settings</span>
                  <span className="font-mono text-sm">12,876</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm">/analytics</span>
                  <span className="font-mono text-sm">9,543</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm">/projects</span>
                  <span className="font-mono text-sm">8,234</span>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>{t.pages.analytics?.userSources || "用户来源"}</CardTitle>
              <CardDescription>{t.pages.analytics?.userSourcesDesc || "用户访问来源分布"}</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <span className="text-sm">直接访问</span>
                  <span className="font-mono text-sm">45.2%</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm">搜索引擎</span>
                  <span className="font-mono text-sm">32.8%</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm">社交媒体</span>
                  <span className="font-mono text-sm">12.5%</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm">外部链接</span>
                  <span className="font-mono text-sm">9.5%</span>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </PageLayout>
  )
}