// Path: ./core/init_mysql_es.go

package core

import (
	"blogX_server/global"
	"blogX_server/service/river_service"
	"github.com/juju/errors"
	"github.com/sirupsen/logrus"
	"runtime"
)

func InitMysqlES() {
	if !global.Config.River.Enable {
		logrus.Warnln("river is not enabled, skip loading river")
		return
	}
	if global.ESClient == nil {
		logrus.Warnln("ES is not enabled, skip loading river")
		return
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	r, err := river_service.NewRiver()
	if err != nil {
		println(errors.ErrorStack(err))
		return
	}

	go func() {
		r.Run()
		r.Close()
	}()
}
