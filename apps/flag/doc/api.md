# Flag App API 文档

> 头像框生成器 / 活动头像工具 API

## 基础信息

- **Base URL**: `/api/v1/flag`
- **认证方式**: JWT Bearer Token
- **返回格式**: JSON

### 通用响应结构

```json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| code | int | 状态码，0 表示成功 |
| msg | string | 消息 |
| data | object | 数据 |

---

## 头像框 (Frames)

### 获取头像框列表

获取所有头像框模板

**GET** `/flag/frames`

**Query Parameters:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| category_id | uint | 否 | 分类ID |
| page | int | 否 | 页码，默认 1 |
| size | int | 否 | 每页数量，默认 20 |

**响应示例:**

```json
{
  "code": 0,
  "msg": "ok",
  "data": [
    {
      "id": 1,
      "tenant_id": 0,
      "category_id": 1,
      "name": "程序员专属",
      "description": "适合程序员的简约头像框",
      "preview_url": "/frames/preview/programmer.png",
      "template_url": "/frames/template/programmer.png",
      "template_config": {
        "avatar_x": 100,
        "avatar_y": 100,
        "avatar_width": 300,
        "avatar_height": 300
      },
      "type": "public",
      "sort": 1,
      "status": 1
    }
  ]
}
```

---

### 获取单个头像框

**GET** `/flag/frames/:id`

**Path Parameters:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint | 是 | 头像框ID |

**响应示例:**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1,
    "name": "程序员专属",
    "template_url": "/frames/template/programmer.png"
  }
}
```

---

## 头像框分类 (Categories)

### 获取头像框分类列表

**GET** `/flag/categories`

**响应示例:**

```json
{
  "code": 0,
  "msg": "ok",
  "data": [
    {
      "id": 1,
      "tenant_id": 0,
      "code": "emotion",
      "name": "情绪类",
      "type": "emotion",
      "sort": 1,
      "status": 1
    },
    {
      "id": 2,
      "tenant_id": 0,
      "code": "school",
      "name": "学校活动",
      "type": "custom",
      "sort": 2,
      "status": 1
    }
  ]
}
```

---

## 活动空间 (Spaces)

### 获取活动 Space (通过邀请码)

**GET** `/flag/spaces/:code`

**Path Parameters:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| code | string | 是 | 邀请码 |

**响应示例:**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1,
    "tenant_id": 1,
    "name": "测试活动",
    "description": "这是一个测试活动",
    "frame_id": 1,
    "space_config": {
      "fields": [
        {"key": "grade", "label": "届数", "required": true, "show": true}
      ]
    },
    "access_type": "public",
    "invite_code": "test",
    "max_usage": 100,
    "usage_count": 10,
    "status": 1
  }
}
```

---

### 创建活动 Space

**POST** `/flag/spaces`

**Headers:**
```
Authorization: Bearer <token>
```

**Request Body:**

```json
{
  "name": "校庆100周年活动",
  "description": "庆祝学校建校100周年",
  "frame_id": 1,
  "access_type": "public",
  "start_at": "2026-05-01T00:00:00Z",
  "end_at": "2026-05-31T23:59:59Z"
}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 活动名称 |
| description | string | 否 | 活动描述 |
| frame_id | uint | 否 | 绑定的头像框ID |
| access_type | string | 否 | 访问类型: public/invite/limit |
| start_at | string | 否 | 开始时间 |
| end_at | string | 否 | 结束时间 |

**响应示例:**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1,
    "name": "校庆100周年活动",
    "invite_code": "abc12345"
  }
}
```

---

### 更新活动 Space

**PUT** `/flag/spaces/:id`

**Headers:**
```
Authorization: Bearer <token>
```

**Request Body:**

```json
{
  "id": 1,
  "name": "更新后的名称",
  "description": "更新后的描述",
  "frame_id": 2,
  "status": 1
}
```

---

### 删除活动 Space

**DELETE** `/flag/spaces/:id`

**Headers:**
```
Authorization: Bearer <token>
```

---

### 获取我的活动 Space 列表

**GET** `/flag/spaces`

**Headers:**
```
Authorization: Bearer <token>
```

---

## 头像生成

### 生成头像

**POST** `/flag/generate`

**Headers:**
```
Authorization: Bearer <token>
```

**Request Body:**

```json
{
  "frame_id": 1,
  "space_id": 1,
  "source_image": "https://example.com/my-photo.jpg",
  "field_values": {
    "grade": "2024届",
    "college": "计算机学院"
  }
}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| frame_id | uint | 是 | 头像框ID |
| space_id | uint | 否 | 关联的Space ID |
| source_image | string | 是 | 用户上传的原图URL |
| field_values | object | 否 | 动态字段值 |

**响应示例:**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "id": 1,
    "result_url": "https://img.gx1727.com/flag/1/abc123.png",
    "share_text": "我正在参加活动，快来一起玩！"
  }
}
```

---

### 获取我生成的头像列表

**GET** `/flag/my-avatars`

**Headers:**
```
Authorization: Bearer <token>
```

---

## 头像分类 (Avatar Categories)

> 头像分类管理（需要管理员权限）

### 获取头像分类列表

**GET** `/flag/avatar-categories`

**Query Parameters:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | string | 否 | 类型: public/custom |

**响应示例:**

```json
{
  "code": 0,
  "msg": "ok",
  "data": [
    {
      "id": 1,
      "tenant_id": 0,
      "code": "selfie",
      "name": "自拍头像",
      "icon": "/icons/selfie.png",
      "type": "public",
      "sort": 1,
      "status": 1
    }
  ]
}
```

---

### 创建头像分类

**POST** `/flag/avatar-categories`

**Headers:**
```
Authorization: Bearer <token>
```

**Request Body:**

```json
{
  "code": "my-category",
  "name": "我的分类",
  "icon": "/icons/custom.png",
  "type": "custom",
  "sort": 10
}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| code | string | 是 | 分类编码 |
| name | string | 是 | 分类名称 |
| icon | string | 否 | 分类图标 |
| type | string | 否 | 类型: public/custom |
| sort | int | 否 | 排序号 |

---

### 更新头像分类

**PUT** `/flag/avatar-categories/:id`

**Headers:**
```
Authorization: Bearer <token>
```

---

### 删除头像分类

**DELETE** `/flag/avatar-categories/:id`

**Headers:**
```
Authorization: Bearer <token>
```

---

## 头像管理 (Avatars)

### 获取头像列表

**GET** `/flag/avatars`

**Query Parameters:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| category_id | uint | 否 | 分类ID |
| user_id | uint | 否 | 用户ID |
| type | string | 否 | 类型: custom/system |
| page | int | 否 | 页码，默认 1 |
| size | int | 否 | 每页数量，默认 20 |

**响应示例:**

```json
{
  "code": 0,
  "msg": "ok",
  "data": [
    {
      "id": 1,
      "tenant_id": 1,
      "user_id": 1,
      "category_id": 1,
      "name": "我的头像1",
      "source_url": "/avatars/source/1.png",
      "thumbnail_url": "/avatars/thumb/1.png",
      "file_size": 102400,
      "width": 500,
      "height": 500,
      "type": "custom",
      "is_public": true,
      "like_count": 10,
      "view_count": 100,
      "status": 1
    }
  ]
}
```

---

### 获取单个头像

**GET** `/flag/avatars/:id`

---

### 上传头像

**POST** `/flag/avatars`

**Headers:**
```
Authorization: Bearer <token>
```

**Request Body:**

```json
{
  "category_id": 1,
  "name": "我的头像",
  "source_url": "https://cos.example.com/avatars/123.jpg",
  "thumbnail_url": "https://cos.example.com/avatars/123_thumb.jpg",
  "file_size": 204800,
  "width": 500,
  "height": 500,
  "is_public": true
}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| source_url | string | 是 | 原图URL |
| category_id | uint | 否 | 分类ID |
| name | string | 否 | 头像名称 |
| thumbnail_url | string | 否 | 缩略图URL |
| file_size | int64 | 否 | 文件大小(字节) |
| width | int | 否 | 图片宽度 |
| height | int | 否 | 图片高度 |
| is_public | bool | 否 | 是否公开 |

---

### 更新头像

**PUT** `/flag/avatars/:id`

**Headers:**
```
Authorization: Bearer <token>
```

---

### 删除头像

**DELETE** `/flag/avatars/:id`

**Headers:**
```
Authorization: Bearer <token>
```

---

## 错误码

| 错误码 | 说明 |
|--------|------|
| 15001 | 头像框不存在 |
| 15002 | 活动空间不存在 |
| 15003 | 头像生成失败 |
| 15004 | 头像不存在 |
