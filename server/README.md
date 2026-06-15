# XinFramework Server

> Go 1.25 + Gin + pgx + PostgreSQL。多租户 / RBAC / 插件化。

## 目录

- [架构总览](file:///d:\work\xin\XinFramework\server\doc\architecture.md) — Go module 切分、模块注册流程、Phase 1-5 重构
- [快速开始](file:///d:\work\xin\XinFramework\server\doc\quickstart.md) — 安装、配置、首次启动
- [模块清单](file:///d:\work\xin\XinFramework\server\doc\modules.md) — 内置模块 + apps 列表
- [API 参考](file:///d:\work\xin\XinFramework\server\doc\api.md) — HTTP 端点
- [数据库](file:///d:\work\xin\XinFramework\server\doc\database.md) — 表结构、迁移
- [权限](file:///d:\work\xin\XinFramework\server\doc\permissions.md) — RBAC、数据范围、平台角色
- [开发指南](file:///d:\work\xin\XinFramework\server\doc\developing.md) — 新增模块
- [部署](file:///d:\work\xin\XinFramework\server\doc\deployment.md) — 编译、systemd
- [AGENTS.md](file:///d:\work\xin\XinFramework\server\AGENTS.md) — 给 AI agent 的高密度参考

## 一句话概览

```
启动 → 配置加载 → DB 池化 → 注册模块（side-effect import）→ 拓扑排序 →
执行 Init → 跑迁移 → 注册全局中间件（recovery / CORS / logger）→ 暴露
/api/v1/{public,protected} 两个 gin.RouterGroup → 各模块 Register 路由 →
监听端口 + 优雅退出。
```

## 关键约定

1. **统一响应**：所有 handler 用 `resp.OK(c, data)` / `resp.Fail(c, code, msg)` 返回 `{code, msg, data}`
2. **认证中间件**：`middleware.OptionalAuth`（public 组）和 `middleware.Auth`（protected 组）按顺序挂载
3. **权限中间件**：`middleware.Require(spec)` / `RequireAll(specs)` / `RequireAny(specs)` 装饰具体路由
4. **平台角色中间件**：`middleware.RequirePlatformRole("super_admin", ...)` 装饰跨租户操作
5. **审计**：业务关键操作走 `audit.WithContext(c)` 在中间件里捕获
6. **错误**：业务错误用 `resp.ErrXxx`（如 `resp.ErrUserNotFound`），系统错误用 `fmt.Errorf` 包上下文

## 命令行

```
xin start          # 守护进程启动
xin stop           # 停止
xin restart        # 重启
xin reload         # 平滑重载
xin run            # 前台运行（开发用）
xin status         # 查看状态
```

构建：
```bash
go build -ldflags="-s -w" -o bin/xin ./cmd/xin
```

## 平台支持

- Linux（systemd）：[framework/xin-server.service](file:///d:\work\xin\XinFramework\server\framework\xin-server.service) + [build.sh](file:///d:\work\xin\XinFramework\server\build.sh)
- Windows：[build.ps1](file:///d:\work\xin\XinFramework\server\build.ps1)
- macOS / Linux：[build.sh](file:///d:\work\xin\XinFramework\server\build.sh)