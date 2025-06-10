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

	//sc := make(chan os.Signal, 1)
	//signal.Notify(sc)//os.Kill,
	//os.Interrupt,
	//syscall.SIGHUP,
	//syscall.SIGINT,
	//syscall.SIGTERM,
	//syscall.SIGQUIT,

	r, err := river_service.NewRiver()
	if err != nil {
		println(errors.ErrorStack(err))
		return
	}

	go func() {
		r.Run(false)
		r.Close()
	}()

	//select {
	//case n := <-sc:
	//	logrus.Debugf("receive signal %v, closing", n)
	//case <-rs.Ctx().Done():
	//	logrus.Debugf("context is done with %v, closing", rs.Ctx().Err())
	//}
}
