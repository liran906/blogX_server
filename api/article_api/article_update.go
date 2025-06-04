// Path: ./api/article_api/article_update.go

package article_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/ctype"
	"blogX_server/models/enum"
	"blogX_server/service/log_service"
	"blogX_server/utils/jwts"
	"blogX_server/utils/markdown"
	"blogX_server/utils/xss"
	"github.com/gin-gonic/gin"
)

type ArticleUpdateReq struct {
	ArticleID      uint               `json:"articleID" binding:"required"`
	Title          string             `json:"title" binding:"required"`
	Abstract       string             `json:"abstract"`
	CoverURL       string             `json:"coverURL"`
	Content        string             `json:"content" binding:"required"`
	CategoryID     *uint              `json:"categoryID"`
	Tags           ctype.List         `json:"tags"`
	OpenForComment bool               `json:"openForComment"`
	Status         enum.ArticleStatus `json:"status" binding:"required,oneof=1 2"` // 点提交就是 2，点存为草稿就是 1
}

func (ArticleApi) ArticleUpdateView(c *gin.Context) {
	req := c.MustGet("bindReq").(ArticleUpdateReq)
	u, err := jwts.MustGetClaimsFromGin(c).GetUserFromClaims()
	if err != nil {
		res.FailWithMsg("用户信息获取失败", c)
		return
	}

	// 取文章
	var a models.ArticleModel
	err = global.DB.Take(&a, req.ArticleID).Error
	if err != nil {
		res.Fail(err, "文章不存在", c)
		return
	}

	// 非管理员只能修改自己的文章
	claims := jwts.MustGetClaimsFromGin(c)
	if claims.UserID != a.UserID && claims.Role != enum.AdminRoleType {
		res.FailWithMsg("只能修改自己的文章", c)
		return
	}

	// 取分类
	var cat models.CategoryModel
	if req.CategoryID != nil {
		err = global.DB.Take(&cat, "id = ? and user_id = ?", req.CategoryID, u.ID).Error
		if err != nil {
			res.Fail(err, "文章分类不存在", c)
			return
		}
	}

	// 文章正文防止 xss 注入
	req.Content = xss.Filter(req.Content)

	// 自动提取正文前 100 字作为摘要
	if req.Abstract == "" {
		txt, err := markdown.ExtractContent(req.Content, 100)
		if err != nil {
			res.Fail(err, "摘要提取失败", c)
		} else {
			req.Abstract = txt
		}
	} else {
		// 摘要防止 xss 注入
		req.Abstract = xss.Filter(req.Abstract)
	}

	// 正文内容图片转存给前端去做。后端留了个接口 ImageCache

	// log
	log := log_service.GetActionLog(c)
	log.ShowAll()
	log.SetTitle("修改文章")

	m := map[string]any{
		"title":            req.Title,
		"abstract":         req.Abstract,
		"cover_url":        req.CoverURL,
		"content":          req.Content,
		"category_id":      req.CategoryID,
		"Tags":             req.Tags,
		"open_for_comment": req.OpenForComment,
		"status":           req.Status,
	}
	if req.Status == enum.ArticleStatusReview && global.Config.Site.Article.AutoApprove {
		m["status"] = enum.ArticleStatusPublish
	}

	// 入库
	err = global.DB.Model(&a).Updates(m).Error
	if err != nil {
		res.Fail(err, "文章修改失败", c)
		return
	}
	res.SuccessWithMsg("文章修改成功", c)
}
