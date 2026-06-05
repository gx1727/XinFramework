package asset

import (
	"strconv"

	"github.com/gin-gonic/gin"
	xinContext "gx1727.com/xin/framework/pkg/context"
	"gx1727.com/xin/framework/pkg/resp"
)

type FileHandler struct {
	svc *FileService
}

func NewFileHandler(svc *FileService) *FileHandler {
	return &FileHandler{svc: svc}
}

// Upload handles file upload requests
func (h *FileHandler) Upload(c *gin.Context) {
	uc := xinContext.NewUserContext(c)
	if uc.TenantID == 0 {
		resp.Unauthorized(c, "tenant required")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		resp.HandleError(c, ErrUploadFailed.WithMsg("获取上传文件失败"))
		return
	}

	// File size check (e.g. 50MB limit)
	if file.Size > 50*1024*1024 {
		resp.HandleError(c, ErrFileTooLarge)
		return
	}

	res, err := h.svc.Upload(c.Request.Context(), uc.TenantID, uc.UserID, file)
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, res)
}

// Delete handles file deletion requests
func (h *FileHandler) Delete(c *gin.Context) {
	uc := xinContext.NewUserContext(c)
	if uc.TenantID == 0 {
		resp.Unauthorized(c, "tenant required")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		resp.HandleError(c, resp.NewError(400, "无效的文件ID"))
		return
	}

	err = h.svc.Delete(c.Request.Context(), uc.TenantID, uint(id))
	if err != nil {
		resp.HandleError(c, err)
		return
	}

	resp.Success(c, nil)
}
