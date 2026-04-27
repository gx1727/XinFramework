---
title: "XinFramework 权限系统设计笔记"
description: "记录 XinFramework 开发过程中关于 REST API 设计、HTTP 状态码、Token 机制和 RBAC 权限系统的深度讨论。"
pubDate: 2026-04-26
---

# XinFramework 权限系统设计笔记

今天继续推进 XinFramework 的开发，核心集中在**权限系统的表结构设计**和** API 设计规范**上。记录一些关键的设计决策。

---

## 一、REST API 方法选择

### 不要走"全 POST"的偷懒路线

在设计 CMS 相关接口时，讨论了是否只使用 GET + POST。结论是：**不要**。

标准 REST 语义：

| 方法 | 用途 |
|------|------|
| GET | 查询 |
| POST | 创建 |
| PUT | 全量更新 |
| PATCH | 部分更新 |
| DELETE | 删除 |

### 为什么不能全 POST？

1. **语义混乱** — `/user/update` 是改一个字段还是全部？
2. **缓存/网关能力废掉** — GET 可以天然缓存，全 POST 就没有了
3. **权限系统更难做** — 无法天然区分 read/write/delete 权限
4. **前端协作困难** — React Query / TanStack Query 默认按 REST 设计

---

## 二、HTTP 状态码设计

### 结论：HTTP 状态码 + 业务 Code 双层结构

| 类型 | HTTP 状态码 |
|------|------------|
| 成功 | 200 |
| 参数错误 | 400 |
| 未登录 | 401 |
| 无权限 | 403 |
| 不存在 | 404 |
| 服务器错误 | 500 |

业务 Code 表达具体错误（如 `1001` = 用户名已存在）。

"全 200" 设计只适合网关封装或历史包袱系统，不适合从零搭建的 SaaS。

---

## 三、Token 刷新机制

### access_token 过期时返回 401

| 错误码 | 含义 | 前端处理 |
|--------|------|----------|
| 40101 | access_token 过期 | 自动刷新 |
| 40102 | refresh_token 过期 | 强制登录 |
| 40103 | token 非法 | 安全警告 |

---

## 四、权限系统表结构评审

### 当前问题与优化

| 问题 | 建议 |
|------|------|
| tenant_users 表 | 删除（冗余设计） |
| permissions 表 | 改为 ID 模型（resource_id + action） |
| roles.scope_orgs (JSONB) | 改为结构化（role_data_scopes 表） |

### 核心原则

> **权限不是一张表解决的，而是三层叠加：**
> - 租户隔离（RLS）
> - 功能权限（RBAC）
> - 数据范围（Data Scope）

### 数据权限（Data Scope）

```go
type DataScope struct {
    Type   int     // 1全部 2自定义 3本部门 4本部门及以下 5本人
    OrgIDs []int64 // 自定义时使用
}
```

根据 DataScope 动态拼接 SQL：

```go
func BuildDataScopeSQL(u *UserContext) (string, []any) {
    switch u.DataScope.Type {
    case 1:  return "", nil                                    // 全部
    case 3:  return " AND org_id = $1", []any{u.OrgID}        // 本部门
    case 4:  return ltree_query, []any{u.OrgID}               // 本部门及以下
    case 5:  return " AND id = $1", []any{u.UserID}           // 仅本人
    case 2:  return " AND org_id = ANY($1)", []any{u.DataScope.OrgIDs}
    }
}
```

---

## 五、澄清：tenant_users 表

**结论**：不需要。

现有模型已经支持多角色：

```
accounts 1 --- N users
users N --- 1 tenant
users 1 --- N user_roles
user_roles N --- 1 roles
```

"一个用户多个角色" = `user_roles` 表，与租户无关。

---

## 总结

今天的核心收获：

1. **API 设计** — 遵循 REST 语义，不用偷懒的全 POST
2. **HTTP 状态码** — 正确使用 4xx/5xx，让网关和前端都能正确处理
3. **权限系统** — 三层叠加：RLS + RBAC + DataScope
4. **表结构** — 删除冗余，用 ID 关联替代字符串 code

下一步将把这些设计决策落地到代码中。
