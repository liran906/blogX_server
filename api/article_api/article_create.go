// Path: ./api/article_api/article_create.go

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

type ArticleCreateReq struct {
	Title          string             `json:"title" binding:"required"`
	Abstract       string             `json:"abstract"`
	CoverURL       string             `json:"coverURL"`
	Content        string             `json:"content" binding:"required"`
	CategoryID     *uint              `json:"categoryID"`
	Tags           ctype.List         `json:"tags"`
	OpenForComment bool               `json:"openForComment"`
	Status         enum.ArticleStatus `json:"status" binding:"required,oneof=1 2"` // 点提交就是 2，点存为草稿就是 1
}

func (ArticleApi) ArticleCreateView(c *gin.Context) {
	req := c.MustGet("bindReq").(ArticleCreateReq)
	u, err := jwts.MustGetClaimsFromGin(c).GetUserFromClaims()
	if err != nil {
		res.FailWithMsg("用户信息获取失败", c)
		return
	}

	// 取分类
	var cat models.CategoryModel
	if req.CategoryID != nil {
		err = global.DB.Take(&cat, "id = ? and user_id = ?", req.CategoryID, u.ID).Error
		if err != nil {
			res.FailWithMsg("文章分类不存在", c)
			return
		}
	}

	// 自动提取正文前 100 字作为摘要
	if req.Abstract == "" {
		txt, err := markdown.ExtractContent(req.Content, 100)
		if err != nil {
			res.Fail(err, "摘要提取失败", c)
			return
		}
		req.Abstract = txt
	} else {
		// 摘要防止 xss 注入
		req.Abstract = xss.Filter(req.Abstract)
		// 摘要不能超过 200 字
		txt, err := markdown.ExtractContent(req.Abstract, 200)
		if err != nil {
			res.Fail(err, "摘要提取失败", c)
			return
		}
		req.Abstract = txt
	}

	// 正文内容图片转存给前端去做。后端留了个接口 ImageCache

	// log
	log := log_service.GetActionLog(c)
	log.ShowAll()
	log.SetTitle("发布文章")

	var article = models.ArticleModel{
		Title:          req.Title,
		Abstract:       req.Abstract,
		CoverURL:       req.CoverURL,
		Content:        req.Content,
		CategoryID:     req.CategoryID,
		Tags:           req.Tags,
		OpenForComment: req.OpenForComment,
		UserID:         u.ID,
		Status:         req.Status,
	}
	if req.Status == enum.ArticleStatusReview && global.Config.Site.Article.AutoApprove {
		article.Status = enum.ArticleStatusPublish
	}

	// 入库
	err = global.DB.Create(&article).Error
	if err != nil {
		res.Fail(err, "文章创建失败", c)
		return
	}
	res.SuccessWithMsg("文章创建成功", c)
}
