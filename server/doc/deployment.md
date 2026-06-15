# 部署

> 编译、打包、systemd、监控。

## 1. 构建

### Linux / macOS

```bash
cd server
./build.sh
# 产物：bin/xin
```

或手工：

```bash
cd server
CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/xin ./cmd/xin
```

参数说明：

- `CGO_ENABLED=0`：纯静态二进制，方便 alpine / scratch 镜像
- `-ldflags="-s -w"`：去掉符号表与调试信息，体积减小约 30%

### Windows

```powershell
cd server
.\build.ps1
# 产物：bin\xin.exe
```

### 交叉编译

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/xin-linux-amd64 ./cmd/xin

# Linux ARM64
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/xin-linux-arm64 ./cmd/xin

# macOS ARM64
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/xin-darwin-arm64 ./cmd/xin
```

## 2. 运行模式

### 2.1 前台（开发）

```bash
./bin/xin run
# Ctrl+C 退出
```

### 2.2 守护进程（生产）

```bash
./bin/xin start
# fork + 写 pid 到 ./xin.pid
# 日志输出到 ./xin.log
```

```bash
./bin/xin stop     # 发 SIGTERM，graceful shutdown
./bin/xin restart  # stop + start
./bin/xin status   # 查看 pid 与运行状态
./bin/xin reload   # 平滑重载配置（未来支持）
```

### 2.3 systemd（推荐生产）

`framework/xin-server.service`：

```ini
[Unit]
Description=XinFramework Server
After=network.target postgresql.service

[Service]
Type=simple
User=xin
Group=xin
WorkingDirectory=/opt/xin
ExecStart=/opt/xin/bin/xin run
Restart=always
RestartSec=5
LimitNOFILE=65536

EnvironmentFile=/opt/xin/config/xin.env

[Install]
WantedBy=multi-user.target
```

安装：

```bash
sudo cp framework/xin-server.service /etc/systemd/system/
sudo useradd -r -s /sbin/nologin xin
sudo mkdir -p /opt/xin/{bin,config,uploads,logs}
sudo cp bin/xin /opt/xin/bin/
sudo cp config/config.prod.yaml /opt/xin/config/config.yaml

# 环境变量（密码等敏感信息放这里，不要进 YAML）
sudo tee /etc/xin/xin.env <<EOF
DB_PASSWORD=xxx
JWT_SECRET=xxx
EOF

sudo systemctl daemon-reload
sudo systemctl enable --now xin-server
sudo systemctl status xin-server
```

## 3. Docker

### Dockerfile

```dockerfile
# server/Dockerfile
FROM golang:1.25 AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/xin ./cmd/xin

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /out/xin /app/xin
COPY config /app/config
COPY migrations /app/migrations
EXPOSE 8080
CMD ["/app/xin", "run"]
```

构建运行：

```bash
docker build -t xin-framework:latest .
docker run --rm -p 8080:8080 \
  -e DB_HOST=host.docker.internal \
  -v $(pwd)/config:/app/config:ro \
  xin-framework:latest
```

### docker-compose.yml

```yaml
version: "3.8"
services:
  postgres:
    image: postgres:16
    environment:
      POSTGRES_USER: xin
      POSTGRES_PASSWORD: dev
      POSTGRES_DB: xin
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data

  xin:
    build: .
    depends_on:
      - postgres
    ports:
      - "8080:8080"
    environment:
      DB_HOST: postgres
      DB_USER: xin
      DB_PASSWORD: dev
      DB_NAME: xin
      JWT_SECRET: change-me
    volumes:
      - ./uploads:/app/uploads
    restart: unless-stopped

volumes:
  pgdata:
```

## 4. 反向代理

### Nginx

```nginx
server {
    listen 80;
    server_name api.example.com;

    client_max_body_size 20M;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 60s;
    }

    location /files/ {
        # 静态文件由 nginx 直接服务，不走 Go
        alias /opt/xin/uploads/;
        expires 7d;
    }
}
```

### Caddy

```
api.example.com {
    reverse_proxy 127.0.0.1:8080
}
```

## 5. HTTPS

用 Caddy 或 nginx + Let's Encrypt：

```bash
# Caddy 自动 HTTPS
sudo caddy reverse-proxy --from api.example.com --to 127.0.0.1:8080

# 或 nginx + certbot
sudo certbot --nginx -d api.example.com
```

Go 端不需要任何改动，证书由前端代理处理。

## 6. 配置热加载

当前版本：**不支持**。改 `config.yaml` 后需要 `xin restart`。

未来计划：用 fsnotify 监听配置文件变化 + 重新加载。

## 7. 监控

### 7.1 健康检查

```bash
curl http://localhost:8080/api/v1/system/health
```

返回：

```json
{
  "code": 0,
  "data": {
    "db": "ok",
    "cache": "ok",
    "uptime_seconds": 3600
  }
}
```

### 7.2 Prometheus（计划中）

未来通过 `/metrics` 暴露 Prometheus 指标。

当前可用：自己打点到日志，配合 Loki / ELK 分析。

### 7.3 审计日志

`framework/pkg/audit` 写入 `audit_logs` 表：

```sql
SELECT user_id, action, target, created_at
FROM audit_logs
WHERE created_at > NOW() - INTERVAL '1 day'
ORDER BY created_at DESC;
```

## 8. 数据库备份

```bash
#!/bin/bash
# /opt/xin/scripts/backup.sh
set -e
BACKUP_DIR=/var/backups/xin
DATE=$(date +%Y%m%d_%H%M%S)
mkdir -p $BACKUP_DIR
pg_dump -U xin -h localhost -Fc xin > $BACKUP_DIR/xin_$DATE.dump
# 保留最近 30 天
find $BACKUP_DIR -name "xin_*.dump" -mtime +30 -delete
```

加到 crontab：

```cron
0 3 * * * /opt/xin/scripts/backup.sh
```

恢复：

```bash
pg_restore -U xin -h localhost -d xin_new /var/backups/xin/xin_20260615_030000.dump
```

## 9. 升级流程

```bash
# 1. 备份数据库
./scripts/backup.sh

# 2. 拉新代码
cd /opt/xin
git pull

# 3. 重新构建
./build.sh

# 4. 跑迁移（启动时会自动跑）
# 也可以手动验证：
psql -U xin -d xin -f migrations/framework.sql

# 5. 重启
sudo systemctl restart xin-server

# 6. 验证
curl http://localhost:8080/api/v1/system/health
```

## 10. 常见故障

### 10.1 启动失败 "address already in use"

端口被占用。找到进程：

```bash
sudo lsof -i :8080
# 或
sudo ss -tlnp | grep 8080
```

杀掉：

```bash
sudo kill -9 <pid>
```

### 10.2 数据库连接耗尽

`db.max_conns: 20` 不够。调大 + 检查 connection leak：

```sql
SELECT * FROM pg_stat_activity WHERE state = 'idle';
```

### 10.3 内存泄漏

pprof：

```go
import _ "net/http/pprof"

go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

```bash
go tool pprof http://localhost:6060/debug/pprof/heap
```

### 10.4 上传 413

nginx 默认 `client_max_body_size 1m`。调到 20M+：

```nginx
client_max_body_size 20M;
```

### 10.5 JWT 过期 / 时区错误

确保服务器时区是 UTC 或正确的本地时区：

```bash
timedatectl
# 设置
sudo timedatectl set-timezone Asia/Shanghai
```

JWT 过期时间用 `time.Now()` 比较，要求服务器时间准确。