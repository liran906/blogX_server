// Path: ./core/init_mysql_es.go

package core

import (
	"blogX_server/global"
	"blogX_server/service/river_service"
	"github.com/juju/errors"
	"runtime"
)

func InitMysqlES() {
	if !global.Config.River.Enable {
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
