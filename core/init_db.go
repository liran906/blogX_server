// Path: ./core/init_db.go

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
	gdb := global.Config.DB
	n := len(gdb) // 配置数据库数量
	if n == 0 {
		logrus.Fatalln("DB is not configured, please check the settings.yaml file")
	}

	dbWriteConf := gdb[0] // 写库

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

	// 这里先插播读库连接成功的消息
	logrus.Infof("DataBase (master) [%s:%d] connection successful", gdb[0].Host, gdb[0].Port)

	if n > 1 {
		// 读取读库（从库）列表
		var replicas []gorm.Dialector
		for i := 1; i < n; i++ {
			replicas = append(replicas, mysql.Open(gdb[i].DSN()))
			logrus.Infof("DataBase (servant) [%s:%d] config successful", gdb[i].Host, gdb[i].Port)
		}

		err := db.Use(dbresolver.Register(dbresolver.Config{
			Sources:  []gorm.Dialector{mysql.Open(dbWriteConf.DSN())}, // 写
			Replicas: replicas,                                        // 读
			// sources/replicas load balancing policy
			Policy: dbresolver.RandomPolicy{},
		}))
		if err != nil {
			logrus.Fatalln("DB resolver error: ", err)
		}
		logrus.Infof("DataBase (servant) connection successful")
		logrus.Info("DataBase resolver successful")
	}
	return
}

func createDB() *gorm.DB {
	// 创建数据库
	d := global.Config.DB[0]
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8&parseTime=true&loc=Local", d.User, d.Password, d.Host, d.Port)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		logrus.Fatalln("DB open error: ", err)
	}

	dbName := global.Config.DB[0].DBname
	createDBSQL := "CREATE DATABASE IF NOT EXISTS " + dbName + " DEFAULT CHARSET utf8mb4 COLLATE utf8mb4_general_ci;"
	if err := db.Exec(createDBSQL).Error; err != nil {
		logrus.Fatalln("Create database error: ", err.Error())
		return nil
	}
	logrus.Infoln("Database created: ", dbName)
	return db
}
