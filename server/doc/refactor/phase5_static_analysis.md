# Phase 5 静态分析报告

> 时间:Phase 4 完成后,Phase 5 开始前
> 范围:`framework/pkg/authz` + `framework/internal/service` + `framework/internal/core/boot::globalApp`
> 目标:在动手 Phase 5 之前,穷举所有调用点,识别风险

---

## 一、当前全局状态图

```
┌──────────────────────────────────────────────────────────────────────────┐
│                           boot.Init()                                     │
│                                                                          │
│  ① NewPermissionService(...)            ← service 包                    │
│  ② NewAuthorizationService(perm)       ← service 包                    │
│  ③ service.SetGlobalAuthorizationService(authzService)  ← 写入全局      │
│  ④ authz.Set(authz.Wrap(authzService))  ← 写入全局                      │
│  ⑤ globalApp = &App{..., Authz: authzService}                          │
└────────────────────┬──────────────────────────────────┬─────────────────┘
                     │                                  │
                     ▼                                  ▼
       ┌─────────────────────────┐        ┌──────────────────────────────┐
       │ service.globalAuthz     │        │ authz.global                 │
       │   = *AuthorizationService│        │   = Authorization (interface)│
       └──────────┬──────────────┘        └──────────┬───────────────────┘
                  │                                  │
       ⑥ service.GlobalAuthorizationService()       ⑧ authz.Get()
                  │                                  │
                  ▼                                  ▼
       ┌──────────────────────────────────────────────────────────────┐
       │  framework 内: middleware/auth.go (Require/RequireAll/...) │
       │    通过 service.GlobalAuthorizationService() 拿              │
       │  apps 内: role/permission/resource service.go              │
       │    通过 authz.Get() 拿(14 处调用)                          │
       └──────────────────────────────────────────────────────────────┘

       ⑨ boot.globalApp = *App  ← 框架内部 / 业务模块没发现使用
```

> ⚠️ **关键发现**:`service.GlobalAuthorizationService()` 在 framework **0 处** 调用!但 service 包对外提供了它。
>
> 摸底脚本数 `globalAuthorizationService = 5 reads`,这是**自我引用 + 函数定义本身**。
> 实际"被外部读取"是 **0 次**(待确认)。

---

## 二、framework/pkg/authz 静态分析

### 2.1 接口与全局

| 项目 | 内容 |
|---|---|
| 类型 | `Authorization interface` (6 个方法) |
| 全局变量 | `var global Authorization` (line 50) |
| 写入 API | `authz.Set(a Authorization)` |
| 读取 API | `authz.Get() Authorization` |
| Wrapper | `authz.Wrap(inner) Authorization` (用于适配 *DataScope 返回类型) |

### 2.2 调用点全表

#### 写入端(`authz.Set` / `authz.Wrap`)

| 文件 | 行 | 调用 | 作用 |
|---|---:|---|---|
| `framework/internal/core/boot/boot.go` | 72 | `authz.Set(authz.Wrap(authzService))` | boot 时 wire 一次 |

**写入端唯一** = boot.go

#### 读取端(`authz.Get`)

| 文件 | 行 | 上下文 |
|---|---:|---|
| `apps/rbac/role/service.go` | 107 | 角色删除 → 缓存失效 |
| `apps/rbac/role/service.go` | 135 | 角色权限更新 → 缓存失效 |
| `apps/rbac/role/service.go` | 147 | 角色状态更新 → 缓存失效 |
| `apps/rbac/role/service.go` | 166 | 批量更新用户角色 → 用户缓存失效 |
| `apps/rbac/role/service.go` | 198 | 角色绑用户 → 缓存失效 |
| `apps/rbac/resource/service.go` | 90 | 资源删除 → 缓存失效 |
| `apps/rbac/resource/service.go` | 99 | 资源更新 → 缓存失效 |
| `apps/rbac/permission/service.go` | 42 | 角色权限分配 → 缓存失效 |

**读取端 = 8 处,全部在 apps/,全部用于缓存失效**

**注释引用**(不算):authz.go 自身有 5 处注释提到 `authz.Get()`,plan.md 有 1 处。

### 2.3 现状特征

- **全部 8 处 Get 都是缓存失效**:`InvalidateUser` / `InvalidateRole` / `InvalidateResource`
- **全部都是 nil-safe**:每个调用都包在 `if authz := authz.Get(); authz != nil` 里
- **都是 fire-and-forget**:`_ = authz.InvalidateRole(context.Background(), ...)` 用 background ctx

**意味着**:authz 失效逻辑是"可选的 best-effort",即使拿不到 authz 也不会让请求失败。

---

## 三、framework/internal/service 静态分析

### 3.1 三个文件

| 文件 | 导出 | 角色 |
|---|---|---|
| `authorization_service.go` | `AuthorizationService` (struct) | 门面:对 apps 暴露 |
| `permission_service.go` | `PermissionService` (struct) | 实际查询 + 缓存 |
| `authorization_service.go` | 全局 `globalAuthorizationService` + `Set/Get` | 跨包访问入口 |

### 3.2 `*AuthorizationService` API 全表

```go
LoadPermissions(ctx, userID) map[string]bool, error         // apps/permission/resource 通过 authz 调
LoadRoles(ctx, userID) []string, error                      // 同上
LoadDataScope(ctx, userID) *permission.DataScope, error     // apps 通过 type assertion 拿
LoadUserSecurityContext(ctx, userID) → (perms, roles, ds, orgID, err)  // framework 内部
Can(ctx, userID, spec) bool, error                          // framework middleware 调
BuildScopeFilter(ctx, userID, columns) ScopeFilter, error   // framework 内部
BuildDataScopeSQL(ctx, userID) (sql, args, err)             // framework 内部
InvalidateUser(ctx, userID) error                           // apps + framework
InvalidateRole(ctx, roleID) error                           // apps + framework
InvalidateResource(ctx, resourceID) error                   // apps + framework
```

### 3.3 调用点全表

#### 写入端(`service.SetGlobalAuthorizationService` / `service.NewAuthorizationService` / `service.NewPermissionService`)

| 文件 | 行 | 调用 |
|---|---:|---|
| `framework/internal/core/boot/boot.go` | 62 | `service.NewPermissionService(...)` |
| `framework/internal/core/boot/boot.go` | 68 | `service.NewAuthorizationService(permService)` |
| `framework/internal/core/boot/boot.go` | 69 | `service.SetGlobalAuthorizationService(authzService)` |

**写入端唯一** = boot.go (3 处)

#### 读取端(`service.GlobalAuthorizationService()`)

| 文件 | 行 | 上下文 |
|---|---:|---|
| (无匹配) | — | framework **0 处**, apps **0 处**(apps 不能 import internal) |

**关键发现**:`service.GlobalAuthorizationService()` 在整个 codebase **0 次调用**!

摸底脚本数 5 reads 是因为它把"方法定义本身"和"全局变量声明"算进去了。

#### 其他读取端(`Authz` 字段、`PermService` 字段)

`App` struct 有 `Authz *service.AuthorizationService` 和 `PermService *service.PermissionService` 字段,但 `boot.AppInstance()` 是 0 次调用。

---

## 四、framework/internal/core/boot::globalApp 静态分析

### 4.1 全局状态

```go
var globalApp *App  // line 32

func Init(cfg *config.Config) (*App, error) {
    // ...
    globalApp = &App{...}  // line 74
}

func AppInstance() *App {  // line 93
    return globalApp
}
```

### 4.2 调用点

| 调用 | 位置 | 计数 |
|---|---|---:|
| `boot.AppInstance()` | (全 codebase) | 0 |
| `boot.globalApp` | (全 codebase) | 0 |
| `boot.Init` 返回值 *App | framework.go | 1(框架主入口拿 app 后调 `app.Server.Start()`) |
| `app.Authz` 字段 | `framework.go:157, 160` | 2(middleware 注入) |

**关键发现 1**:`boot.globalApp` 实际上**完全没用**!`Init` 返回值已经够用,`AppInstance()` 函数是 dead code。

**关键发现 2**:`app.Authz` 字段被 `framework.go:157,160` 用来注入到 middleware。这是 framework 内部传参,**不依赖任何全局**。也就是说:

- `service.SetGlobalAuthorizationService(authzService)` 调用是 **写完就扔**
- `app.Authz` 直接从 `authzService` 变量赋值,不读全局
- 中间件通过 `app.Authz` 拿到 service,再以参数形式传到 `middleware.Auth()`

### 4.3 结论

`service.globalAuthorizationService` + `SetGlobalAuthorizationService()` + `GlobalAuthorizationService()` 三件套可以**直接删除**,不需要任何修改外部代码。

`boot.globalApp` + `AppInstance()` 也可以**直接删除**。

`authz.Set(...)` + `authz.Get()` 必须**迁移**到 AppContext(8 处 apps 调用需要从 `ctx.Authz()` 拿)。

---

## 五、Phase 5 改动清单(基于静态分析)

### 5.1 必做改动

| # | 改动 | 工作量 | 风险 |
|---|---|---|---|
| 1 | `apps/rbac/role/service.go` 8 处 → 注入 Authz | 0.5h | 低:8 处都是 cache 失效 |
| 2 | `apps/rbac/resource/service.go` 2 处 → 注入 Authz | 0.3h | 低 |
| 3 | `apps/rbac/permission/service.go` 1 处 → 注入 Authz | 0.1h | 低 |
| 4 | `framework/internal/core/middleware/auth.go` 用 ctx.Authz() 替代 `service.GlobalAuthorizationService()` | 0.3h | 中:替换路径 |
| 5 | `framework/internal/core/boot/boot.go` 把 authzService 写到 `appCtx.SetAuthz(...)` | 0.2h | 低 |
| 6 | `framework/pkg/authz/authz.go` 删 `global` / `Set` / `Get` | 0.2h | 低(无外部调用) |
| 7 | `framework/internal/service/authorization_service.go` 删 `globalAuthorizationService` / `Set/Get` | 0.1h | 低 |
| 8 | `framework/internal/core/boot/boot.go` 删 `globalApp` / `AppInstance()`(dead code) | 0.1h | 低 |

### 5.2 注入策略选择

**问题**:AuthzService 是个有状态、有方法的对象。8 处 apps 缓存失效调用 + 1 处 middleware 调用,怎么注入?

**方案对比**:

#### 方案 A:构造时注入(推荐)

```go
// apps/rbac/role/module.go
RegFn: func(ctx, public, protected) {
    authz := ctx.Authz()
    if authz == nil {
        log.Fatal("authz not loaded")  // boot 一定先注入,失败是配置错
    }
    svc := NewService(..., authz)
    Register(protected, h)
}

// apps/rbac/role/service.go
type Service struct {
    authz authz.Authorization  // 新增字段
}

func (s *Service) updateRole(...) error {
    if s.authz != nil {
        _ = s.authz.InvalidateRole(ctx.Background(), id)
    }
}
```

**优点**:显式依赖,编译期约束
**缺点**:8 处 service.go 都要改 constructor

#### 方案 B:中间件式 ServiceProvider

```go
// 在 framework/pkg/authz 加 lazy 上下文访问器
type CtxKey struct{}
func WithAuthz(ctx, authz) ctx { return context.WithValue(ctx, ctxKey, authz) }
func From(ctx) Authorization { return ctx.Value(ctxKey).(Authorization) }

// apps/rbac/role/service.go
func (s *Service) updateRole(...) error {
    if a := authz.From(ctx); a != nil {  // 业务 ctx 已注入
        _ = a.InvalidateRole(ctx.Background(), id)
    }
}
```

**优点**:不改 constructor,只在 service 方法里取
**缺点**:增加 context 传递复杂度,每个 handler 要 wire

**决策**:选 **方案 A**。理由:
1. `role/service.go` 等的 `NewService(...)` 已经显式列参数,加一个 `authz` 不破坏风格
2. Authz 在 boot 阶段一定存在,Register 阶段 `ctx.Authz()` 必然 non-nil
3. 测试时可以传 nil 或 mock,无需 context plumbing

### 5.3 框架 middleware 的 Authz 访问

`framework/internal/core/middleware/auth.go` 当前怎么拿 Authz?让我看。

实际是:`middleware/auth.go` 通过 `authz.Get()` 拿 framework 公共接口(14 处摸底计数主要来自这里)。

> 等下 — grep 没看到 `middleware/auth.go` 里 `authz.Get`。让我重新检查。

实际再扫一遍 framework pkg/authz 调用点。

---

## 六、middleware/auth.go 静态分析(补充)

### 6.1 调用 authz 的位置

实际不在 `framework/internal/core/middleware/auth.go`,而是可能在 `framework/pkg/middleware/auth.go`。让我先确认。

需要再 grep 一次 `framework/pkg/middleware/auth.go`。

### 6.2 注入方式

middleware 是跨 module 共享的"框架级"组件。两种方案:
- a) 每个 middleware 函数接收 `authz` 参数(函数式)
- b) middleware 在创建时持有 `authz`(闭包 / struct)
- c) middleware 从 RequestContext 拿(Phase 2 已经预留 `xin.Context`)

**决策**:保持**现状**(middleware 用 `service.GlobalAuthorizationService()`)— 这是 framework 内部的全局,**Phase 5 不必清这个**。

理由:`service.GlobalAuthorizationService()` 在 framework 内部,有 `internal/` 边界保护,不构成跨包/跨模块污染。Phase 5 主要是清 apps 端的 `authz.Get()`。

但**全局本身应该清掉**,改成 `appCtx.Authz()`:
1. 删 `service.globalAuthorizationService` + `Set/Get`
2. middleware 从 `xin.Context` 拿,或者从 `appCtx` 拿(需注入 middleware factory)

---

## 七、风险评估与缓解

| 风险 | 等级 | 缓解 |
|---|---|---|
| apps 8 处 authz.Get 漏改 | 中 | 写完后 grep `authz.Get()` 必须命中 0 行 |
| middleware 拿不到 Authz | 中 | boot 顺序:先 SetAuthz,再 setup middleware |
| 测试覆盖 | 低 | Phase 5 不强求,但 Phase 8 必加 |
| `service.GlobalAuthorizationService` 误删导致 framework 编译失败 | 高 | **先 grep `GlobalAuthorizationService()`,如果 0 处调用才能删** |

---

## 八、Phase 5 推荐执行顺序

```
Step 1: 改 boot.go: appCtx.SetAuthz(authz.Wrap(authzService))
        (注: 已经有 SetAuthz,改写时机)
Step 2: 删 framework/pkg/authz/authz.go::global + Set + Get
Step 3: 删 framework/internal/service/authorization_service.go::globalAuthz + Set/Get
Step 4: 改 framework/internal/core/middleware/auth.go (如有)
Step 5: 改 apps/rbac/role/service.go (5 处)+ module.go (注入)
Step 6: 改 apps/rbac/resource/service.go (2 处) + module.go
Step 7: 改 apps/rbac/permission/service.go (1 处) + module.go
Step 8: 删 framework/internal/core/boot/boot.go::globalApp + AppInstance() (dead code)
Step 9: build + vet + grep + 摸底回归
```

每步独立编译通过,build exit 0 才能进下一步。

---

## 九、Phase 5 验收标准(最终)

- [ ] `grep "authz.Get()" -rn server/apps/` 命中 0 行
- [ ] `grep "authz.Set(" -rn server/` 命中 0 行
- [ ] `grep "service.GlobalAuthorizationService" -rn server/` 命中 0 行
- [ ] `grep "service.SetGlobalAuthorizationService" -rn server/` 命中 0 行
- [ ] `grep "boot.AppInstance\|boot.globalApp" -rn server/` 命中 0 行
- [ ] 摸底脚本:`authz.global`、`service.globalAuthorizationService`、`boot.globalApp` 引用数 = 0
- [ ] `go build ./...` 退出码 0
- [ ] `go vet ./...` 退出码 0
- [ ] 8 处 apps 调用全部能编译(类型正确)

---

## 十、依赖此分析的关键决策

| 决策 | 依据 |
|---|---|
| **构造时注入**(方案 A) | 8 处调用都是 cache 失效,fire-and-forget,无 context 依赖 |
| **删 `boot.globalApp` 顺手做** | grep 全 codebase 0 次外部使用,dead code |
| **`service.globalAuthz` 必删** | 摸底 5 reads 中 4 是函数自身定义 + 注释,真正外部调用 0 次 |
| **middleware 拿 Authz 的方式待定** | 需进一步 grep `framework/internal/core/middleware/auth.go` 和 `framework/pkg/middleware/auth.go` |
| **Phase 5 拆 9 个原子步骤** | 每个步骤独立 compile,失败回退成本低 |