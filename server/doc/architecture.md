# 架构总览

> 本文档解释 XinFramework 后端的模块划分、运行时序、以及 [Phase 1-5 重构](file:///d:\work\xin\XinFramework\server\doc\architecture.md#重构方案phase-1-5) 的来龙去脉。

## 一句话

```
framework/ = 基础设施（pkg）+ framework 内部业务模块（Phase 3 即将清空）
apps/      = 业务模块（auth / tenant / cms / flag / 未来的 RBAC 与参考实现）
```

任何业务模块的"是否启用"由 `config.yaml` 的 `module:` 列表控制；任何业务模块的"是否打包"由 `cmd/xin/main.go` 的 side-effect import 控制。

---

## 1. 目录与 module 切分

```
server/
├── go.mod                      # module gx1727.com/xin
├── go.work                     # use(., ./framework, ./apps)
│
├── framework/                  # 独立 Go module: gx1727.com/xin/framework
│   ├── pkg/                    # 公开 SDK（任意位置可 import）
│   │   ├── plugin/             #   模块注册中心
│   │   ├── auth/               #   Account / AccountAuth / AccountRepository
│   │   ├── tenant/             #   TenantRepository + 注册钩子
│   │   ├── middleware/         #   Require / RequirePlatformRole
│   │   ├── permission/         #   P() Spec / 常量 / DataScope
│   │   ├── resp/               #   统一响应 / 错误码
│   │   ├── session/ jwt/ db/   #   会话 / JWT / pgx
│   │   ├── cache/ migrate/     #   LRU 缓存 / SQL 迁移执行器
│   │   ├── config/ audit/      #   配置 / 审计
│   │   ├── storage/ logger/    #   文件存储 / 日志
│   │   └── extapi/             #   跨模块 Provider 接口
│   ├── internal/               # 仅 framework 内可见（Go internal/ 规则）
│   │   ├── core/               #   boot / server / middleware / ext_impl
│   │   ├── service/            #   框架级服务（Authz / Perm）
│   │   └── module/             #   仍在 framework 内的业务模块（Phase 3 待迁出）
│   └── xin-server.service
│
└── apps/                       # 独立 Go module: gx1727.com/xin/apps
    ├── boot/{auth,tenant}/     # Phase 2 已迁入
    ├── rbac/{user,role,menu,resource,permission,organization}/   # Phase 3 待迁入
    ├── reference/{dict,asset,weixin}/                             # Phase 3 待迁入
    ├── cms/                    # 已就位
    └── flag/                   # 已就位
```

**为什么用三个 module 而不是单体**：

- `framework/` 必须独立——它是"框架核心"，可能单独发版或被第三方 import
- `apps/` 必须独立——业务模块如果与 framework 在同一 module，cross-import 就完全失控
- 根 module 仅承载 `cmd/xin/main.go`，把"产品入口"和"框架/业务"分开

代价：每个 module 自己一份 `go.sum`，加新依赖要改两次。Go 1.21+ 的 `go.work sync` 能缓解。

---

## 2. 模块的生命周期

```
side-effect import
  └─> 包 init()
        └─> plugin.Register(Module())
              └─> framework.Run()
                    └─> initModules(cfg)
                          └─> for m in plugin.Apps():
                                if cfg.Module contains m.Name():
                                  m.Init()
                                  ↓ 之后 ↓
                                  m.Register(public, protected)
```

每个模块实现 `framework/pkg/plugin.Module` 接口：

```go
type Module interface {
    Name() string
    Init() error
    Register(public, protected *gin.RouterGroup)
    Shutdown() error
}
```

`plugin.Register()` 对同名模块只注册一次，所以误重复 import 不会 panic。

---

## 3. 路由分层

```
/api/v1/
├── public/                      # OptionalAuth 中间件（可选登录）
│   └── <module-public-routes>   # 如 /auth/login, /dicts/:code
└── protected/                   # Auth 中间件（强制登录 + 注入 XinContext）
    └── <module-protected-routes>   # 如 /users, /roles, /cms/articles
```

每个模块在 `Register(public, protected, h)` 里自挂路由，公共前缀由模块自己起（如 `protected.Group("/users")`）。

全局中间件顺序：

```
Recovery → RequestID → CORS → ClientIP → Logger → (OptionalAuth | Auth)
```

---

## 4. 跨模块依赖规则

| 方向 | 规则 | 例 |
| --- | --- | --- |
| apps → framework/pkg | ✅ 直接 import | `apps/boot/auth` → `framework/pkg/auth` |
| apps → framework/internal | ❌ internal/ 规则拒绝 | —— |
| framework/internal → apps | ❌ 同上 | —— |
| framework/internal → apps（间接） | ✅ 通过 framework/pkg 的注册钩子 | framework/user 通过 `pkgauth.Get()` 拿到 apps/boot/auth 的 AccountRepository |
| apps/X → apps/Y | ✅ 同 module | `apps/boot/auth` → `apps/boot/tenant` |

**注册钩子模式**（[framework/pkg/auth/registry.go](file:///d:\work\xin\XinFramework\server\framework\pkg\auth\registry.go) ↔ [apps/boot/auth/module.go](file:///d:\work\xin\XinFramework\server\apps\boot\auth\module.go)）：

```go
// 1. framework/pkg/auth 定义接口 + 注册函数
package auth
var globalFactory func() AccountRepository
func Register(f func() AccountRepository) { globalFactory = f }
func Get() func() AccountRepository { return globalFactory }

// 2. apps/boot/auth 在 init() 里推入自己的实现
func init() {
    pkgauth.Register(func() pkgauth.AccountRepository {
        return NewAccountRepository(db.Get())
    })
}

// 3. framework/internal/module/user 用 Get() 取
import pkgauth "gx1727.com/xin/framework/pkg/auth"
repo := pkgauth.Get()()  // 注意：nil 检查
```

---

## 5. 统一响应与错误码

所有 handler 用 `framework/pkg/resp`：

```go
resp.OK(c, data)                          // 成功
resp.Fail(c, resp.ErrBadRequest, msg)     // 业务失败
```

响应体固定 `{code, msg, data}`，前端 `api/client.ts` 解析一致。

错误码分段：

| 段 | 含义 | 示例 |
| --- | --- | --- |
| 0 | 成功 | —— |
| 4xxx | 客户端错误 | `ErrBadRequest=4000`、`ErrUnauthorized=4001` |
| 5xxx | 服务端错误 | `ErrInternal=5000` |
| 9xxx | 业务错误 | `ErrUserNotFound=9001`、`ErrRoleExists=9100` |

详见 [resp/errors.go](file:///d:\work\xin\XinFramework\server\framework\pkg\resp\errors.go)。

---

## 6. 重构方案（Phase 1-5）

### 6.1 问题的提出

旧的 `framework/internal/module/*` 同时塞了 auth / user / tenant / dict / cms 等业务模块。结果：

- 业务代码被 Go 的 `internal/` 规则锁死在 framework module 内，**用户不能 fork / 覆盖**
- main.go 里硬编码 `builtinMap` + `appsRegistry` 双轨注册
- `apps/cms` 和 `apps/flag` 各自独立 go.mod，依赖重复声明
- framework 的 `Plugin` 机制是"假"的——main.go 编译期就知道所有 app

详细分析见 [Phase 0 之前的对话历史](file:///d:\work\xin\XinFramework\README.md)（git log 也保留）。

### 6.2 Phase 1：统一注册

**目标**：删 `builtinMap`，所有模块走 `plugin.Apps()`。

- ✅ `framework/pkg/plugin/plugin.go` 加重名保护
- ✅ 每个内置模块 `init() { plugin.Register(Module()) }`
- ✅ `framework/builtin_modules.go` 用 side-effect 解决 internal/ 限制
- ✅ `cmd/xin/main.go` 简化为只 import + framework.Run

### 6.3 Phase 2：auth / tenant 出 framework

**目标**：把 framework 启动期必须的 auth + tenant 搬到 `apps/boot/`。

- ✅ apps/boot/auth 复制过来 + 改 import path
- ✅ apps/boot/tenant 复制过来 + 改 import path
- ✅ HashPassword 上提到 `framework/pkg/auth/`
- ✅ RequirePlatformRole 上提到 `framework/pkg/middleware/`
- ✅ 注册钩子：`framework/pkg/auth.Register` ↔ `apps/boot/auth.init`
- ✅ 注册钩子：`framework/pkg/tenant.Register` ↔ `apps/boot/tenant.init`
- ✅ `framework/internal/core/ext_impl` 通过钩子拿 tenant 数据
- ✅ `framework/internal/module/user`、`weixin` 改用 `pkgauth` / `pkgtenant` 接口

**结果**：auth 和 tenant 完全可被外部 fork / 覆写，framework 不再锁住这两块核心业务逻辑。

### 6.4 Phase 3：RBAC + reference 全部出 framework（待执行）

待办：
- `framework/internal/module/{user,role,menu,resource,permission,organization}` → `apps/rbac/<name>/`
- `framework/internal/module/{dict,asset,weixin}` → `apps/reference/<name>/`
- 沿用 Phase 2 的注册钩子模式

完成后 `framework/internal/module/` 完全清空，framework 真正变成"纯基础设施"。

### 6.5 Phase 4：apps module 合并（已部分执行）

- ✅ 删除 `apps/cms/go.mod`、`apps/flag/go.mod`
- ✅ 创建 `apps/go.mod`（`gx1727.com/xin/apps`）
- ✅ `root/go.mod` 的 replace 改为 `apps => ./apps`

加新 app 不再需要建 go.mod，直接 `apps/<x>/`。

### 6.6 Phase 5：动态插件发现（可选）

- 用 build tag 区分商业版 / 社区版
- 或用 `go plugin.Open(.so)` 运行时加载

Phase 5 不是必需——当前 side-effect import 已经够用且 0 运行时成本。

---

## 7. 关键文件位置

| 关注点 | 路径 |
| --- | --- |
| 入口 | [cmd/xin/main.go](file:///d:\work\xin\XinFramework\server\cmd\xin\main.go) |
| 框架入口 | [framework/framework.go](file:///d:\work\xin\XinFramework\server\framework\framework.go) |
| 内置模块汇总 | [framework/builtin_modules.go](file:///d:\work\xin\XinFramework\server\framework\builtin_modules.go) |
| 注册中心 | [framework/pkg/plugin/plugin.go](file:///d:\work\xin\XinFramework\server\framework\pkg\plugin\plugin.go) |
| 启动流程 | [framework/internal/core/boot/boot.go](file:///d:\work\xin\XinFramework\server\framework\internal\core\boot\boot.go) |
| 全局中间件 | [framework/internal/core/middleware/](file:///d:\work\xin\XinFramework\server\framework\internal\core\middleware) |
| 公开权限中间件 | [framework/pkg/middleware/auth.go](file:///d:\work\xin\XinFramework\server\framework\pkg\middleware\auth.go) |
| auth 钩子 | [framework/pkg/auth/registry.go](file:///d:\work\xin\XinFramework\server\framework\pkg\auth\registry.go) |
| tenant 钩子 | [framework/pkg/tenant/registry.go](file:///d:\work\xin\XinFramework\server\framework\pkg\tenant\registry.go) |
| 多 module 编排 | [go.work](file:///d:\work\xin\XinFramework\server\go.work) |
| Root module | [go.mod](file:///d:\work\xin\XinFramework\server\go.mod) |
| Apps module | [apps/go.mod](file:///d:\work\xin\XinFramework\server\apps\go.mod) |