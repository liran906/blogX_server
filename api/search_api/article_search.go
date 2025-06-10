// Path: ./api/search_api/article_search.go

package search_api

import (
	"blogX_server/common"
	"blogX_server/global"
	"blogX_server/models"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
)

type ArticleSearchReq struct {
	common.PageInfo
	Type int8 `form:"type" binding:"required, oneof=0 1 2 3 4"` // 0-猜你喜欢 1-最新发布 2-最多回复 3-最多点赞 4-最多收藏
}

func (SearchApi) ArticleSearchView(c *gin.Context) {
	//req := c.MustGet("bindReq").(ArticleSearchReq)
	//claims := jwts.MustGetClaimsFromGin(c)

	// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++
	query := elastic.NewBoolQuery()

	query.Must(elastic.NewMatchQuery("title", "python"))

	highlight := elastic.NewHighlight().Field("title")

	res, err := global.ESClient.
		Search(models.ArticleModel{}.GetIndex()).
		Query(query).Highlight(highlight).
		Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
	count := res.Hits.TotalHits.Value // 总数
	fmt.Println(count)
	for _, hit := range res.Hits.Hits {
		fmt.Println(string(hit.Source))
		fmt.Println(hit.Highlight["title"])
	}
}
