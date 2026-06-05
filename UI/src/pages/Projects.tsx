import { PageLayout } from "@/components/page-layout"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { PlusIcon, SearchIcon, FolderIcon, MoreHorizontalIcon } from "lucide-react"
import { useTranslation } from "@/locales"

const mockProjects = [
  { id: 1, name: "企业管理系统", description: "公司内部管理系统", status: "active", tasks: 45, members: 8, progress: 78 },
  { id: 2, name: "客户关系系统", description: "CRM系统开发", status: "active", tasks: 32, members: 5, progress: 65 },
  { id: 3, name: "数据分析平台", description: "数据分析和可视化", status: "pending", tasks: 18, members: 3, progress: 30 },
  { id: 4, name: "移动端应用", description: "iOS和Android应用", status: "completed", tasks: 56, members: 6, progress: 100 },
]

export function ProjectsPage() {
  const t = useTranslation()

  return (
    <PageLayout>
      <div className="px-4 lg:px-6">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-2xl font-bold">{t.pages.projects?.title || "项目管理"}</h1>
            <p className="text-sm text-muted-foreground">{t.pages.projects?.subtitle || "管理系统中的所有项目"}</p>
          </div>
          <Button>
            <PlusIcon className="mr-2 h-4 w-4" />
            {t.common.add}
          </Button>
        </div>

        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4 mb-6">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">总项目数</CardTitle>
              <FolderIcon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">12</div>
              <p className="text-xs text-muted-foreground">+3 本季度</p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">进行中</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">8</div>
              <p className="text-xs text-muted-foreground">67% 项目活跃</p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">已完成</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">4</div>
              <p className="text-xs text-muted-foreground">全部按时完成</p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">团队成员</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">24</div>
              <p className="text-xs text-muted-foreground">参与各项目</p>
            </CardContent>
          </Card>
        </div>

        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <div className="relative flex-1 max-w-sm">
                <SearchIcon className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input placeholder="搜索项目..." className="pl-9" />
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>ID</TableHead>
                  <TableHead>{t.pages.projects?.name || "项目名称"}</TableHead>
                  <TableHead>{t.pages.projects?.description || "描述"}</TableHead>
                  <TableHead>{t.pages.projects?.status || "状态"}</TableHead>
                  <TableHead>{t.pages.projects?.tasks || "任务数"}</TableHead>
                  <TableHead>{t.pages.projects?.members || "成员"}</TableHead>
                  <TableHead>{t.pages.projects?.progress || "进度"}</TableHead>
                  <TableHead>{t.common.edit}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {mockProjects.map((project) => (
                  <TableRow key={project.id}>
                    <TableCell className="font-medium">{project.id}</TableCell>
                    <TableCell className="font-medium">{project.name}</TableCell>
                    <TableCell className="text-muted-foreground">{project.description}</TableCell>
                    <TableCell>
                      <Badge variant={project.status === "active" ? "default" : project.status === "completed" ? "secondary" : "outline"}>
                        {project.status === "active" ? (t.pages.projects?.active || "进行中") : 
                         project.status === "completed" ? (t.pages.projects?.completed || "已完成") : 
                         (t.pages.projects?.pending || "待开始")}
                      </Badge>
                    </TableCell>
                    <TableCell>{project.tasks}</TableCell>
                    <TableCell>{project.members}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <div className="w-16 h-2 bg-secondary rounded-full overflow-hidden">
                          <div className="h-full bg-primary" style={{ width: `${project.progress}%` }} />
                        </div>
                        <span className="text-sm">{project.progress}%</span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Button variant="ghost" size="icon">
                        <MoreHorizontalIcon className="h-4 w-4" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </div>
    </PageLayout>
  )
}