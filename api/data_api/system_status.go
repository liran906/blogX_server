// Path: ./api/data_api/system_status.go

package data_api

import (
	"blogX_server/common/res"
	"blogX_server/utils/system"
	"github.com/gin-gonic/gin"
)

type SystemStatusResp struct {
	CpuUsage  float64 `json:"cpuUsage"`
	MemUsage  float64 `json:"memUsage"`
	DiskUsage float64 `json:"diskUsage"`
}

func (DataApi) SystemStatusView(c *gin.Context) {
	var data = SystemStatusResp{
		CpuUsage:  system.GetCpuUsage(),
		MemUsage:  system.GetMemUsage(),
		DiskUsage: system.GetDiskUsage(),
	}
	res.SuccessWithData(data, c)
}
