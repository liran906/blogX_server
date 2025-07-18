// Path: ./core/init_es.go

package core

import (
	"blogX_server/global"
	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
)

func InitES() *elastic.Client {
	es := global.Config.ES
	if es.Addr == "" {
		// es 吃配置，所以如果不足以运行，url 留空就不加载 es 了
		logrus.Warnln("ES addr is empty, skip loading ES")
		return nil
	}
	client, err := elastic.NewClient(
		elastic.SetURL(es.GetURL()),
		elastic.SetSniff(false),
		elastic.SetBasicAuth(es.Username, es.Password),
	)
	if err != nil {
		logrus.Panicln("ES connect error: ", err)
		return nil
	}
	logrus.Infof("ES [%s] connection successful\n", es.Addr)
	return client
}
