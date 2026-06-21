// Placeholder：平台域 Config 管理页面
//
// 实际实现可参考 pages/Configs.tsx（业务域）+ 切换到 platformMenuApi。
// 当前作为入口占位，避免 tsc 报错；后续按 Phase 0022 后续 phase 接入完整 CRUD。

import { PageLayout } from "@/components/page-layout"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { GlobeIcon } from "lucide-react"

export function PlatformConfigsPage() {
  return (
    <PageLayout>
      <div className="space-y-4 px-4 lg:px-6">
        <div className="flex items-center gap-2">
          <GlobeIcon className="size-5 text-primary" />
          <h1 className="text-2xl font-bold">平台配置</h1>
        </div>
        <Card>
          <CardHeader>
            <CardTitle>平台 Config 管理</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              平台级配置项（scope=platform）的 CRUD 入口。Phase 0022 后续接入，
              当前与 /admin/platform-configs/* 后端端点已就绪。
            </p>
          </CardContent>
        </Card>
      </div>
    </PageLayout>
  )
}