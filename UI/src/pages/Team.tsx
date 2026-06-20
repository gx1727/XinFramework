import { PageLayout } from "@/components/page-layout"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { SearchIcon, PlusIcon, MailIcon, MessageSquareIcon, UsersIcon } from "lucide-react"
import { t } from "@/locales"

const mockTeamMembers = [
  { id: 1, name: "张三", avatar: "", role: "超级管理员", email: "zhangsan@example.com", status: "online", projects: 5 },
  { id: 2, name: "李四", avatar: "", role: "开发人员", email: "lisi@example.com", status: "online", projects: 3 },
  { id: 3, name: "王五", avatar: "", role: "产品经理", email: "wangwu@example.com", status: "offline", projects: 4 },
  { id: 4, name: "赵六", avatar: "", role: "设计师", email: "zhaoliu@example.com", status: "online", projects: 2 },
  { id: 5, name: "钱七", avatar: "", role: "测试工程师", email: "qianqi@example.com", status: "away", projects: 3 },
  { id: 6, name: "孙八", avatar: "", role: "运维工程师", email: "sunba@example.com", status: "online", projects: 4 },
]

const mockTeams = [
  { id: 1, name: "前端开发组", members: 5, projects: 8 },
  { id: 2, name: "后端开发组", members: 6, projects: 10 },
  { id: 3, name: "产品设计组", members: 4, projects: 5 },
  { id: 4, name: "测试组", members: 3, projects: 7 },
]

export function TeamPage() {
  return (
    <PageLayout>
      <div className="px-4 lg:px-6">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-2xl font-bold">{t.pages.team?.title || "团队管理"}</h1>
            <p className="text-sm text-muted-foreground">{t.pages.team?.subtitle || "管理团队成员和组织结构"}</p>
          </div>
          <Button>
            <PlusIcon className="mr-2 h-4 w-4" />
            {t.pages.team?.addMember || "添加成员"}
          </Button>
        </div>

        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4 mb-6">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">总成员数</CardTitle>
              <UsersIcon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">24</div>
              <p className="text-xs text-muted-foreground">+2 本月</p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">在线成员</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">18</div>
              <p className="text-xs text-muted-foreground">75% 在线</p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">团队数</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">4</div>
              <p className="text-xs text-muted-foreground">跨部门协作</p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">进行中项目</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">12</div>
              <p className="text-xs text-muted-foreground">本月新增 3 个</p>
            </CardContent>
          </Card>
        </div>

        <div className="grid gap-6 lg:grid-cols-3">
          <div className="lg:col-span-2">
            <Card>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <CardTitle>{t.pages.team?.members || "团队成员"}</CardTitle>
                  <div className="relative w-64">
                    <SearchIcon className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                    <Input placeholder="搜索成员..." className="pl-9" />
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                <div className="grid gap-4 md:grid-cols-2">
                  {mockTeamMembers.map((member) => (
                    <div key={member.id} className="flex items-center gap-4 p-4 border rounded-lg">
                      <Avatar className="h-12 w-12">
                        <AvatarImage src={member.avatar} />
                        <AvatarFallback>{member.name.slice(0, 2)}</AvatarFallback>
                      </Avatar>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <span className="font-medium truncate">{member.name}</span>
                          <span className={`w-2 h-2 rounded-full ${
                            member.status === "online" ? "bg-green-500" : 
                            member.status === "away" ? "bg-yellow-500" : "bg-gray-400"
                          }`} />
                        </div>
                        <p className="text-sm text-muted-foreground truncate">{member.email}</p>
                        <div className="flex items-center gap-2 mt-1">
                          <Badge variant="secondary" className="text-xs">{member.role}</Badge>
                          <span className="text-xs text-muted-foreground">{member.projects} {t.pages.projects?.projects || "项目"}</span>
                        </div>
                      </div>
                      <div className="flex gap-1">
                        <Button variant="ghost" size="icon">
                          <MailIcon className="h-4 w-4" />
                        </Button>
                        <Button variant="ghost" size="icon">
                          <MessageSquareIcon className="h-4 w-4" />
                        </Button>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </div>

          <div>
            <Card>
              <CardHeader>
                <CardTitle>{t.pages.team?.teams || "团队列表"}</CardTitle>
                <CardDescription>{t.pages.team?.teamsDesc || "按团队分组查看成员"}</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                {mockTeams.map((team) => (
                  <div key={team.id} className="flex items-center justify-between p-3 border rounded-lg hover:bg-accent cursor-pointer">
                    <div>
                      <p className="font-medium">{team.name}</p>
                      <p className="text-sm text-muted-foreground">{team.members} {t.pages.team?.members || "成员"}</p>
                    </div>
                    <Badge variant="secondary">{team.projects} {t.pages.projects?.projects || "项目"}</Badge>
                  </div>
                ))}
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </PageLayout>
  )
}