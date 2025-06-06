// Path: ./router/article_router.go

package router

import (
	"blogX_server/api"
	"blogX_server/api/article_api"
	mdw "blogX_server/middleware"
	"blogX_server/models"
	"github.com/gin-gonic/gin"
)

func ArticleRouter(rg *gin.RouterGroup) {
	app := api.App.ArticleModel

	// 文章 CRUD
	rg.POST("article", mdw.BindJsonMiddleware[article_api.ArticleCreateReq], mdw.CaptchaMiddleware, mdw.AuthMiddleware, mdw.VerifySiteModeMiddleware, app.ArticleCreateView)
	rg.PUT("article", mdw.BindJsonMiddleware[article_api.ArticleUpdateReq], mdw.AuthMiddleware, mdw.VerifySiteModeMiddleware, app.ArticleUpdateView)
	rg.GET("article", mdw.BindQueryMiddleware[article_api.ArticleListReq], app.ArticleListView)
	rg.GET("article/collection", mdw.BindQueryMiddleware[article_api.ArticleCollectionListReq], app.ArticleCollectionListView)
	rg.GET("article/:id", mdw.BindUriMiddleware[models.IDRequest], app.ArticleDetailView)
	rg.DELETE("article/:id", mdw.BindUriMiddleware[models.IDRequest], mdw.AuthMiddleware, app.ArticleRemoveView)
	rg.DELETE("article", mdw.BindJsonMiddleware[models.RemoveRequest], mdw.AdminMiddleware, app.ArticleBatchRemoveView)

	// 用户置顶
	rg.PUT("article/pin", mdw.BindJsonMiddleware[article_api.ArticlePinReq], mdw.AuthMiddleware, app.ArticlePinView)

	// 审核
	rg.POST("article/review", mdw.BindJsonMiddleware[article_api.ArticleReviewReq], mdw.AdminMiddleware, app.ArticleReviewView)

	// 点赞收藏 CD
	rg.POST("article/like/:id", mdw.BindUriMiddleware[models.IDRequest], mdw.AuthMiddleware, app.ArticleLikeView)
	rg.POST("article/collect/", mdw.BindJsonMiddleware[article_api.ArticleCollectReq], mdw.AuthMiddleware, app.ArticleCollectView)

	// 浏览记录 CRD
	rg.POST("article/history", mdw.BindJsonMiddleware[article_api.ArticleCountReadReq], app.ArticleCountReadView)
	rg.GET("article/history", mdw.BindQueryMiddleware[article_api.ArticleReadListReq], mdw.AuthMiddleware, app.ArticleReadListView)
	rg.DELETE("article/history", mdw.BindJsonMiddleware[models.RemoveRequest], mdw.AuthMiddleware, app.ArticleReadRemoveView)

	// 文章分类 CRUD
	rg.POST("article/category", mdw.BindJsonMiddleware[article_api.ArticleCategoryCreateReq], mdw.AuthMiddleware, app.ArticleCategoryCreateView)
	rg.PUT("article/category", mdw.BindJsonMiddleware[article_api.ArticleCategoryUpdateReq], mdw.AuthMiddleware, app.ArticleCategoryUpdateView)
	rg.GET("article/category", mdw.BindQueryMiddleware[article_api.ArticleCategoryListReq], mdw.AuthMiddleware, app.ArticleCategoryListView)
	rg.DELETE("article/category", mdw.BindJsonMiddleware[models.RemoveRequest], mdw.AuthMiddleware, app.ArticleCategoryRemoveView)

	// 收藏夹 CRUD
	rg.POST("article/collections", mdw.BindJsonMiddleware[article_api.ArticleCollectionCreateReq], mdw.AuthMiddleware, app.ArticleCollectionCreateView)
	rg.PUT("article/collections", mdw.BindJsonMiddleware[article_api.ArticleCollectionUpdateReq], mdw.AuthMiddleware, app.ArticleCollectionUpdateView)
	rg.GET("article/collections", mdw.BindQueryMiddleware[article_api.ArticleCollectionFolderListReq], mdw.AuthMiddleware, app.ArticleCollectionFolderListView)
	rg.DELETE("article/collections", mdw.BindJsonMiddleware[models.RemoveRequest], mdw.AuthMiddleware, app.ArticleCollectionRemoveView)
}
