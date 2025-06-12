// Path: ./flags/flag_es.go

package flags

import (
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/service/es_service"
	"github.com/sirupsen/logrus"
)

func ESInitIndex() {
	if global.ESClient == nil {
		logrus.Warnf("未开启ES")
		return
	}
	article := models.ArticleModel{}
	es_service.InitIndex(article.GetIndex(), article.Mapping())

	text := models.TextModel{}
	es_service.InitIndex(text.GetIndex(), text.Mapping())
}

// 查询某个 index:
// $ curl [ip:port]/[name]_index/_mapping
// eg. curl 192.168.88.129:9200/article_index/_mapping
