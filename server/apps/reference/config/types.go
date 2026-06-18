// Package config 通用配置 - 请求/响应类型
package config

// Group 创建请求
type createGroupRequest struct {
	Code        string  `json:"code" binding:"required,max=64"`
	Name        string  `json:"name" binding:"required,max=64"`
	Description *string `json:"description" binding:"omitempty,max=255"`
	Icon        *string `json:"icon" binding:"omitempty,max=64"`
	Sort        int     `json:"sort"`
	IsPublic    bool    `json:"is_public"`
}

// Group 更新请求
type updateGroupRequest struct {
	Name        *string `json:"name" binding:"omitempty,max=64"`
	Description *string `json:"description" binding:"omitempty,max=255"`
	Icon        *string `json:"icon" binding:"omitempty,max=64"`
	Sort        *int    `json:"sort"`
	IsPublic    *bool   `json:"is_public"`
	Status      *int8   `json:"status"`
}

// Item 创建请求
type createItemRequest struct {
	Key          string      `json:"key" binding:"required,max=128"`
	Value        interface{} `json:"value"`
	DefaultValue interface{} `json:"default_value"`
	Type         string      `json:"type" binding:"required,oneof=string number boolean json image color select multiselect text password"`
	Label        *string     `json:"label" binding:"omitempty,max=128"`
	Description  *string     `json:"description" binding:"omitempty,max=512"`
	Options      interface{} `json:"options"`
	Validation   interface{} `json:"validation"`
	Sort         int         `json:"sort"`
	IsPublic     bool        `json:"is_public"`
	IsReadonly   bool        `json:"is_readonly"`
}

// Item 更新请求
type updateItemRequest struct {
	Value       *interface{} `json:"value"`
	Label       *string      `json:"label" binding:"omitempty,max=128"`
	Description *string      `json:"description" binding:"omitempty,max=512"`
	Sort        *int         `json:"sort"`
	IsPublic    *bool        `json:"is_public"`
	IsReadonly  *bool        `json:"is_readonly"`
	Status      *int8        `json:"status"`
}

// publicConfigResponse 公共读响应：扁平化为 key→value
type publicConfigResponse struct {
	Group  string                 `json:"group"`
	Values map[string]interface{} `json:"values"`
}
