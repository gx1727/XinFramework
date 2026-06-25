# 多实例部署指南

> 本文描述 XinFramework 在多实例部署场景下的配置、Redis 角色、
> 权限缓存生效路径、灰度滚动策略与监控指标。
> 单实例部署请忽略本文，按 [deployment.md](./deployment.md) 即可。

---

## 1. 为什么需要多实例

单实例部署在以下场景会出现瓶颈：

- 单机 CPU/内存打满，无法通过纵向扩容解决
- 进程崩溃即服务中断，可用性不足
- 部署新版本必须停机

多实例 + 负载均衡是标准应对方案。但多实例会引入新的分布式问题——
**Redis 成为状态共享的关键组件**。

---

## 2. 必备组件

| 组件 | 要求 | 说明 |
|---|---|---|
| **PostgreSQL** | 单主或主从，RPO≈0 | 所有实例共享同一份业务数据 |
| **Redis** | **必须共享**（Sentinel / Cluster / 单机均可） | 存储 Session + 权限缓存 |
| **负载均衡器** | Nginx / HAProxy / 云 LB | 7 层或 4 层均可 |
| **静态资源** | CDN / OSS | asset 上传走 object storage，不依赖实例 |

**最小部署架构：**

```
                ┌────────────────┐
                │ Load Balancer  │
                └────────┬───────┘
                         │
        ┌────────────────┼────────────────┐
        ▼                ▼                ▼
   ┌─────────┐      ┌─────────┐      ┌─────────┐
   │ xin #1  │      │ xin #2  │      │ xin #N  │
   └────┬────┘      └────┬────┘      └────┬────┘
        │                │                │
        └────────────────┼────────────────┘
                         │
                  ┌──────┴──────┐
                  ▼             ▼
            ┌─────────┐   ┌─────────┐
            │   PG    │   │  Redis  │
            └─────────┘   └─────────┘
```

---

## 3. 配置

### 3.1 所有实例共用一份配置

`config/config.prod.yaml`：

```yaml
app:
  name: xin
  env: prod
  host: 0.0.0.0
  port: 8087

database:
  host: pg.prod.internal
  port: 5432
  user: xin
  password: ${XIN_DB_PASSWORD}    # env 注入
  dbname: xin
  sslmode: require
  max_open_conns: 100             # 每实例；N 实例合计 N×100
  max_idle_conns: 20

redis:
  host: redis.prod.internal
  port: 6379
  password: ${XIN_REDIS_PASSWORD}
  db: 0
  enabled: true
  required: true                   # prod 强制要求 Redis 可用
  pool_size: 20                    # 每实例；N 实例合计 N×20
  min_idle_conns: 5

permission_cache:
  perm_ttl_seconds: 900            # 15 分钟
  data_scope_ttl_seconds: 1800     # 30 分钟
  key_prefix: "xin:user:"          # 多服务共享 Redis 时建议带前缀

jwt:
  secret: ${XIN_JWT_SECRET}       # 必填，≥32 字节，校验在 config.Load() 阶段
  expire: 3600
  refresh_expire: 86400

cors:
  enabled: true
  allow_origins:
    - https://app.example.com
  allow_credentials: true
```

### 3.2 实例差异化配置（如果需要）

通过环境变量覆盖：

```bash
# 实例 A：监听 8087
XIN_APP_PORT=8087 ./xin run

# 实例 B：监听 8088
XIN_APP_PORT=8088 ./xin run
```

PostgreSQL / Redis 连接池的 max_open_conns 是**每实例**值，N 实例
同时跑就是 N×100 个连接。规划容量时务必算上。

---

## 4. 权限缓存的生效路径

### 4.1 缓存选型规则

boot.Init 启动期自动选择：

```
cfg.redis.enabled=true + cache.Get() != nil
   → RedisPermissionCache（多实例共享，所有实例看到一致权限）

cfg.redis.enabled=false 或 cache.Init Ping 失败且 required=false
   → MemoryPermissionCache（仅本进程，多实例不一致）
```

### 4.2 Redis 不可用的代价

如果某实例启动时 Redis 不可用：

- 该实例的所有权限 / 数据范围查询会直接走 PG（无缓存）
- Session 也会自动 fallback 到 DB SessionManager
- 用户能登录能用，但每次请求都会查 DB，性能下降

**生产环境务必**：`redis.required=true`，启动期 Redis 不可达时进程直接退出，
避免无声降级。

### 4.3 权限修改的传播

管理员在实例 A 上修改角色权限：

```
实例 A：业务事务 COMMIT
       ↓
       authz.InvalidateRole(roleID)
       ↓
       删 Redis key：xin:user:perm:* （该 role 的所有 user）
       ↓
实例 B / C：下次 LoadPermissions(role.user) miss
       ↓
       重新查 DB → 写入 Redis
```

**TTL 期内的最长不一致窗口** = `perm_ttl_seconds`（默认 900s）。
**主动失效下** = 0（删除 Redis key 后所有实例立刻 miss）。

---

## 5. Session 共享

SessionManager 通过 cfg.redis.enabled 自动选型：

- Redis 可用 → RedisSessionManager（所有实例共享）
- Redis 不可用 → DBSessionManager（多实例通过 PG 共享）

DBSessionManager 把 session 写入 `auth_sessions` 表，所有实例都查同一份。
性能比 Redis 略差（多一次 DB IO），但保证了一致性。

---

## 6. 灰度滚动

### 6.1 滚动策略

```
[实例 1] 部署新版本  →  验证健康
         ↓
[实例 2] 部署新版本  →  验证健康
         ↓
...
[实例 N]
```

**关键不变量**：同一时刻新旧版本共存，行为应兼容。

### 6.2 兼容性检查清单

- [ ] 新增权限常量（`Res*` / `Act*`）向后兼容（旧代码引用旧常量仍能编译）
- [ ] 修改的错误码段位不在已发布区间
- [ ] DB migration 加列用 `ADD COLUMN IF NOT EXISTS`，不删旧列
- [ ] 新接口只在 RouterGroup 内追加，不删除既有
- [ ] JWT Claims 字段只在末尾加（保持顺序）

### 6.3 回滚

任意实例健康异常时，立即通过 LB 摘除该实例；旧实例仍可服务新实例流量。
最坏情况：整批回滚到上一版本镜像。

---

## 7. 配置建议

### 7.1 实例数量

| 业务规模 | 实例数 | Redis 部署 |
|---|---|---|
| QPS < 100、用户 < 1k | 1 | 单机 |
| QPS 100-1k、用户 1k-10k | 2-3 | 单机 |
| QPS 1k-10k、用户 10k-100k | 3-10 | Sentinel 1 主 1 从 |
| QPS > 10k、用户 > 100k | 10+ | Cluster（≥ 3 主 3 从） |

### 7.2 连接池容量

N 实例 × max_open_conns（PG） ≤ PG `max_connections`（默认 100）。
N 实例 × redis.pool_size ≤ Redis `maxclients`（默认 10000，足够）。

### 7.3 缓存 TTL

| 数据 | 推荐 TTL | 理由 |
|---|---|---|
| 权限码 | 15 分钟（默认） | 改权限概率低，TTL 影响小 |
| 数据范围 | 30 分钟（默认） | 改 data_scope 极低频 |
| Session | 与 JWT expire 一致 | 提前过期会要求重新登录 |

如果对"权限立即生效"要求高，主动调 `InvalidateUser` 是唯一可靠路径。

---

## 8. 监控指标

### 8.1 必须监控

| 指标 | 采集方式 | 告警阈值 |
|---|---|---|
| PostgreSQL 连接数 | `pg_stat_activity` | > 80% max_connections |
| Redis 内存使用 | `INFO memory` | > 70% maxmemory |
| Redis 连接数 | `INFO clients` | > 80% maxclients |
| 进程 goroutine 数 | `runtime.NumGoroutine` | > 10k |
| 进程 RSS | `/proc/$pid/status` | > 1GB |

### 8.2 推荐监控

| 指标 | 采集方式 | 说明 |
|---|---|---|
| 权限缓存命中率 | Prometheus counter `perm_cache_hit_total / (hit+miss)` | < 80% 说明 TTL 过短或权限变更频繁 |
| Session 失效次数 | `INFO keyspace` 命中率 | 突然下降 = 大量强制退出 |
| `RunInTx` 嵌套层级 | 自定义 metric | > 5 提示代码有循环 |
| RLS bypass 次数 | `pg_stat_statements` 过滤 `bypass_rls=on` | 突然增加 = 平台域操作增多 |

### 8.3 日志

每个实例写本地日志 + （可选）聚合到 Loki / ELK。
搜索 `[migrate]` / `[audit]` / `[weixin]` 前缀可过滤模块日志。

---

## 9. 故障场景

### 9.1 Redis 故障

- 所有实例自动 fallback 到 DBSessionManager
- 权限缓存失效，所有请求查 DB
- 性能下降，建议立即恢复 Redis

### 9.2 PostgreSQL 故障

- 所有实例启动期会失败（`Ping` 报错）
- 启动中的实例：业务请求 500（事务回滚）
- 立即切换到 PG 主从切换 / 提升从库

### 9.3 单一实例 OOM

- LB 通过健康检查摘除该实例
- 剩余实例继续服务
- 通过日志分析 OOM 原因（通常是大 SQL / goroutine 泄漏）

### 9.4 旧实例不能解析新权限常量

灰度滚动期间，旧实例可能不识别新增 `Res*` 常量。
**应对**：所有 `Res*` 常量在新版本发版前先 merge 到所有版本（双发），
再启用新逻辑。

---

## 10. 快速验证清单

部署完成后，按以下顺序验证：

1. [ ] 所有实例启动日志 `[INFO] module XXX initialized` 完整
2. [ ] LB 健康检查全绿（`/api/v1/health`）
3. [ ] 任一实例登录后，其他实例能访问受保护资源（Session 共享）
4. [ ] 实例 A 修改某用户角色，实例 B 上同一用户下一次请求立即生效
5. [ ] Redis 停掉，所有实例仍能服务（fallback 到 DB）
6. [ ] 单实例重启后，登录态不丢（其他实例 Session 仍可用）

完成以上 6 步方可上线。