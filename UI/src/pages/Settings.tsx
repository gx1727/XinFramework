import { PageLayout } from "@/components/page-layout"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { useTranslation } from "@/locales"

export function SettingsPage() {
  const t = useTranslation()

  return (
    <PageLayout>
      <div className="px-4 lg:px-6">
        <div className="mb-6">
          <h1 className="text-2xl font-bold">{t.pages.settings?.title || "系统设置"}</h1>
          <p className="text-sm text-muted-foreground">{t.pages.settings?.subtitle || "管理系统配置和偏好设置"}</p>
        </div>

        <Tabs defaultValue="general" className="space-y-6">
          <TabsList>
            <TabsTrigger value="general">{t.pages.settings?.general || "通用设置"}</TabsTrigger>
            <TabsTrigger value="security">{t.pages.settings?.security || "安全设置"}</TabsTrigger>
            <TabsTrigger value="notifications">{t.pages.settings?.notifications || "通知设置"}</TabsTrigger>
            <TabsTrigger value="appearance">{t.pages.settings?.appearance || "外观"}</TabsTrigger>
          </TabsList>

          <TabsContent value="general">
            <Card>
              <CardHeader>
                <CardTitle>{t.pages.settings?.generalSettings || "通用设置"}</CardTitle>
                <CardDescription>{t.pages.settings?.generalDesc || "配置系统通用选项"}</CardDescription>
              </CardHeader>
              <CardContent className="space-y-6">
                <div className="space-y-2">
                  <Label htmlFor="siteName">{t.pages.settings?.siteName || "网站名称"}</Label>
                  <Input id="siteName" defaultValue="XinFramework" />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="siteUrl">{t.pages.settings?.siteUrl || "网站地址"}</Label>
                  <Input id="siteUrl" defaultValue="https://example.com" />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="timezone">{t.pages.settings?.timezone || "时区"}</Label>
                  <Select defaultValue="asia-shanghai">
                    <SelectTrigger id="timezone">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="asia-shanghai">Asia/Shanghai (UTC+8)</SelectItem>
                      <SelectItem value="america-new-york">America/New_York (UTC-5)</SelectItem>
                      <SelectItem value="europe-london">Europe/London (UTC+0)</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="language">{t.pages.settings?.language || "语言"}</Label>
                  <Select defaultValue="zh-CN">
                    <SelectTrigger id="language">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="zh-CN">简体中文</SelectItem>
                      <SelectItem value="en-US">English</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <Button>{t.common.save}</Button>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="security">
            <Card>
              <CardHeader>
                <CardTitle>{t.pages.settings?.securitySettings || "安全设置"}</CardTitle>
                <CardDescription>{t.pages.settings?.securityDesc || "管理账户安全选项"}</CardDescription>
              </CardHeader>
              <CardContent className="space-y-6">
                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <Label>{t.pages.settings?.twoFactor || "双因素认证"}</Label>
                    <p className="text-sm text-muted-foreground">{t.pages.settings?.twoFactorDesc || "为账户添加额外的安全验证"}</p>
                  </div>
                  <Button variant="outline">{t.pages.settings?.enable || "启用"}</Button>
                </div>
                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <Label>{t.pages.settings?.changePassword || "修改密码"}</Label>
                    <p className="text-sm text-muted-foreground">{t.pages.settings?.changePasswordDesc || "定期更换密码可以保护账户安全"}</p>
                  </div>
                  <Button variant="outline">{t.pages.settings?.update || "更新"}</Button>
                </div>
                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <Label>{t.pages.settings?.sessionTimeout || "会话超时"}</Label>
                    <p className="text-sm text-muted-foreground">{t.pages.settings?.sessionTimeoutDesc || "设置空闲自动退出时间"}</p>
                  </div>
                  <Select defaultValue="30">
                    <SelectTrigger className="w-[120px]">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="15">15 分钟</SelectItem>
                      <SelectItem value="30">30 分钟</SelectItem>
                      <SelectItem value="60">60 分钟</SelectItem>
                      <SelectItem value="120">2 小时</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="notifications">
            <Card>
              <CardHeader>
                <CardTitle>{t.pages.settings?.notificationSettings || "通知设置"}</CardTitle>
                <CardDescription>{t.pages.settings?.notificationDesc || "配置系统通知和提醒方式"}</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <Label>{t.pages.settings?.emailNotifications || "邮件通知"}</Label>
                    <p className="text-sm text-muted-foreground">{t.pages.settings?.emailDesc || "接收重要事件的邮件通知"}</p>
                  </div>
                  <Badge>已开启</Badge>
                </div>
                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <Label>{t.pages.settings?.pushNotifications || "推送通知"}</Label>
                    <p className="text-sm text-muted-foreground">{t.pages.settings?.pushDesc || "接收系统推送消息"}</p>
                  </div>
                  <Badge variant="secondary">已关闭</Badge>
                </div>
                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <Label>{t.pages.settings?.weeklyReport || "周报"}</Label>
                    <p className="text-sm text-muted-foreground">{t.pages.settings?.weeklyDesc || "每周发送系统使用报告"}</p>
                  </div>
                  <Badge>已开启</Badge>
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="appearance">
            <Card>
              <CardHeader>
                <CardTitle>{t.pages.settings?.appearanceSettings || "外观设置"}</CardTitle>
                <CardDescription>{t.pages.settings?.appearanceDesc || "自定义界面外观"}</CardDescription>
              </CardHeader>
              <CardContent className="space-y-6">
                <div className="space-y-2">
                  <Label>{t.pages.settings?.theme || "主题"}</Label>
                  <div className="flex gap-4">
                    <Button variant="outline" className="flex-1">浅色</Button>
                    <Button variant="secondary" className="flex-1">深色</Button>
                    <Button variant="secondary" className="flex-1">跟随系统</Button>
                  </div>
                </div>
                <div className="space-y-2">
                  <Label>{t.pages.settings?.compactMode || "紧凑模式"}</Label>
                  <p className="text-sm text-muted-foreground">{t.pages.settings?.compactModeDesc || "减少界面元素间距以显示更多内容"}</p>
                  <Button variant="outline">开启</Button>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </div>
    </PageLayout>
  )
}