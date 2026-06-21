// Placeholder：平台域 Dict 管理页面

import { PageLayout } from "@/components/page-layout"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { GlobeIcon } from "lucide-react"

export function PlatformDictsPage() {
  return (
    <PageLayout>
      <div className="space-y-4 px-4 lg:px-6">
        <div className="flex items-center gap-2">
          <GlobeIcon className="size-5 text-primary" />
          <h1 className="text-2xl font-bold">平台字典</h1>
        </div>
        <Card>
          <CardHeader>
            <CardTitle>平台 Dict 管理</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              平台级数据字典（scope=platform / visibility=public）的 CRUD 入口。
              Phase 0022 后续接入，当前与 /admin/platform-dicts/* 后端端点已就绪。
            </p>
          </CardContent>
        </Card>
      </div>
    </PageLayout>
  )
}