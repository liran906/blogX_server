// Path: ./api/data_api/site_growth.go

package data_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/redis_service/redis_site"
	"fmt"
	"github.com/gin-gonic/gin"
	"strings"
	"time"
)

// SiteGrowthReq 定义了请求参数结构体
type SiteGrowthReq struct {
	Type     int8 `form:"type" binding:"required,oneof=1 2 3 4"`   // 1-流量 2-点击率 3-新增文章 4-新增用户
	Interval int8 `form:"interval" binding:"required,oneof=1 2 3"` // 1-每天 2-每周 3-每月
	Length   int  `form:"length"`                                  // 查询的单位数量（默认为12）
}

// SiteGrowthResp 定义了响应结构体
type SiteGrowthResp struct {
	GrowthRateMap map[string]string `json:"growthRateMap"` // 增长率（百分比字符串）
	GrowthMap     map[string]int    `json:"growthMap"`     // 增长值（本期减去上期）
	ValueMap      map[string]int    `json:"valueMap"`      // 每个周期的原始值
}

// Table 用于接收数据库查询结果
type Table struct {
	Date  string `gorm:"column:date"`  // 聚合时间点（天/周/月）
	Count int    `gorm:"column:count"` // 聚合数量
}

// SiteGrowthView 是主处理函数
func (DataApi) SiteGrowthView(c *gin.Context) {
	// 从 Gin 上下文中获取绑定的请求体
	req := c.MustGet("bindReq").(SiteGrowthReq)

	// 默认长度为12
	if req.Length == 0 {
		req.Length = 12
	}

	now := time.Now()

	var startTime time.Time         // 查询起始时间
	var timeDelta time.Duration = 0 // 时间步长（仅用于天、周）
	var dataList []Table            // 查询结果列表
	var selectClause string         // SELECT 语句部分（Article/User）
	var flowClause string           // SELECT 语句部分（DataModel - 流量）
	var clickClause string          // SELECT 语句部分（DataModel - 点击率）
	var groupClauseDate string      // GROUP BY 针对 `date` 字段的语句（DataModel）
	var groupClauseCreate string    // GROUP BY 针对 `created_at` 字段的语句（Article/User）

	// 根据 interval 选择不同的时间格式和SQL语句
	switch req.Interval {
	case 1: // 按天
		startTime = now.AddDate(0, 0, -req.Length)
		timeDelta = 24 * time.Hour
		selectClause = "DATE(created_at) AS date, COUNT(id) AS count"
		flowClause = "DATE(date) AS date, SUM(flow_count) AS count"
		clickClause = "DATE(date) AS date, SUM(click_count) AS count"
		groupClauseDate = "DATE(date)"
		groupClauseCreate = "DATE(created_at)"

	case 2: // 按周
		weekday := int(now.Weekday())
		if weekday == 0 { // 周日为0，需转为7
			weekday = 7
		}
		// 获取本周的周一作为起点，再向前推 length - 1 周
		startOfWeek := now.AddDate(0, 0, -(weekday - 1))
		startTime = startOfWeek.AddDate(0, 0, -7*(req.Length-1))
		timeDelta = 7 * 24 * time.Hour

		// MySQL 中 WEEK(x,1) 表示以周一为每周的开始
		selectClause = "CONCAT(YEAR(created_at), '-', LPAD(WEEK(created_at, 1), 2, '0')) AS date, COUNT(id) AS count"
		flowClause = "CONCAT(YEAR(date), '-', LPAD(WEEK(date, 1), 2, '0')) AS date, SUM(flow_count) AS count"
		clickClause = "CONCAT(YEAR(date), '-', LPAD(WEEK(date, 1), 2, '0')) AS date, SUM(click_count) AS count"
		groupClauseDate = "CONCAT(YEAR(date), '-', LPAD(WEEK(date, 1), 2, '0'))"
		groupClauseCreate = "CONCAT(YEAR(created_at), '-', LPAD(WEEK(created_at, 1), 2, '0'))"

	case 3: // 按月
		// 从当前月份往前推 length-1 个月，并归整为每月1号
		startTime = time.Date(now.Year(), now.Month()-time.Month(req.Length)+1, 1, 0, 0, 0, 0, now.Location())
		selectClause = "DATE_FORMAT(created_at, '%Y-%m') AS date, COUNT(id) AS count"
		flowClause = "DATE_FORMAT(date, '%Y-%m') AS date, SUM(flow_count) AS count"
		clickClause = "DATE_FORMAT(date, '%Y-%m') AS date, SUM(click_count) AS count"
		groupClauseDate = "DATE_FORMAT(date, '%Y-%m')"
		groupClauseCreate = "DATE_FORMAT(created_at, '%Y-%m')"
	}

	// 执行不同类型的数据查询
	switch req.Type {
	case 1: // 流量
		global.DB.Model(models.DataModel{}).
			Where("date >= ? AND date <= ?", startTime.Format("2006-01-02")+" 00:00:00", now.Format("2006-01-02 15:04:05")).
			Select(flowClause).
			Group(groupClauseDate).
			Scan(&dataList)

	case 2: // 点击率
		global.DB.Model(models.DataModel{}).
			Where("date >= ? AND date <= ?", startTime.Format("2006-01-02")+" 00:00:00", now.Format("2006-01-02 15:04:05")).
			Select(clickClause).
			Group(groupClauseDate).
			Scan(&dataList)

	case 3: // 新增文章
		global.DB.Model(models.ArticleModel{}).
			Where("created_at >= ? AND created_at <= ? AND status = ?", startTime, now, enum.ArticleStatusPublish).
			Select(selectClause). // Debug() 可用于打印 SQL
			Group(groupClauseCreate).
			Scan(&dataList)

	case 4: // 新增用户
		global.DB.Model(models.UserModel{}).
			Where("created_at >= ? AND created_at <= ?", startTime, now).
			Select(selectClause).
			Group(groupClauseCreate).
			Scan(&dataList)
	}

	// 把查询结果转成 map[时间字符串] = 数值
	var dateMap = map[string]int{}
	for _, data := range dataList {
		date := strings.Split(data.Date, "T")[0] // 如果是时间戳格式，去掉时间部分
		dateMap[date] = data.Count
	}

	// 初始化返回结构体
	resp := SiteGrowthResp{
		ValueMap:      map[string]int{},
		GrowthMap:     map[string]int{},
		GrowthRateMap: map[string]string{},
	}

	var prev string // 记录上一个周期的 key（用于计算增长）

	for i := 0; i < req.Length; i++ {
		var currTime time.Time
		var key string        // 查询键值
		var displayKey string // 前端展示的 key（一般为日期）

		// 生成周期键值（天、周、月）
		if req.Interval == 1 {
			currTime = startTime.Add(timeDelta * time.Duration(i+1))
			key = currTime.Format("2006-01-02")
			displayKey = key
		} else if req.Interval == 2 {
			currTime = startTime.Add(timeDelta * time.Duration(i))
			_, week := currTime.ISOWeek()
			key = fmt.Sprintf("%d-%02d", currTime.Year(), week)
			displayKey = currTime.Format("2006-01-02")
		} else if req.Interval == 3 {
			currTime = startTime.AddDate(0, i, 0)
			key = currTime.Format("2006-01")
			displayKey = key
		}

		// 从数据库结果填入当前周期数据
		resp.ValueMap[displayKey] = dateMap[key]

		// 如果当前是“今天/本周/本月”，则还需加上 Redis 实时值
		_, nowWeek := now.ISOWeek()
		isLastestDay := req.Interval == 1 && key == now.Format("2006-01-02")
		isLastestWeek := req.Interval == 2 && key == fmt.Sprintf("%d-%02d", now.Year(), nowWeek)
		isLastestMonth := req.Interval == 3 && key == now.Format("2006-01")

		if isLastestDay || isLastestWeek || isLastestMonth {
			fmt.Println("now.Format(\"2006-01-02\")", now.Format("2006-01-02"))
			fmt.Println(key)
			if req.Type == 1 {
				resp.ValueMap[displayKey] += redis_site.GetFlow(now.Format("2006-01-02"))
			}
			if req.Type == 2 {
				resp.ValueMap[displayKey] += redis_site.GetClick(now.Format("2006-01-02"))
			}
		}

		// 计算增长值和增长率
		if prev != "" {
			resp.GrowthMap[displayKey] = resp.ValueMap[displayKey] - resp.ValueMap[prev]
			if resp.ValueMap[prev] == 0 {
				if resp.ValueMap[displayKey] == 0 {
					resp.GrowthRateMap[displayKey] = "0%"
				} else {
					resp.GrowthRateMap[displayKey] = "infinite%"
				}
			} else {
				rate := float64(resp.GrowthMap[displayKey]) / float64(resp.ValueMap[prev]) * 100
				resp.GrowthRateMap[displayKey] = fmt.Sprintf("%.1f%%", rate)
			}
		}
		prev = displayKey
	}

	// 返回 JSON 响应
	res.SuccessWithData(resp, c)
}
