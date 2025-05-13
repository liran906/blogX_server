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
	dbConf_read := global.Config.DB_r  // 读库
	dbConf_write := global.Config.DB_w // 写库

	db, err := gorm.Open(mysql.Open(dbConf_read.DSN()), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true, // 不生成外键约束
	})
	if err != nil {
		logrus.Fatalln("DB open error: ", err)
	}
	logrus.Infoln("DB connection successful")

	// 如果写库不为空（就是这个配置文件下有写库数据），那么就配置读写分离
	// 用了 dbresolver 库，可以自动区分读库与写库，无需代码时显性区分
	// 所以最后返回一个 db 即可
	if !dbConf_write.IsEmpty() {
		err := db.Use(dbresolver.Register(dbresolver.Config{
			// use `db2` as sources, `db3`, `db4` as replicas
			Sources:  []gorm.Dialector{mysql.Open(dbConf_write.DSN())}, // 写
			Replicas: []gorm.Dialector{mysql.Open(dbConf_read.DSN())},  // 读
			// sources/replicas load balancing policy
			Policy: dbresolver.RandomPolicy{},
		}))
		if err != nil {
			logrus.Fatalln("DB resolver error: ", err)
		}
		logrus.Infoln("DB resolver successful")
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Minute * 20)

	return db
}
