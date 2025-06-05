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

	rg.POST("article", mdw.BindJsonMiddleware[article_api.ArticleCreateReq], mdw.CaptchaMiddleware, mdw.AuthMiddleware, mdw.VerifySiteModeMiddleware, app.ArticleCreateView)
	rg.GET("article", mdw.BindQueryMiddleware[article_api.ArticleListReq], app.ArticleListView)
	rg.GET("article/:id", mdw.BindUriMiddleware[models.IDRequest], app.ArticleDetailView)
	rg.PUT("article/pin", mdw.BindJsonMiddleware[article_api.ArticlePinReq], mdw.AuthMiddleware, app.ArticlePinView)
	rg.PUT("article", mdw.BindJsonMiddleware[article_api.ArticleUpdateReq], mdw.AuthMiddleware, mdw.VerifySiteModeMiddleware, app.ArticleUpdateView)
	rg.POST("article/review", mdw.BindJsonMiddleware[article_api.ArticleReviewReq], mdw.AdminMiddleware, app.ArticleReviewView)
	rg.POST("article/like/:id", mdw.BindUriMiddleware[models.IDRequest], mdw.AuthMiddleware, app.ArticleLikeView)
	rg.POST("article/collect/", mdw.BindJsonMiddleware[article_api.ArticleCollectReq], mdw.AuthMiddleware, app.ArticleCollectView)
	rg.POST("article/history", mdw.BindJsonMiddleware[article_api.ArticleReadCountReq], app.ArticleReadCountView)
	rg.DELETE("article/:id", mdw.BindUriMiddleware[models.IDRequest], mdw.AuthMiddleware, app.ArticleRemoveView)
	rg.DELETE("article", mdw.BindJsonMiddleware[models.RemoveRequest], mdw.AdminMiddleware, app.ArticleBatchRemoveView)
}
