// Path: ./service/es_service/index.go

package es_service

import (
	"blogX_server/global"
	"context"
	"github.com/sirupsen/logrus"
)

func InitIndex(index, mapping string) {
	if ExistsIndex(index) {
		DeleteIndex(index)
	}
	CreateIndex(index, mapping)
}

func CreateIndex(index, mapping string) {
	_, err := global.ESClient.
		CreateIndex(index).
		BodyString(mapping).Do(context.Background())
	if err != nil {
		logrus.Errorf("ES index [%s] init fail: %s", index, err)
		return
	}
	logrus.Infof("ES index [%s] init success", index)
}

// ExistsIndex 判断索引是否存在
func ExistsIndex(index string) bool {
	exists, _ := global.ESClient.IndexExists(index).Do(context.Background())
	return exists
}

func DeleteIndex(index string) {
	_, err := global.ESClient.
		DeleteIndex(index).Do(context.Background())
	if err != nil {
		logrus.Errorf("%s 索引删除失败 %s", index, err)
		return
	}
	logrus.Infof("%s 索引删除成功", index)
}
