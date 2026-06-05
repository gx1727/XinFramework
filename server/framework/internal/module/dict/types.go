// Package dict ????
package dict

// listRequest ??????
type listRequest struct {
	Keyword string `form:"keyword"`
	Page    int    `form:"page,default=1"`
	Size    int    `form:"size,default=20"`
}

// listResponse ??????
type listResponse struct {
	List  []Dict `json:"list"`
	Total int64  `json:"total"`
	Page  int    `json:"page"`
	Size  int    `json:"size"`
}

// createRequest ????
type createRequest struct {
	Code   string                 `json:"code" binding:"required,max=32"`
	Name   string                 `json:"name" binding:"required,max=64"`
	Sort   int                    `json:"sort"`
	Extend map[string]interface{} `json:"extend"`
}

// updateRequest ??????????? code?code ??????????
type updateRequest struct {
	Name   string                 `json:"name" binding:"required,max=64"`
	Sort   int                    `json:"sort"`
	Status int8                   `json:"status" binding:"omitempty,oneof=0 1 2"`
	Extend map[string]interface{} `json:"extend"`
}

// createItemRequest ???????????
type createItemRequest struct {
	Code   string                 `json:"code" binding:"required,max=64"`
	Name   string                 `json:"name" binding:"required,max=128"`
	Sort   int                    `json:"sort"`
	Extend map[string]interface{} `json:"extend"`
}

// updateItemRequest ?????
type updateItemRequest struct {
	Name   string                 `json:"name" binding:"required,max=128"`
	Sort   int                    `json:"sort"`
	Status int8                   `json:"status" binding:"omitempty,oneof=0 1 2"`
	Extend map[string]interface{} `json:"extend"`
}

// listItemsRequest ????????? dict_id ???
type listItemsRequest struct {
	DictID uint `form:"dict_id" binding:"required"`
}
