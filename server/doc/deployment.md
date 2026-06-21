# 部署

> XinFramework 是单一二进制 + PostgreSQL + (可选 Redis) 的标准 SaaS 后端。
>
> 推荐部署：**systemd + nginx 反代 + PostgreSQL 主从 + Redis sentinel**

## 1. 编译

### 1.1 标准编译

仓库根目录：

```bash
./build.sh           # Linux/macOS
.\build.ps1          # Windows
```

脚本做的事：

1. `go mod tidy`（单 module `gx1727.com/xin`）
2. `go build -ldflags="-s -w"` 编译到 `bin/xin`（或 `bin/xin.exe`）
3. 复制 `config/` 和 `migrations/` 到产物目录（可选）

### 1.2 交叉编译

```bash
GOOS=windows GOARCH=amd64 ./build.sh
GOOS=darwin GOARCH=arm64 ./build.sh
```

### 1.3 减小二进制

```bash
go build -ldflags="-s -w" -o bin/xin ./cmd/xin
upx --best --lzma bin/xin
```

`-s -w` 去掉符号表和调试信息，通常能从 45MB 减到 30MB 左右。

---

## 2. 单机部署（开发 / 中小规模）

```
┌──────────────┐
│   nginx:443  │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ xin:8087     │ ← 单实例
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ PostgreSQL   │ ← 单实例
└──────────────┘
```

### 2.1 systemd unit

```ini
[Unit]
Description=XinFramework Server
After=network.target postgresql.service

[Service]
Type=simple
User=xin
WorkingDirectory=/opt/xin-server
ExecStart=/opt/xin-server/bin/xin run
Restart=on-failure
RestartSec=5s

Environment="XIN_APP_ENV=prod"
Environment="XIN_JWT_SECRET=CHANGEME-PROD-SECRET-MIN-32BYTES"

StandardOutput=append:/var/log/xin/xin.log
StandardError=append:/var/log/xin/xin.err

LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

启用：

```bash
sudo cp framework/xin-server.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now xin-server
sudo systemctl status xin-server
```

### 2.2 文件结构

```
/opt/xin-server/
├── bin/
│   └── xin
├── config/
│   ├── config.yaml
│   ├── config.prod.yaml
│   └── ...                   # 子模块 yaml
├── migrations/
│   ├── framework.sql
│   ├── asset.sql
│   ├── config.sql
│   ├── config_alignment.sql  # Phase 0022 新加
│   ├── dict.sql
│   ├── flag.sql
│   └── cms.sql
├── uploads/                  # local 存储
├── logs/                     # 自定义 logger 输出
├── xin.pid
└── xin.log
```

权限：`xin` 用户拥有 `/opt/xin-server` 和 `/var/log/xin`，不要 root 跑。

---

## 3. 高可用部署（生产推荐）

```
              ┌─────────────┐
┌────────┐ ┌─►│   nginx A   │
│ client │─┤  └──────┬──────┘
└────────┘ │         │
           │  ┌─────────────┐
           ├─►│ xin-server A │──┐
           │  └─────────────┘  │
           │  ┌─────────────┐  │
           └─►│ xin-server B │──┤
              └─────────────┘  │
                              ▼
                ┌──────────────────────┐
                │ PostgreSQL primary    │
                └─────────┬────────────┘
                          │
                ┌─────────▼────────────┐
                │ PostgreSQL replicas  │
                └──────────────────────┘

        ┌─────────────────────┐
        │ Redis sentinel (3)   │
        └─────────────────────┘
```

### 3.1 nginx 反代

```nginx
upstream xin_backend {
    server 10.0.0.1:8087 max_fails=3 fail_timeout=30s;
    server 10.0.0.2:8087 max_fails=3 fail_timeout=30s;
    keepalive 32;
}

server {
    listen 443 ssl http2;
    server_name api.example.com;

    ssl_certificate     /etc/letsencrypt/live/api.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.example.com/privkey.pem;

    client_max_body_size 100m;

    location / {
        proxy_pass http://xin_backend;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        proxy_read_timeout  300s;
        proxy_send_timeout  300s;
    }

    location /uploads/ {
        alias /opt/xin-server/uploads/;
        expires 7d;
    }
}
```

### 3.2 多实例无状态

`xin` 是无状态服务（状态都在 DB / Redis）：

- 可以开任意多个实例横扩
- `xin.pid` 文件每实例独立（`/var/run/xin-N.pid`）
- 日志按实例分开（`/var/log/xin/xin-N.log`）

### 3.3 滚动升级

```bash
scp bin/xin user@server:/opt/xin-server/bin/xin.new

ssh user@server
cd /opt/xin-server
sudo systemctl stop xin-server
sudo mv bin/xin bin/xin.bak
sudo mv bin/xin.new bin/xin
sudo systemctl start xin-server
sudo systemctl status xin-server
```

---

## 4. PostgreSQL 配置建议

### 4.1 postgresql.conf 关键参数

```ini
max_connections = 200
shared_buffers = 4GB
effective_cache_size = 12GB
work_mem = 64MB
maintenance_work_mem = 1GB

wal_level = replica
wal_log_hints = on
max_wal_size = 4GB
min_wal_size = 1GB

log_min_duration_statement = 500ms
log_lock_waits = on
log_temp_files = 0

timezone = 'UTC'
```

### 4.2 RLS 性能

RLS 会让每条 SQL 多一次 policy 评估：

- `tenant_id` 列建索引（几乎所有租户级查询都需要）
- RLS policy 表达式尽量简单
- 大量导入用 `SET LOCAL role = bypass_rls_role` 绕过

### 4.3 主从 + 读写分离（可选）

当前框架**所有查询走主库**。如果未来要读写分离：

1. 用 `pgxpool` 的两个 pool：`writerPool` + `readerPool`
2. `AppContext.Reader` 暴露两个 pool，handler 选哪个
3. 框架层暂时不做这件事，留给业务层

---

## 5. Redis 配置

### 5.1 单实例

```yaml
redis:
  enabled: true
  required: true            # 生产建议 required
  host: redis.internal
  port: 6379
  pool_size: 50
  min_idle_conns: 10
```

### 5.2 Sentinel

需要在代码里改 `cache.Init` 接 sentinel 地址。当前 `framework/pkg/cache/cache.go` 只支持单实例：

```go
redis.NewFailoverClient(&redis.FailoverOptions{
    MasterName:    "mymaster",
    SentinelAddrs: []string{":26379", ":26380", ":26381"},
    Password:      cfg.Redis.Password,
    DB:            cfg.Redis.DB,
})
```

### 5.3 Cluster

Cluster 模式同样需要扩展 `cache.Init`，引入 hash tag 让同一用户的 key 落在同一 slot。

---

## 6. 监控

### 6.1 应用层 metrics

- `GET /system/server-info`（已有）：进程级 + DB + Redis 状态
- 加 Prometheus 端点：`/metrics`

### 6.2 关键监控项

| 指标 | 阈值 | 来源 |
|---|---|---|
| 请求 P99 延迟 | < 500ms | nginx access log |
| 错误率 | < 0.1% | gin middleware |
| Goroutines | < 1000 | `runtime.NumGoroutine` |
| DB 连接池空闲率 | > 20% | `pg_stat_activity` |
| Redis 命中率 | > 80% | `INFO stats` |
| 磁盘空间 | < 80% | OS |

### 6.3 日志

`logs/` 目录按天滚动（框架的 logger.Init 配 `cfg.Log.Dir` / `cfg.Log.Level`）。

```go
log.Printf("module %s initialized", m.Name())
logger.Errorf("[%s] %s %s | %d | %s", reqID, method, path, code, msg)
```

nginx access log 也保留（trace 完整链路）。

### 6.4 Trace ID

`middleware.RequestID()` 给每个请求注入 `X-Request-ID`（Header 或自动生成），写入 gin context。

- 框架 logger 自带
- 传给下游（DB log, Redis log）的 `tx.ctx` 也带
- 前端出问题拿 `X-Request-ID` 来 trace

---

## 7. 备份策略

### 7.1 PostgreSQL

```bash
pg_dump -Fc -h db.host -U xin_user xin > backup_$(date +%F).dump
pg_restore -d xin backup_2026-06-21.dump
```

| 备份类型 | 频率 | 保留 |
|---|---|---|
| 全量 | 每天 03:00 | 7 天 |
| WAL 归档 | 持续 | 30 天 |
| 异地副本 | 实时 | 永久 |

### 7.2 Redis

- 开启 AOF（`appendonly yes`）
- `appendfsync everysec`
- 主从复制 + Sentinel 自动 failover

业务上 Redis 是缓存，丢了不致命（数据可重建），但 session 丢失会强制用户重新登录。

### 7.3 uploads/ 本地文件

```bash
rsync -av /opt/xin-server/uploads/ backup:/path/xin-uploads/
```

或者改用 COS（腾讯云对象存储），由 COS 自己保证可用性。

---

## 8. Docker 镜像（可选）

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY . .
RUN cd server && go build -ldflags="-s -w" -o /out/xin ./cmd/xin

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=builder /out/xin /app/xin
COPY server/config /app/config
COPY server/migrations /app/migrations
EXPOSE 8087
ENTRYPOINT ["/app/xin", "run"]
```

```bash
docker build -t xin:1.0.0 .
docker run -d \
  --name xin \
  -p 8087:8087 \
  -e XIN_DB_HOST=postgres \
  -e XIN_DB_PASSWORD=secret \
  -e XIN_JWT_SECRET=$(openssl rand -base64 48) \
  -v /opt/xin-uploads:/app/uploads \
  xin:1.0.0
```

---

## 9. Kubernetes（高级）

StatefulSet + Headless Service：

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: xin
spec:
  replicas: 3
  serviceName: xin
  template:
    spec:
      containers:
      - name: xin
        image: xin:1.0.0
        env:
        - name: XIN_DB_HOST
          value: postgres.default.svc
        - name: XIN_REDIS_HOST
          value: redis.default.svc
        ports:
        - containerPort: 8087
        livenessProbe:
          httpGet: { path: /api/v1/health, port: 8087 }
          initialDelaySeconds: 30
        readinessProbe:
          httpGet: { path: /api/v1/health, port: 8087 }
          initialDelaySeconds: 5
```

注意事项：

- StatefulSet 给每个实例**固定 hostname**（`xin-0`, `xin-1`, `xin-2`），日志 / pid 隔离方便
- 用 `PodDisruptionBudget` 防止滚动升级全部 kill
- `sessionAffinity: ClientIP` 让同一 IP 落到同一实例

---

## 10. 环境变量参考

| 变量 | 对应 YAML | 说明 |
|---|---|---|
| `XIN_APP_PORT` | `app.port` | HTTP 端口（默认 8087） |
| `XIN_DB_HOST` | `database.host` | PG host |
| `XIN_DB_PASSWORD` | `database.password` | PG password |
| `XIN_REDIS_ENABLED` | `redis.enabled` | Redis 是否启用 |
| `XIN_REDIS_REQUIRED` | `redis.required` | Redis 挂掉是否启动失败 |
| `XIN_JWT_SECRET` | `jwt.secret` | JWT 签名 key，**prod 必填 ≥32 字节** |
| `XIN_CORS_ALLOW_ORIGINS` | `cors.allow_origins` | CORS 白名单（逗号分隔） |
| `XIN_MODULE` | `module` | 模块白名单（逗号分隔） |
| `XIN_STORAGE_COS_SECRET_ID` | `storage.cos_secret_id` | COS 凭据 |
| `XIN_STORAGE_COS_SECRET_KEY` | `storage.cos_secret_key` | COS 凭据 |

---

## 11. 健康检查清单

```bash
ps aux | grep xin
ss -tlnp | grep 8087           # Linux
Get-NetTCPConnection -LocalPort 8087  # Windows

curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8087/api/v1/health

psql -h db -U xin_user -d xin -c '\dt'

psql -h db -U xin_user -d xin -c "SELECT * FROM tenants WHERE code='bootstrap';"

redis-cli -h redis ping

curl -X POST http://localhost:8087/api/v1/auth/tenant-login \
  -H 'Content-Type: application/json' \
  -d '{"account":"admin","password":"...","tenant_code":"bootstrap"}'
```