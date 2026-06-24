package message

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"gx1727.com/xin/framework/pkg/resp"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// List 收件箱 / 发件箱（按 ?box=inbox|sent 区分）
func (h *Handler) List(c *gin.Context) {
	var req ListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		resp.BadRequest(c, "invalid parameters: "+err.Error())
		return
	}

	out, err := h.svc.List(c.Request.Context(), req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, out)
}

// Get 单条消息详情
func (h *Handler) Get(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	msg, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, msg)
}

// Send 发送站内信
func (h *Handler) Send(c *gin.Context) {
	var req SendReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	msg, err := h.svc.Send(c.Request.Context(), req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, msg)
}

// Update 局部更新（仅发件人）
func (h *Handler) Update(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	var req UpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	msg, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, msg)
}

// MarkRead 标记已读
func (h *Handler) MarkRead(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	if err := h.svc.MarkRead(c.Request.Context(), id); err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{"ok": true})
}

// Delete 软删除
func (h *Handler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, gin.H{"ok": true})
}

// UnreadCount 未读数（导航栏 badge）
func (h *Handler) UnreadCount(c *gin.Context) {
	out, err := h.svc.CountUnread(c.Request.Context())
	if err != nil {
		resp.HandleError(c, err)
		return
	}
	resp.Success(c, out)
}

func parseID(c *gin.Context) (uint, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}