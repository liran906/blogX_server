// Path: ./service/river_service/master.go

package river_service

import (
	"blogX_server/global"
	"bytes"
	"fmt"
	"github.com/pingcap/errors"
	"github.com/siddontang/go-mysql/mysql"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/siddontang/go/ioutil2"
)

type masterInfo struct {
	sync.RWMutex

	Name string `toml:"bin_name"`
	Pos  uint32 `toml:"bin_pos"`

	filePath     string
	lastSaveTime time.Time
}

//func loadMasterInfo(dataDir string) (*masterInfo, error) {
//	var m masterInfo
//
//	if len(dataDir) == 0 {
//		return &m, nil
//	}
//
//	m.filePath = path.Join(dataDir, "master.info")
//	m.lastSaveTime = time.Now()
//
//	if err := os.MkdirAll(dataDir, 0755); err != nil {
//		return nil, errors.Trace(err)
//	}
//
//	f, err := os.Open(m.filePath)
//	if err != nil && !os.IsNotExist(errors.Cause(err)) {
//		return nil, errors.Trace(err)
//	} else if os.IsNotExist(errors.Cause(err)) {
//		return &m, nil
//	}
//	defer f.Close()
//
//	_, err = toml.DecodeReader(f, &m)
//	return &m, errors.Trace(err)
//}

func loadMasterInfo(dataDir string) (*masterInfo, error) {
	var m masterInfo

	if len(dataDir) == 0 {
		return &m, nil
	}

	m.filePath = path.Join(dataDir, "master.info")
	m.lastSaveTime = time.Now()

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, errors.Trace(err)
	}

	f, err := os.Open(m.filePath)
	if err != nil {
		if os.IsNotExist(errors.Cause(err)) {
			// 如果文件不存在，返回一个空的 masterInfo
			// 会在后续操作中根据当前 binlog 位置初始化
			return &m, nil
		}
		return nil, errors.Trace(err)
	}
	defer f.Close()

	_, err = toml.DecodeReader(f, &m)
	return &m, errors.Trace(err)
}

func (m *masterInfo) Save(pos mysql.Position) error {
	logrus.Debugf("save position %s", pos)

	m.Lock()
	defer m.Unlock()

	m.Name = pos.Name
	m.Pos = pos.Pos

	if len(m.filePath) == 0 {
		logrus.Warnf("canal master info file path is empty")
		return nil
	}

	// 我观察到的现象是：
	// 1|如果删除了 index 但是 var 中有配置文件，启动的时候不会报错，但也不会同步
	// 2|如果删除了 index 但是 var 中没有配置文件，启动的时候可以成功同步 mysql 的信息，
	//   但也会报错。后续继续启动也会报错 panic，且 var 中一直都没有配置文件
	// 所以这里我就先硬编码了，如果文件不存在，立即写入，不考虑时间间隔
	// 没有 master.info 的话第一次启动会报错，第二次就好了
	_, err1 := os.Stat(m.filePath)
	if os.IsNotExist(err1) {
		err := readBackupMasterInfo(global.Config.River.DataDir)
		if err == nil {
			// 恢复备份成功
			logrus.Infof("已从备份恢复 master.info")
			return nil
		} else {
			// 恢复备份失败
			logrus.Debugf("备份恢复失败 %v", err)

			//m.Name = "mysql-bin.000004"
			//m.Pos = 42597
			logrus.Infof("master.info 及备份不存在，立即创建: %s", m.filePath)
			var buf bytes.Buffer
			e := toml.NewEncoder(&buf)
			if err := e.Encode(m); err != nil {
				return errors.Trace(err)
			}
			if err1 = ioutil2.WriteFileAtomic(m.filePath, buf.Bytes(), 0644); err1 != nil {
				logrus.Errorf("创建 master.info 失败: %v", err1)
				return errors.Trace(err1)
			}
			m.lastSaveTime = time.Now()
			return nil
		}
	}

	n := time.Now()
	if n.Sub(m.lastSaveTime) < time.Second {
		logrus.Debugf("save position %s too frequent", pos)
		return nil
	}

	m.lastSaveTime = n
	var buf bytes.Buffer
	e := toml.NewEncoder(&buf)

	e.Encode(m)

	var err error
	if err = ioutil2.WriteFileAtomic(m.filePath, buf.Bytes(), 0644); err != nil {
		logrus.Errorf("canal save master info to file %s err %v", m.filePath, err)
	}
	logrus.Debugf("saved position %s to file %s", pos, m.filePath)

	return errors.Trace(err)
}

func (m *masterInfo) Position() mysql.Position {
	m.RLock()
	defer m.RUnlock()

	return mysql.Position{
		Name: m.Name,
		Pos:  m.Pos,
	}
}

func (m *masterInfo) Close() error {
	pos := m.Position()

	return m.Save(pos)
}

func readBackupMasterInfo(dataDir string) (err error) {
	backupPath := path.Join(dataDir, "backup")
	_, err = os.Stat(backupPath)
	if err != nil {
		return
	}

	dirEntry, err := os.ReadDir(backupPath)
	if err != nil {
		return fmt.Errorf("读取备份目录失败: %v", err)
	}

	backupFile := path.Join(backupPath, dirEntry[len(dirEntry)-1].Name())
	byteData, err := os.ReadFile(backupFile)
	if err != nil {
		return fmt.Errorf("读取备份文件 %s 失败: %v", backupFile, err)
	}

	// 确保目标目录存在
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %v", err)
	}

	err = os.WriteFile(path.Join(dataDir, "master.info"), byteData, 0644)
	if err != nil {
		return fmt.Errorf("恢复 master.info 失败: %v", err)
	}

	return nil
}
