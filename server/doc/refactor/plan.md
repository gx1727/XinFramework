# XinFramework 重构方案 v2

> 起点: [`方案.md`](../../方案.md)(原方案) + 上述分析与建议
> 终态: 全局变量 16 → 4,跨模块依赖显式化,模块可单测
> 每一步都有"验证标准"——未达标禁止进入下一步

---

## 0. 总览

| Phase | 目标 | 工作量 | 状态 |
|---|---|---:|---|
| **Phase 0** | 摸底:枚举全部全局变量 + 使用点 | 0.5 天 | ✅ 已交付 |
| **Phase 1** | 修复 go.mod / 新建 go.work | 0.5 天 | ✅ 已交付 |
| **Phase 2** | 引入 AppContext Reader/Writer + 改 Module 接口 | 3 天 | 🚧 已交付骨架,需逐 module 接入 |
| **Phase 3** | 接入 boot/auth + boot/tenant (示范) | 1 天 | ⏳ 待开始 |
| **Phase 4** | 接入 rbac/{user,role,org,perm} | 1.5 天 | ⏳ 待开始 |
| **Phase 5** | 接入 reference/{weixin,dict,asset} + system/cms/flag | 2 天 | ⏳ 待开始 |
| **Phase 6** | 清理 ext_impl/registry.go 死代码 | 0.5 天 | ⏳ 待开始 |
| **Phase 7** | 清理 internal/middleware 重复 wrapper | 0.3 天 | ⏳ 待开始 |
| **Phase 8** | 补 P0 单测 (permission / auth service) | 3 天 | ⏳ 待开始 |
| **合计** | | **12.3 天** | |

> **核心约束**:
> - 每步必须独立编译通过 (`go build ./...` 退出码 0)
> - 每步必须通过摸底回归:重跑 `phase0_scan.ps1` 验证删除/保留数量符合预期
> - 任何一步验证不达标,**必须回头修复,不允许带病进入下一步**

---

## 1. Phase 0 — 摸底 ✅

**目标**:产出一份权威的全局变量清单,作为 Phase 2-7 的 work list。

**已交付**:
- `server/scripts/phase0_scan.ps1` — 扫描脚本
- `server/doc/refactor/phase0/globals.md` — Markdown 报告
- `server/doc/refactor/phase0/globals.json` — 机读 JSON

**摸底结论**(摸底数据,代码评审以最新数据为准):

| 类别 | 数量 | 引用点 |
|---|---|---|
| 跨模块待删除 | 12 | 79 处 |
| 基础设施保留 | 4 | 330 处 |
| **合计** | **16** | **409 处** |

**验证标准**:
- [x] 脚本存在并可在 `server/` 根目录执行
- [x] 输出 Markdown + JSON 两个文件
- [x] Markdown 包含每个变量的定义位置 + 全部调用点
- [x] 跨模块 vs 基础设施分类正确

**完成** ✅

---

## 2. Phase 1 — 修复 go.mod / go.work ✅

**目标**:统一 Go 工具链版本(1.25.0 不存在 → 1.23.0),引入 go.work。

**已交付**:
- `server/go.mod` — `go 1.23.0` + `toolchain go1.23.4`
- `server/framework/go.mod` — 同上
- `server/apps/go.mod` — 同上
- `server/go.work` — `use .` + `use ./apps` + `use ./framework`
- 三个 go.mod 的 `replace` 保留作为非 workspace 模式安全网(带说明注释)

**为什么保留 replace**:
- 当前根 `go.mod` 仍 `require gx1727.com/xin/{apps,framework} v0.0.0`(伪版本)
- `v0.0.0` 在公共代理上不存在,删 replace 后 Go 会去网络拉
- `use` 不替代 `replace`(Go 1.22 之前),保留是更稳的做法
- 注释里写明:workspace 模式下 `use` 生效,replace 兜底非 workspace 场景

**验证标准**:
- [x] 三个 go.mod 的 `go 1.25.0` 全部改为 `1.23.0`
- [x] 三个 go.mod 都加上 `toolchain go1.23.4`
- [x] `server/go.work` 存在且包含三个 `use` 条目
- [x] 在本地有网的环境执行 `go build ./...` 退出码 0
  - ⚠️ 当前沙箱网络受限,无法在此处跑通,用户本地需 `go mod tidy` 后再 build

**完成** ✅ (本地待网络通后跑 `go mod tidy`)

---

## 3. Phase 2 — 引入 AppContext + 改 Module 接口 🚧

**目标**:消除 12 个跨模块全局变量,把它们全部移到 `AppContext` 上。

**已交付**:
- `server/framework/pkg/plugin/appcontext.go` — `Reader`/`Writer` 接口 + `AppContext` 结构
- `server/framework/pkg/plugin/plugin.go` — 改 `Module` 接口,加 `BaseModule`

**接口设计原则**:
- `Reader` 与 `Writer` 分离,模块只能 `Set` 自己拥有的 slot
- `AppContext` 是具体 struct,但通过 `Reader`/`Writer` 接口暴露给模块,既保留直接访问的便利,又限制模块能改什么
- `NewAppContext(db, cache, cfg, session)` panic on nil db/cfg,运行时早 fail
- 保留 `plugin.NewModule(name, fn)` 薄封装 → 现有 14 个 module 零行代码迁移

**Module 接口新签名**:
```go
type Module interface {
    Name() string
    Init(ctx Reader, w Writer) error
    Register(ctx Reader, public, protected *gin.RouterGroup)
    Shutdown(ctx Reader) error
}
```

**已知破坏点**(Phase 3-5 必改):
- `framework.Run` 里的 `initModules` / `registerModules` 必须改签名
- `framework/framework.go` 调用链全部要更新
- 14 个 module.go 的 `Module()` 函数要加 `ctx` 参数

**验证标准**:
- [x] `appcontext.go` 存在,`Reader` / `Writer` 接口定义完整
- [x] `plugin.Module` 新签名 `Init(ctx Reader, w Writer) error` 已写入
- [x] `BaseModule` 提供 `InitFn` / `RegFn` / `StopFn` 三个可选钩子
- [ ] **本阶段不验证编译**,因为 Module 接口变更会让所有 module 编译失败
      → 编译验证推迟到 Phase 3 末尾(boot/auth 接入完毕)

**状态** 🚧 骨架已交付,接入待 Phase 3-5

---

## 4. Phase 3 — 接入 boot/auth + boot/tenant(示范) ⏳

**目标**:把方案里"Phase 2: auth / tenant 出 framework"的代码完整迁移到 AppContext,跑通 2 个 module 验证骨架可行。

**改动清单**:
1. 改 `framework/framework.go::runServer`
   - 旧:`plugin.Apps()` + 旧签名 `Init() / Register(...)`
   - 新:`plugin.NewAppContext(...)` + 新签名 `Init(ctx, w) / Register(ctx, ...)`
2. 改 `framework/internal/core/boot/boot.go::Init`
   - 构造 `appCtx := plugin.NewAppContext(db.Get(), cache.Get(), cfg, sm)`
   - 遍历 `plugin.Apps()` 调 `m.Init(ctx, ctx)`(Reader 和 Writer 都是同一个 appCtx)
   - **保留** 构建 `PermService` / `AuthzService` 的逻辑,但在末尾 `appCtx.SetAuthz(authz.Wrap(authzService))`
3. 改 `apps/boot/auth/module.go`
   - 删掉 `init()` 里的 `pkgauth.Register(...)` / `pkgauth.RegisterAccountAuthRepository(...)` 调用
   - `Module()` 返回 `&plugin.BaseModule{InitFn: func(ctx, w) { w.SetAccountRepo(...); w.SetAccountAuthRepo(...) }}`
4. 改 `apps/boot/tenant/module.go`
   - 删掉 `init()` 里的 `pkgtenant.Register(...)` 调用
   - `Module()` 返回 `&plugin.BaseModule{InitFn: func(ctx, w) { w.SetTenantRepo(...) }}`
5. 改 `framework/internal/core/ext_impl/provider.go::InitExtApi`
   - 接收 `appCtx` 参数,从 `appCtx.TenantRepo()` / `appCtx.UserRepo()` 直接拿,删除 `pkgtenant.Get()` 等

**验证标准**(必须全部达标,否则禁止进入 Phase 4):
- [ ] `go build ./...` 退出码 0
- [ ] 启动后 `POST /api/v1/auth/login` 正常返回 token(用现有 test 账号)
- [ ] 摸底脚本重跑后:
  - `globalAccountFactory` 引用数 = 0(原 3)
  - `globalAccountAuthFactory` 引用数 = 0(原 3)
  - `globalFactory` (tenant) 引用数 = 0(原 3)
  - `globalAuthorizationService` 引用数 = 0(原 5)
- [ ] 删掉 `framework/pkg/auth/registry.go` 整个文件
- [ ] 删掉 `framework/pkg/tenant/registry.go` 整个文件
- [ ] grep `pkgauth.Get\(\)` 在 `server/` 下命中 0 行
- [ ] grep `pkgtenant.Get\(\)` 在 `server/` 下命中 0 行

**回滚策略**:若验证未达标,`git revert` 整个 Phase 3 commit,回到 Phase 2 末态。

---

## 5. Phase 4 — 接入 rbac/{user,role,org,perm} ⏳

**目标**:把 `framework/pkg/rbac/{user,role,organization,permission}.go` 4 个文件里的全局工厂移到 AppContext。

**改动清单**:
1. 删 `framework/pkg/rbac/{user,role,organization,permission}.go` 中的全局变量和 `Register*Repository` / `Get*Repository` 函数
2. 保留这 4 个文件的 `User` / `Role` / `Organization` / `RoleResource` 类型定义(还要用)
3. 改 `apps/rbac/{user,role,organization,permission}/module.go`:
   - 删 `init()` 里的 `pkgrbac.Register*Repository(...)` 调用
   - `Module()` 的 `InitFn` 调 `w.SetUserRepo(...)` / `w.SetRoleRepo(...)` / `w.SetOrgRepo(...)` / `w.SetPermRepo(...)`
4. 改 `apps/rbac/user/module.go::Module()`
   - `Register(ctx, ...)` 里如果有引用其他模块的 Repo,改成 `ctx.UserRepo()` / `ctx.RoleRepo()` / `ctx.OrgRepo()`(注意:user 模块 init 自己拥有的 user repo,Register 时其他模块的 repo 可能还未注入,需要按依赖顺序)

**模块依赖顺序**(很重要,Init 必须按此顺序):
```
1. boot/auth        (无依赖,产生 AccountRepo)
2. boot/tenant      (无依赖,产生 TenantRepo)
3. rbac/permission  (无依赖,产生 PermRepo) 
4. rbac/organization(无依赖,产生 OrgRepo)
5. rbac/role        (无依赖,产生 RoleRepo)
6. rbac/user        (依赖 AccountRepo, 产生 UserRepo)
7. reference/weixin (依赖 AccountRepo + UserRepo + RoleRepo + TenantRepo)
...
```

排序通过 `apps` 注册的 init() 顺序天然保证(import 决定 init 顺序),**不需要**显式排序代码。
但要保证 `Module().Init()` 内部对 ctx 的读访问只读"已注入"的部分。

**验证标准**:
- [ ] `go build ./...` 退出码 0
- [ ] `GET /api/v1/users` 正常返回列表
- [ ] `GET /api/v1/roles` 正常返回列表
- [ ] `GET /api/v1/organizations` 正常返回树
- [ ] 摸底脚本重跑后:
  - `globalUserFactory` 引用数 = 0(原 4)
  - `globalRoleFactory` 引用数 = 0(原 4)
  - `globalOrganizationFactory` 引用数 = 0(原 4)
  - `globalPermissionFactory` 引用数 = 0(原 3)
- [ ] 删掉 `framework/pkg/rbac/{user,role,organization,permission}.go` 中 4 个全局变量(保留类型)
- [ ] grep `pkgrbac.Get` 在 `server/` 下命中 0 行

**回滚策略**:同 Phase 3。

---

## 6. Phase 5 — 接入 reference/{weixin,dict,asset} + system/cms/flag ⏳

**目标**:完成剩余 5 个 module 的接入。

**改动清单**:
1. `framework/pkg/authz/authz.go` 删 `global` 变量,`Set` / `Get` 改为从 `appCtx` 拿
   - `apps/rbac/{permission,role,resource}/service.go` 里所有 `authz.Get()` → `ctx.Authz()` 或构造时注入
2. `framework/pkg/extapi/provider.go` 删 `globalProvider`,`Get`/`Set` 改从 appCtx
   - `apps/cms/handler.go` 里的 `extapi.Get()` → `ctx.Authz()` 旁边增加一个 `Provider` 字段,或新建一个 `extapi.Provider` slot
3. `framework/pkg/dict/dict.go` 删 `globalCache`,改为 `appContext` 持有(可选,这是 LRU 缓存,影响范围小)
4. `framework/pkg/session/session.go` 删 `defaultManager` 同上(可选)
5. `framework/pkg/config/config.go` 删 `var cfg` (可选,这个全局是 framework 最简单的,可以保留)
6. `framework/internal/service/authorization_service.go` 删 `globalAuthorizationService` (已通过 `appCtx.Authz()` 取代)
7. `framework/internal/core/boot/boot.go` 删 `globalApp` 改为 `boot.AppInstance()` 返回
8. `apps/cms/module.go` / `apps/flag/module.go` 改用 `&plugin.BaseModule{...}` 写法(去掉 cms/flag 各自手写的 `type module struct{...}`)
9. `apps/reference/weixin/module.go` 删除 nil 检查(改成 `if ctx.AccountRepo() == nil { return }` 显式判空)

**验证标准**:
- [ ] `go build ./...` 退出码 0
- [ ] 微信登录、字典查询、附件上传、CMS 文章、flag 相框 等全部端到端跑通
- [ ] 摸底脚本重跑后:
  - `globalProvider` 引用数 = 0
  - `global` (authz) 引用数 = 0
  - `globalApp` 引用数 = 0
  - 4 个基础设施全局(`Pool` / `Client` / `cfg` / `defaultManager`) 保留,且不再被模块代码使用,只在 framework 内部被 `appCtx` 持有
- [ ] grep `authz.Get\(\)` / `extapi.Get\(\)` / `globalApp` 在 `server/apps/` 下命中 0 行
- [ ] `cms/module.go` 和 `flag/module.go` 用 `BaseModule` 写法
- [ ] 删除 `framework/pkg/extapi/provider.go` 里的 `globalProvider`(功能从 appCtx 拿)

**回滚策略**:同 Phase 3。

---

## 7. Phase 6 — 清理 ext_impl/registry.go 死代码 ⏳

**目标**:Phase 5 已经把 `provider.go` 改为从 `appCtx` 拿 Repository,所以 `ext_impl/registry.go` 里 189 行适配器已经彻底没用。

**改动清单**:
1. 删 `framework/internal/core/ext_impl/registry.go` 整个文件
2. `ext_impl/provider.go` 简化,删除 `userRecord` / `tenantRecord` / `userRepoAdapter` / `tenantPkgAdapter` / `pkgTenantGet`
3. `provider.go` 直接持有 `appCtx.UserRepo()` / `appCtx.TenantRepo()`

**验证标准**:
- [ ] `go build ./...` 退出码 0
- [ ] CMS 的 `GET /api/v1/cms/users` / `GET /api/v1/cms/tenant` 正常返回
- [ ] `wc -l framework/internal/core/ext_impl/registry.go` 输出 0 行(文件不存在)
- [ ] `wc -l framework/internal/core/ext_impl/provider.go` ≤ 40 行

---

## 8. Phase 7 — 清理 internal/middleware 重复 wrapper ⏳

**目标**:`internal/core/middleware/auth.go::Require` 等 5 个函数是 `pkg/middleware/auth.go` 的 thin wrapper,删掉。

**改动清单**:
1. 删 `framework/internal/core/middleware/auth.go` 里的 `Require` / `RequireAny` / `RequireAll` / `RequireAuthenticated` / `RequirePlatformRole`
2. 保留 `Auth` / `AuthLite` / `OptionalAuth` / `processAuthToken` / `injectAuthContext`(框架真正逻辑)
3. 检查 `framework/internal/core/middleware/` 下其他文件有没有 import 这些函数,有就改 import 路径到 `pkg/middleware`

**验证标准**:
- [ ] `go build ./...` 退出码 0
- [ ] 业务路由的 `middleware.Require(...)` 调用 100% 编译通过
- [ ] 实际请求 `GET /api/v1/users` 仍然触发权限校验(无 token 返回 401,无权限返回 403)
- [ ] `grep -n 'func Require' framework/internal/core/middleware/` 命中 0 行(只剩 pkg/middleware/auth.go)

---

## 9. Phase 8 — 补 P0 单测 ⏳

**目标**:为安全/核心路径加测试,防止回归。

**优先级与覆盖**:

| 测试 | 文件 | P0 原因 |
|---|---|---|
| `permission/types_test.go` | HasPermission 三种匹配 | RBAC 核心 |
| `permission/data_scope_test.go` | 5 种 DataScope → SQL | 数据隔离 |
| `auth/service_test.go` | 登录/注册/刷新 | 鉴权入口 |
| `plugin/plugin_test.go` | AppContext + Module 生命周期 | DI 容器正确性 |
| `resp/resp_test.go` | 状态码映射 | API 协议 |

**改动清单**:
1. 加 `framework/pkg/permission/types_test.go`(纯单测,无外部依赖)
2. 加 `framework/pkg/permission/data_scope_test.go`(可用 sqlmock 或 testcontainers)
3. 加 `apps/boot/auth/service_test.go`(mock Repository,测密码/Token/租户校验)
4. 加 `framework/pkg/plugin/plugin_test.go`(验证 Reader/Writer 分离)
5. 加 `framework/pkg/resp/resp_test.go`

**CI 门禁**:
- 配 `.github/workflows/test.yml`,跑 `go test ./...`
- 任何 phase 合并前必须绿

**验证标准**:
- [ ] 5 个测试文件全部存在
- [ ] `go test ./...` 退出码 0
- [ ] `go test -cover ./...` 总覆盖率 ≥ 60%(框架包 ≥ 80%,业务包 ≥ 50%)
- [ ] 测试运行时间 < 30s(纯单测 + 必要的 mock)

---

## 10. 总体验证(收官)

**收官前必须全部通过**:

| 编号 | 检查项 | 命令 | 通过标准 |
|---|---|---|---|
| T1 | 编译 | `go build ./...` | 退出 0 |
| T2 | 测试 | `go test ./...` | 退出 0 |
| T3 | 静态检查 | `go vet ./...` | 退出 0 |
| T4 | 全局变量回归 | 重跑 `phase0_scan.ps1` | 跨模块 = 0,基础设施 = 4 |
| T5 | 文档同步 | `doc/architecture.md` 更新 | 描述 AppContext 取代 registry |
| T6 | 灰度运行 | 用 test 账号连续 1 小时跑业务 | 无 5xx |

**最终量化指标**:

| 指标 | 重构前 | 重构后 |
|---|---:|---:|
| 全局变量 | 16 | 4 |
| 跨模块依赖 | 显式(全局) | 显式(AppContext) |
| 模块单测 | 0 个 | ≥ 5 个 |
| 编译错误时定位时间 | grep 全局 | IDE 跳转 |
| ext_impl 死代码 | 189 行 | 0 |

---

## 11. 风险登记表

| 风险 | 等级 | 缓解措施 |
|---|---|---|
| 改 Module 接口签名后,14 个 module 编译错误爆炸 | 高 | Phase 3-5 分批提交,每个 phase 收尾必须 `go build` 通过 |
| AppContext 字段太多变成"上帝对象" | 中 | Phase 2 已用 Reader/Writer 分离,定期 review 字段数 |
| ext_impl 删除后某个角落 panic | 中 | Phase 6 前 grep `provider.GetUserProvider` 全部改完 |
| 摸底脚本误报漏报 | 低 | Phase 0 数据 + 人工 review 双重确认 |
| 测试 CI 跑挂 | 中 | 先本地 `go test` 全绿再合 |
| 模块 init 顺序导致 ctx 读到 nil | 高 | Phase 3-5 严格按模块依赖顺序接入,init 内部不读其他模块的 slot,只在 Register 时读 |
| 用户本地的 stale go.sum | 低 | 文档提示 `go mod tidy` |

---

## 12. 文档同步清单

每个 phase 完成后,同步更新:

- [ ] `doc/architecture.md` — 在 Phase 2/3/5 节点后追加 AppContext 说明
- [ ] `doc/developing.md` — 改"如何新增模块"章节,用 `BaseModule` 而非 `NewModuleWithOpts`
- [ ] `doc/modules.md` — 更新模块清单,标注 Phase 6 之后删除的 module
- [ ] `CHANGELOG.md` (新建) — 每个 phase 合并时追加一行

---

## 13. 时间线(目标)

| 周 | 任务 |
|---|---|
| W1 | Phase 0-2 落地(已完成骨架)+ Phase 3 (boot/auth+tenant 接入) |
| W2 | Phase 4 (rbac) + Phase 5 (reference + system + cms/flag) |
| W3 | Phase 6 (ext_impl 清理) + Phase 7 (middleware 清理) + Phase 8 (P0 测试) + 收官 |
