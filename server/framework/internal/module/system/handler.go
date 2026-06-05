package system

import (
	"context"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"gx1727.com/xin/framework/pkg/cache"
	"gx1727.com/xin/framework/pkg/resp"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// ServerInfo 获取服务器运行状态信息
func (h *Handler) ServerInfo(c *gin.Context) {
	// 获取 CPU 信息
	cpuPercents, _ := cpu.Percent(time.Second, false)
	var cpuPercent float64
	if len(cpuPercents) > 0 {
		cpuPercent = cpuPercents[0]
	}
	cpuInfo, _ := cpu.Info()
	var cpuModel string
	if len(cpuInfo) > 0 {
		cpuModel = cpuInfo[0].ModelName
	}

	// 获取内存信息
	vMem, _ := mem.VirtualMemory()

	// 获取磁盘信息
	diskUsage, _ := disk.Usage("/")

	info := gin.H{
		"os": gin.H{
			"go_os":   runtime.GOOS,
			"go_arch": runtime.GOARCH,
			"num_cpu": runtime.NumCPU(),
			"version": runtime.Version(),
		},
		"cpu": gin.H{
			"model":   cpuModel,
			"percent": cpuPercent,
		},
		"memory": gin.H{
			"total":        vMem.Total,
			"used":         vMem.Used,
			"free":         vMem.Free,
			"used_percent": vMem.UsedPercent,
		},
		"disk": gin.H{
			"total":        diskUsage.Total,
			"used":         diskUsage.Used,
			"free":         diskUsage.Free,
			"used_percent": diskUsage.UsedPercent,
		},
		"time": time.Now().Format(time.RFC3339),
	}

	resp.Success(c, info)
}

// ClearCache 清除系统缓存 (Redis)
func (h *Handler) ClearCache(c *gin.Context) {
	client := cache.Get()
	if client == nil {
		resp.BadRequest(c, "Redis cache is not enabled")
		return
	}

	// 清空当前 DB 的所有 key
	err := client.FlushDB(context.Background()).Err()
	if err != nil {
		resp.ServerError(c, "Failed to clear cache: "+err.Error())
		return
	}

	resp.Success(c, "Cache cleared successfully")
}
