package core

import (
	"blogX_server/global"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
	"time"
)

func InitDB() *gorm.DB {
	dbReadConf := global.Config.DB_r  // 读库
	dbWriteConf := global.Config.DB_w // 写库

	db, err := gorm.Open(mysql.Open(dbReadConf.DSN()), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true, // 不生成外键约束
	})
	if err != nil {
		logrus.Fatalln("DB open error: ", err)
	}

	// 如果写库不为空（就是这个配置文件下有写库数据），那么就配置读写分离
	// 用了 dbresolver 库，可以自动区分读库与写库，无需代码时显性区分
	// 所以最后返回一个 db 即可
	if !dbWriteConf.IsEmpty() {
		// 这里先插播读库连接成功的消息
		logrus.Infof("DataBase (r) [%s:%d] connection successful", global.Config.DB_r.Host, global.Config.DB_r.Port)

		// 连接写库
		err := db.Use(dbresolver.Register(dbresolver.Config{
			// use `db2` as sources, `db3`, `db4` as replicas
			Sources:  []gorm.Dialector{mysql.Open(dbWriteConf.DSN())}, // 写
			Replicas: []gorm.Dialector{mysql.Open(dbReadConf.DSN())},  // 读
			// sources/replicas load balancing policy
			Policy: dbresolver.RandomPolicy{},
		}))
		if err != nil {
			logrus.Fatalln("DB resolver error: ", err)
		}
		logrus.Infof("DataBase (w) [%s:%d] connection successful", global.Config.DB_w.Host, global.Config.DB_w.Port)
		logrus.Info("DataBase resolver successful")
	} else {
		// 没有区分读写库
		logrus.Infof("DataBase [%s:%d] connection successful", global.Config.DB_r.Host, global.Config.DB_r.Port)
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Minute * 20)

	return db
}
