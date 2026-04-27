---
title: "XinFramework 微信小程序登录实现"
description: "记录 XinFramework 开发过程中微信小程序登录模块的设计与实现过程，包括 SDK 选型、接口定义、数据库扩展。"
pubDate: 2026-04-27
---

# XinFramework 微信小程序登录实现

今天在 XinFramework 中完成了**微信小程序登录模块**的开发。核心工作集中在：SDK 选型分析、直接调 API 实现方案、数据库字段扩展。

---

## 一、SDK 选型：自己实现 vs 用 SDK

在正式动手之前，先梳理了微信小程序开发的几种方案：

| SDK | 地址 | 特点 |
|-----|------|------|
| wechat-go | github.com/skiy/wechat | 简单轻量 |
| go-wechat | github.com/chanxingshun/go-wechat | 功能较全 |
| wechat SDK | github.com/solstice/wechat | 支持微信全系 |
| 官方无 SDK | 直接调 API | 最灵活，参考微信开发文档 |

### 为什么最终选择自己实现

微信小程序登录本质上只需要调用几个核心 API，不需要重型 SDK：

```
wx.login() → 获取 code
→ 传到后端 → https://api.weixin.qq.com/sns/jscode2session?js_code=xxx&...
→ 返回 openid/session_key
→ 生成自定义登录态
```

自己实现的优势：
- **代码轻量**，无额外依赖
- **更符合 XinFramework 架构风格** — 框架坚持手写 SQL、不用 ORM，微信模块也保持一致
- **GitHub 访问不稳定**时不受影响

### 不同微信能力的复杂度

| 微信能力 | 复杂度 | 自己实现的坑 |
|----------|--------|------------|
| 登录(code2session) | 低 | 简单 |
| 获取手机号 | 低 | 需解密 data |
| 小程序码生成 | 中 | 需 access_token 缓存 |
| 消息订阅 | 低 | 参数多 |
| 内容安全检测 | 中 | 需鉴权 |
| 微信支付 | 高 | 签名/回调/证书一堆 |
| 附近的小程序LBS | 中 | 坐标转换 |

**推荐方案：**
- 轻度使用（登录 + 手机号）：自己实现
- 深度使用（支付、订阅消息、内容安全等）：用 SDK

---

## 二、微信模块实现

### 配置结构

在 `config.yaml` 中新增了 wxxcx 配置：

```yaml
wxxcx:
  appid: wx3e18aa16ef3c9ea9
  appsecret: 87cbb533602cde12e30094df147783fb
```

对应的配置加载：

```go
type WxxcxConfig struct {
    AppID     string `yaml:"appid"`
    AppSecret string `yaml:"appsecret"`
}
```

### 核心接口实现

微信小程序登录核心就 3 个接口：

```go
// 1. code2session
GET https://api.weixin.qq.com/sns/jscode2session?appid=APPID&secret=***&js_code=JSCODE&grant_type=authorization_code

// 2. 获取手机号
POST https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token=***

// 3. 内容安全检测
POST https://api.weixin.qq.com/wxa/img_sec_check?access_token=***
```

### 模块文件结构

在 `framework/internal/module/weixin/` 下实现了完整模块：

```
weixin/
├── types.go      # 请求/响应结构体
├── errors.go     # 错误定义
├── service.go    # 微信 API 调用
├── handler.go    # HTTP handlers
├── routes.go     # 路由注册
└── module.go     # 模块定义
```

---

## 三、数据库扩展

### 新增字段

为 `accounts` 表新增微信字段：

```sql
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS wechat_openid VARCHAR(64);
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS wechat_unionid VARCHAR(64);
```

对应 model 变更：

```go
type Account struct {
    ID            uint      `json:"id"`
    Username      string    `json:"username"`
    Phone         string    `json:"phone"`
    Email         string    `json:"email"`
    RealName      string    `json:"real_name"`
    Status        int8      `json:"status"`
    WechatOpenID  string    `json:"wechat_openid"`
    WechatUnionID string    `json:"wechat_unionid"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
}
```

### Repository 新增方法

```go
// 按 OpenID 查询账号
GetByOpenID(ctx context.Context, openID string) (*Account, error)

// 创建微信账号
CreateWeChatAccount(ctx context.Context, openID, unionID, phone string) (uint, error)

// 更新手机号
UpdatePhone(ctx context.Context, userID uint, phone string) error
```

---

## 四、Migrations 整理

今天还完成了一项清理工作：将 `003_framework_create_attachments.sql` 合并到 `001_framework_init.sql` 中。

### 变更内容

在 `001_framework_init.sql` 中新增：

1. **attachments 表**（第 22 个表）
   - 包含字段：id, tenant_id, user_id, file_name, file_ext, mime_type, file_size, storage, object_key, url, hash, status, created_at, updated_at, is_deleted

2. **索引**
   - `idx_attachments_tenant_hash` — 去重索引
   - `idx_attachments_tenant` — 租户查询索引

3. **RLS 策略**
   ```sql
   ALTER TABLE attachments ENABLE ROW LEVEL SECURITY;
   CREATE POLICY tenant_isolation_policy ON attachments ...;
   ```

已删除独立的 003 迁移文件，构建通过。

---

## 总结

今天的核心收获：

1. **微信登录** — 直接调 API 最轻量，符合 XinFramework 风格
2. **模块开发** — types → errors → service → handler → routes → module 标准化流程
3. **数据库设计** — 字段扩展 + Repository 方法 + Model 变更三步走
4. **Migrations** — 及时合并，减少迁移文件数量

下一步将完善微信手机号获取和解密功能。
