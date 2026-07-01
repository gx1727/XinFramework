import { useEffect, useState } from "react"
import { PageLayout } from "@/components/page-layout"
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog"
import { systemApi, type CacheInfo, type CacheValue } from "@/api"
import { DataTablePagination } from "@/components/data-table-pagination"
import {
  SearchIcon,
  TrashIcon,
  RefreshCw,
  EyeIcon,
  DatabaseIcon,
  ServerIcon,
  ActivityIcon,
  ClockIcon,
  UsersIcon,
} from "lucide-react"

export default function Cache() {
  const [cacheInfo, setCacheInfo] = useState<CacheInfo | null>(null)
  const [isLoadingInfo, setIsLoadingInfo] = useState(false)

  const [pattern, setPattern] = useState("*")
  const [keys, setKeys] = useState<string[]>([])
  const [isLoadingKeys, setIsLoadingKeys] = useState(false)
  const [page, setPage] = useState(1)
  const [size, setSize] = useState(50)
  const [total, setTotal] = useState(0)
  // 手动增加以 force re-fetch（setPage 在同值时不会触发 effect）
  const [fetchTrigger, setFetchTrigger] = useState(0)
  // 默认为 false：后端 SCAN 后过滤掉 cache_* 和 sess:* 开头的系统缓存键，减少噪音
  const [showSystemKeys, setShowSystemKeys] = useState(false)

  // 系统缓存键前缀（后端 SCAN 后过滤）
  const SYSTEM_KEY_PREFIXES = ["cache_", "sess:"] as const

  const [selectedKey, setSelectedKey] = useState<string | null>(null)
  const [keyValue, setKeyValue] = useState<CacheValue | null>(null)
  const [isLoadingValue, setIsLoadingValue] = useState(false)
  const [detailOpen, setDetailOpen] = useState(false)

  const fetchCacheInfo = async () => {
    setIsLoadingInfo(true)
    try {
      const res = await systemApi.getCacheInfo()
      setCacheInfo(res)
    } catch (error: any) {
      console.error("获取缓存信息失败", error)
    } finally {
      setIsLoadingInfo(false)
    }
  }

  const fetchKeys = async () => {
    setIsLoadingKeys(true)
    try {
      const res = await systemApi.getCacheKeys(
        pattern || "*",
        page,
        size,
        showSystemKeys ? undefined : [...SYSTEM_KEY_PREFIXES]
      )
      setKeys(res?.list ?? [])
      setTotal(res?.total ?? 0)
    } catch (error: any) {
      console.error("获取缓存键失败", error)
      setKeys([])
      setTotal(0)
    } finally {
      setIsLoadingKeys(false)
    }
  }

  useEffect(() => {
    fetchCacheInfo()
    fetchKeys()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, size, pattern, showSystemKeys, fetchTrigger])

  const handleSearch = () => {
    // 搜索时重置到第 1 页 + bump fetchTrigger（同 page 不触发 effect 的兑底）
    setPage(1)
    setFetchTrigger((k) => k + 1)
  }

  const handleView = async (key: string) => {
    setSelectedKey(key)
    setDetailOpen(true)
    setIsLoadingValue(true)
    setKeyValue(null)
    try {
      const res = await systemApi.getCacheValue(key)
      setKeyValue(res)
    } catch (error: any) {
      console.error("获取缓存值失败", error)
      setDetailOpen(false)
    } finally {
      setIsLoadingValue(false)
    }
  }

  const handleDelete = async (key: string) => {
    if (!window.confirm(`确定要删除缓存键 ${key} 吗？`)) return

    try {
      await systemApi.deleteCacheKey(key)
      alert(`已删除缓存键: ${key}`)
      if (selectedKey === key) {
        setDetailOpen(false)
      }
      // 删除后重置到第 1 页，避免当前页空了
      setPage(1)
      setFetchTrigger((k) => k + 1)
      fetchCacheInfo()
    } catch (error: any) {
      console.error("删除失败", error)
      alert(`删除失败: ${error.message}`)
    }
  }

  // Parses info object for quick stats if it exists
  const getParsedInfo = () => {
    if (!cacheInfo?.info) return {}
    if (typeof cacheInfo.info === "object") {
      return cacheInfo.info
    }
    return {}
  }

  const parsedInfo = getParsedInfo() as any

  return (
    <PageLayout>
      <div className="px-4 lg:px-6">
        <div className="mb-6 flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">缓存管理</h1>
            <p className="text-sm text-muted-foreground">
              管理系统 Redis 缓存数据
            </p>
          </div>
          <Button
            variant="outline"
            onClick={() => {
              fetchCacheInfo()
              fetchKeys()
            }}
            disabled={isLoadingInfo || isLoadingKeys}
          >
            <RefreshCw
              className={`mr-2 h-4 w-4 ${isLoadingInfo || isLoadingKeys ? "animate-spin" : ""}`}
            />
            刷新
          </Button>
        </div>

        <div className="mb-6 grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">Redis 版本</CardTitle>
              <ServerIcon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {parsedInfo.redis_version || "-"}
              </div>
              <p className="text-xs text-muted-foreground">
                运行天数: {parsedInfo.uptime_in_days || "-"} 天
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">
                客户端连接数
              </CardTitle>
              <UsersIcon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {parsedInfo.connected_clients || "-"}
              </div>
              <p className="text-xs text-muted-foreground">
                当前连接的客户端数量
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">内存使用</CardTitle>
              <ActivityIcon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {parsedInfo.used_memory_human || "-"}
              </div>
              <p className="text-xs text-muted-foreground">
                峰值: {parsedInfo.used_memory_peak_human || "-"}
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">Key 总数</CardTitle>
              <DatabaseIcon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{cacheInfo?.dbSize || 0}</div>
              <p className="text-xs text-muted-foreground">当前数据库键数量</p>
            </CardContent>
          </Card>
        </div>

        <div className="grid gap-6 md:grid-cols-12">
          <Card className="md:col-span-12">
            <CardHeader>
              <CardTitle>缓存键列表</CardTitle>
              <CardDescription>
                使用通配符搜索缓存键，例如：user:*
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="mb-4 flex flex-wrap items-center gap-2">
                <div className="relative max-w-sm flex-1">
                  <SearchIcon className="absolute top-2.5 left-2.5 h-4 w-4 text-muted-foreground" />
                  <Input
                    type="text"
                    placeholder="匹配模式 (例如: *)"
                    className="pl-8"
                    value={pattern}
                    onChange={(e) => setPattern(e.target.value)}
                    onKeyDown={(e) => e.key === "Enter" && handleSearch()}
                  />
                </div>
                <Button onClick={handleSearch} disabled={isLoadingKeys}>
                  搜索
                </Button>
                <label
                  htmlFor="show-system-keys"
                  className="ml-2 flex cursor-pointer items-center gap-2 text-sm text-muted-foreground select-none"
                >
                  <Checkbox
                    id="show-system-keys"
                    checked={showSystemKeys}
                    onCheckedChange={(c) => setShowSystemKeys(c === true)}
                  />
                  显示系统缓存键
                </label>
              </div>

              <div className="rounded-md border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-16">序号</TableHead>
                      <TableHead>缓存键名</TableHead>
                      <TableHead className="text-right">操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {keys.map((key, index) => (
                      <TableRow key={key}>
                        <TableCell className="text-muted-foreground">
                          {(page - 1) * size + index + 1}
                        </TableCell>
                        <TableCell className="font-mono text-sm">
                          {key}
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex items-center justify-end gap-1">
                            <Button
                              variant="ghost"
                              size="icon"
                              className="h-8 w-8"
                              onClick={() => handleView(key)}
                              title="查看"
                            >
                              <EyeIcon className="h-4 w-4" />
                            </Button>
                            <Button
                              variant="ghost"
                              size="icon"
                              className="h-8 w-8"
                              onClick={() => handleDelete(key)}
                              title="删除"
                            >
                              <TrashIcon className="h-4 w-4 text-destructive" />
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                    {keys.length === 0 && !isLoadingKeys && (
                      <TableRow>
                        <TableCell
                          colSpan={3}
                          className="py-8 text-center text-muted-foreground"
                        >
                          未找到匹配的缓存键
                        </TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
                {isLoadingKeys && (
                  <div className="flex items-center justify-center py-8">
                    <div className="text-sm text-muted-foreground">
                      加载中...
                    </div>
                  </div>
                )}
              </div>
              <DataTablePagination
                page={page}
                size={size}
                total={total}
                isLoading={isLoadingKeys}
                onPageChange={setPage}
                onSizeChange={setSize}
              />
            </CardContent>
          </Card>
        </div>
      </div>

      <Dialog open={detailOpen} onOpenChange={setDetailOpen}>
        <DialogContent className="sm:max-w-[600px]">
          <DialogHeader>
            <DialogTitle className="pr-6 break-all">缓存详情</DialogTitle>
          </DialogHeader>
          {isLoadingValue ? (
            <div className="py-8 text-center text-muted-foreground">
              加载中...
            </div>
          ) : keyValue ? (
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-1">
                  <p className="text-sm font-medium text-muted-foreground">
                    键名
                  </p>
                  <p className="font-mono text-sm break-all">{keyValue.key}</p>
                </div>
                <div className="space-y-1">
                  <p className="text-sm font-medium text-muted-foreground">
                    数据类型
                  </p>
                  <div className="flex items-center">
                    <DatabaseIcon className="mr-1.5 h-3.5 w-3.5 text-primary" />
                    <span className="font-mono text-sm uppercase">
                      {keyValue.type}
                    </span>
                  </div>
                </div>
                <div className="col-span-2 space-y-1">
                  <p className="text-sm font-medium text-muted-foreground">
                    过期时间 (TTL)
                  </p>
                  <div className="flex items-center">
                    <ClockIcon className="mr-1.5 h-3.5 w-3.5 text-orange-500" />
                    <span className="font-mono text-sm">
                      {keyValue.ttl === -1
                        ? "永久有效 (-1)"
                        : keyValue.ttl === -2
                          ? "已过期 (-2)"
                          : `${keyValue.ttl} 秒`}
                    </span>
                  </div>
                </div>
              </div>

              <div className="space-y-2">
                <p className="text-sm font-medium text-muted-foreground">
                  缓存内容
                </p>
                <div className="h-[250px] w-full overflow-auto rounded-md border bg-muted/30 p-4">
                  <pre className="font-mono text-xs break-all whitespace-pre-wrap">
                    {typeof keyValue.value === "object"
                      ? JSON.stringify(keyValue.value, null, 2)
                      : String(keyValue.value)}
                  </pre>
                </div>
              </div>
            </div>
          ) : (
            <div className="py-8 text-center text-muted-foreground">
              无法获取缓存详情
            </div>
          )}
          <DialogFooter className="gap-2 sm:gap-0">
            <Button
              variant="destructive"
              onClick={() => {
                if (selectedKey) handleDelete(selectedKey)
              }}
            >
              删除此键
            </Button>
            <Button variant="outline" onClick={() => setDetailOpen(false)}>
              关闭
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </PageLayout>
  )
}
