---
title: "记一次 Repository 架构重构：让 apps 也能调用 framework 接口"
description: "XinFramework 开发过程中，如何让 apps 插件调用 framework 数据访问接口的架构设计记录。"
pubDate: 2026-04-25
---

# 记一次 Repository 架构重构：让 apps 也能调用 framework 接口

## 背景

今天继续完善 XinFramework 框架，期间遇到了一个架构设计问题，记录一下发现、思考、解决的全过程。

---

## 一、发现问题：Provider 会无限膨胀

写代码时发现，每新增一个 Model（如 Post、Category、Order），都要改三处地方：

```go
// 1. pkg/model/interfaces.go → 加 Entity + Repository 接口
// 2. internal/repository/ → 加 PostgreSQL 实现
// 3. pkg/repository/repository.go → Provider 加字段 + getter + 包级函数
```

这让我思考：**Provider 会无限膨胀，是否需要用泛型注册表方案？**

### 收益比分析

| 维度 | 当前方案（手动注册） | 泛型注册表 |
|------|---------------------|-----------|
| 新增 model 改动处 | 3 处 | 2 处 |
| 每次节省代码 | — | ~5 行 |
| 编译时类型安全 | ✅ 100% | ⚠️ 运行时 panic 风险 |
| IDE 补全 | ✅ 一目了然 | ❌ 不直观 |
| 调试难度 | ✅ 编译器报错 | ❌ 运行时 nil panic |

**结论**：在我的项目规模下（10-20 个 repo），手动注册的代价是每 1-2 个月写 5 行代码，而泛型注册表带来的隐式风险远超这 5 行的收益。

> Provider struct 的字段列表本身就是一份活文档，告诉所有人系统有哪些数据访问能力。

---

## 二、新问题：apps 无法调用 framework 的接口

写 CMS 插件时，发现 `apps/cms` 无法访问 `internal` 中的 User 模型功能。

这是 Go 语言的设计特性：**`internal` 目录是私有的，只能被同一个模块内的代码访问**。

```
framework/internal/     → 仅框架内部，外部无法访问
framework/pkg/         → 可被外部模块使用
apps/                  → 独立模块，无法 import internal
```

### 思考：apps 如何调用 User 数据？

当时想到几个方案：

| 方案 | 说明 | 评价 |
|------|------|------|
| A. apps 直接写 SQL | 简单直接 | ⚠️ 重复逻辑 |
| B. 暴露实现到 pkg | 把 internal/repository 复制到 pkg | ✅ 复用 |

但这里涉及一个核心问题：**DRY 原则**。

> 系统中，每一段知识、逻辑、功能，在系统内必须有且仅有唯一、权威的实现位置。

如果 `internal/repository` 和 `pkg/repository` 都写 SQL，就是违背 DRY。应该让 `pkg/repository` **复用** `internal/repository` 的实现。

---

## 三、解决：分层架构设计

最终设计如下：

```
framework/
├── pkg/
│   ├── model/
│   │   └── interfaces.go     # Repository 接口定义（稳定 API）
│   └── repository/
│       └── repository.go     # 暴露给 apps 的入口
└── internal/
    └── repository/           # 内部实现（SQL 逻辑只写一次）
        ├── user_repository.go
        ├── tenant_repository.go
        └── ...
```

### 架构分层职责

| 层级 | 位置 | 职责 |
|------|------|------|
| 接口定义 | `pkg/model/interfaces.go` | 供框架内部实现，也供外部参考 |
| 接口实现 | `internal/repository/` | 框架内部使用，复用 SQL 逻辑 |
| 业务模块 | `internal/module/` | 调用 Repository |
| 业务插件 | `apps/*` | 调用 `pkg/repository` |

### pkg/repository 调用 internal/repository

```go
// framework/pkg/repository/user.go
package repository

import "gx1727.com/xin/framework/internal/repository"

// User 返回 UserRepository 实例
func User() model.UserRepository {
    return internal_user.NewUserRepository()
}
```

这样：
- **SQL 逻辑只写一次**（在 internal/repository）
- **apps 可以调用**（通过 pkg/repository）
- **DRY 原则得以遵守**

---

## 四、额外收获：Repository 模式 vs 直接写 SQL

在探索过程中，也理清了 Repository 模式的一些疑问：

### 1. Handler 中直接写 SQL 好不好？

| 方式 | 适用场景 |
|------|---------|
| 直接写 SQL | 小项目、微服务、快速原型 |
| Repository 模式 | 中大型项目、需要测试、多数据源 |

**我的选择**：
- framework 核心代码：用 Repository 模式（长期维护）
- apps 业务代码：可以直接写 SQL（灵活快速）

### 2. Repository 模式有性能损耗吗？

**没有明显性能损耗。** Go 的接口调用是静态绑定，编译时已确定具体实现，不存在 Java 那样的虚函数表查找开销。

---

## 五、最终效果

现在 apps/cms 可以这样调用：

```go
import (
    "gx1727.com/xin/framework/pkg/model"
    "gx1727.com/xin/framework/pkg/repository"
)

func (h *Handler) GetUser(c *gin.Context) {
    ctx := c.Request.Context()

    user, err := repository.User().GetByID(ctx, userID)
    if err == model.ErrUserNotFound {
        // 处理未找到
    }
}
```

### 可用的 Repository 方法

| 调用 | 功能 |
|------|------|
| `repository.User().GetByID(ctx, id)` | 获取用户 |
| `repository.User().List(ctx, tenantID, keyword, page, size)` | 用户列表 |
| `repository.Tenant().GetByID(ctx, id)` | 获取租户 |
| `repository.Account().GetByID(ctx, id)` | 获取账号 |
| `repository.Role().GetUserRoles(ctx, userID)` | 用户角色 |

---

## 六、总结

今天的重构解决了一个关键问题：**如何让 apps 调用 framework 的数据访问接口**。

核心设计原则：

1. **DRY 优先**：SQL 逻辑只写一次
2. **分层明确**：internal 实现，pkg 暴露
3. **灵活适用**：framework 用 Repository，apps 可以直接 SQL
4. **不过度设计**：当前规模不需要泛型注册表

一个好的架构不是"一步到位"，而是**持续演进、解决实际问题**。
