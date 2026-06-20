// Package dict 数据字典 - 请求/响应类型
package dict

// listRequest 字典列表请求
type listRequest struct {
	Keyword string `form:"keyword"`
	Page    int    `form:"page,default=1"`
	Size    int    `form:"size,default=20"`
}

// listResponse 字典列表响应
type listResponse struct {
	List  []Dict `json:"list"`
	Total int64  `json:"total"`
	Page  int    `form:"page"`
	Size  int    `form:"size"`
}

// createRequest 创建字典（租户自建字典）
type createRequest struct {
	Code   string                 `json:"code" binding:"required,max=32"`
	Name   string                 `json:"name" binding:"required,max=64"`
	Sort   int                    `json:"sort"`
	Extend map[string]interface{} `json:"extend"`
}

// updateRequest 更新字典基础信息（不含 code；code 是主键语义的一部分）
type updateRequest struct {
	Name   string                 `json:"name" binding:"required,max=64"`
	Sort   int                    `json:"sort"`
	Status int8                   `json:"status" binding:"omitempty,oneof=0 1 2"`
	Extend map[string]interface{} `json:"extend"`
}

// createItemRequest 在指定字典下创建字典项
type createItemRequest struct {
	Code   string                 `json:"code" binding:"required,max=64"`
	Name   string                 `json:"name" binding:"required,max=128"`
	Sort   int                    `json:"sort"`
	Extend map[string]interface{} `json:"extend"`
}

// updateItemRequest 更新字典项
type updateItemRequest struct {
	Name   string                 `json:"name" binding:"required,max=128"`
	Sort   int                    `json:"sort"`
	Status int8                   `json:"status" binding:"omitempty,oneof=0 1 2"`
	Extend map[string]interface{} `json:"extend"`
}

// listItemsRequest 字典项列表请求（按 dict_id 过滤）
type listItemsRequest struct {
	DictID uint `form:"dict_id" binding:"required"`
}

// ============ Phase 0022: 平台字典 / visibility / override 类型 ============

// platformDictCreateRequest super_admin 创建平台字典
type platformDictCreateRequest struct {
	Code       string                 `json:"code" binding:"required,max=32"`
	Name       string                 `json:"name" binding:"required,max=64"`
	Sort       int                    `json:"sort"`
	Visibility string                 `json:"visibility" binding:"omitempty,oneof=all whitelist blacklist"`
	Extend     map[string]interface{} `json:"extend"`
}

// platformDictUpdateRequest super_admin 更新平台字典
type platformDictUpdateRequest struct {
	Name       string                 `json:"name" binding:"required,max=64"`
	Sort       int                    `json:"sort"`
	Status     int8                   `json:"status" binding:"omitempty,oneof=0 1 2"`
	Visibility string                 `json:"visibility" binding:"omitempty,oneof=all whitelist blacklist"`
	Extend     map[string]interface{} `json:"extend"`
}

// platformItemCreateRequest super_admin 在平台字典下新增字典项
type platformItemCreateRequest struct {
	Code   string                 `json:"code" binding:"required,max=64"`
	Name   string                 `json:"name" binding:"required,max=128"`
	Sort   int                    `json:"sort"`
	Extend map[string]interface{} `json:"extend"`
}

// visibilityUpsertRequest super_admin 维护平台字典对某租户的访问级别
type visibilityUpsertRequest struct {
	TenantID uint   `json:"tenant_id" binding:"required"`
	Access   string `json:"access" binding:"required,oneof=invisible readonly editable"`
}

// visibilityListResponse 平台字典可见性列表
type visibilityListResponse struct {
	List  []DictVisibility `json:"list"`
	Total int64            `json:"total"`
}

// overrideUpsertRequest 租户覆盖某个平台字典项
type overrideUpsertRequest struct {
	Name   string                 `json:"name" binding:"required,max=128"`
	Sort   int                    `json:"sort"`
	Status int8                   `json:"status" binding:"omitempty,oneof=0 1 2"`
	Extend map[string]interface{} `json:"extend"`
}

// resolveResponse 合并后的字典详情（业务最终消费）
type resolveResponse struct {
	Dict ResolvedDict `json:"dict"`
}