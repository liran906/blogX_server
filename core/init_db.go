// Path: ./blogX_server/core/init_db.go

package core

import (
	"blogX_server/global"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
	"strings"
	"time"
)

func InitDB() (db *gorm.DB, masterDB *gorm.DB) {
	dbReadConf := global.Config.DB_r  // 读库
	dbWriteConf := global.Config.DB_w // 写库

	// 读写分离的时候，写库是主库，读库是从库，所以优先读取写库
	db, err := gorm.Open(mysql.Open(dbWriteConf.DSN()), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true, // 不生成外键约束，否则迁移的时候会报错
	})
	if err != nil {
		if strings.Contains(err.Error(), "Unknown database") {
			db = createDB()
		} else {
			logrus.Fatalln("DB open error: ", err)
		}
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Minute * 20)

	// 主库作为一个单独的全局变量，在某些时候（比如读写分离和 gorm 事务会出 bug）有用
	masterDB = db

	// 如果读库不为空（就是这个配置文件下有读库数据），那么就配置读写分离
	// 用了 dbresolver 库，可以自动区分读库与写库，无需代码时显性区分
	// 所以最后返回一个 db 即可
	if !dbWriteConf.IsEmpty() {
		// 这里先插播读库连接成功的消息
		logrus.Infof("DataBase (w) [%s:%d] connection successful", global.Config.DB_w.Host, global.Config.DB_w.Port)

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
		logrus.Infof("DataBase (r) [%s:%d] connection successful", global.Config.DB_r.Host, global.Config.DB_r.Port)
		logrus.Info("DataBase resolver successful")
	} else {
		// 没有区分读写库
		logrus.Infof("DataBase [%s:%d] connection successful", global.Config.DB_w.Host, global.Config.DB_w.Port)
	}
	return
}

func createDB() *gorm.DB {
	// 创建数据库
	d := global.Config.DB_w
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8&parseTime=true&loc=Local", d.User, d.Password, d.Host, d.Port)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		logrus.Fatalln("DB open error: ", err)
	}

	dbName := global.Config.DB_w.DB
	createDBSQL := "CREATE DATABASE IF NOT EXISTS " + dbName + " DEFAULT CHARSET utf8mb4 COLLATE utf8mb4_general_ci;"
	if err := db.Exec(createDBSQL).Error; err != nil {
		logrus.Fatalln("Create database error: ", err.Error())
		return nil
	}
	logrus.Infoln("Database created: ", dbName)
	return db
}
