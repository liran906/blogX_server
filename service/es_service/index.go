// Path: ./service/es_service/index.go

package es_service

import (
	"blogX_server/global"
	"blogX_server/service/river_service"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

func InitIndex(index, mapping string) {
	if ExistsIndex(index) {
		DeleteIndex(index)

		// 备份并删除 master.info
		backupMasterInfo()
	}
	CreateIndex(index, mapping)

	// 启动服务，dump 数据
	if !global.Config.River.Enable {
		return
	}
	runtime.GOMAXPROCS(runtime.NumCPU())
	r, err := river_service.NewRiver()
	if err != nil {
		logrus.Error("river init error: ", err)
		return
	}

	r.Run(false)

	// TODO 这里在 init 时总会报错 目前解决不了, 不管了
	r.Close()
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

func backupMasterInfo() error {
	// 基础路径
	varPath := "var"
	masterFile := filepath.Join(varPath, "master.info")
	backupDir := filepath.Join(varPath, "backup")

	// 检查源文件是否存在
	if _, err := os.Stat(masterFile); os.IsNotExist(err) {
		logrus.Info("master.info 不存在，无需备份")
		return nil
	}

	// 确保备份目录存在
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("创建备份目录失败: %v", err)
	}

	// 备份文件不超过 64 个
	dirEntry, _ := os.ReadDir(backupDir)
	for len(dirEntry) >= 64 {
		os.Remove(filepath.Join(backupDir, dirEntry[0].Name()))
	}

	// 生成备份文件名（使用当前日期）
	now := time.Now()
	backupFile := filepath.Join(backupDir, fmt.Sprintf("master_%03d%02d%02d.info",
		now.YearDay(), now.Hour(), now.Minute()))

	// 读取源文件
	content, err := os.ReadFile(masterFile)
	if err != nil {
		return fmt.Errorf("读取 master.info 失败: %v", err)
	}

	// 写入备份文件
	if err := os.WriteFile(backupFile, content, 0644); err != nil {
		return fmt.Errorf("写入备份文件失败: %v", err)
	}

	// 删除源文件
	if err := os.Remove(masterFile); err != nil {
		return fmt.Errorf("删除源文件失败: %v", err)
	}

	logrus.Debugf("已备份 master.info 到: %s", backupFile)
	return nil
}
