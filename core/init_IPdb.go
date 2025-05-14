package core

import (
	IPutil "blogX_server/utils/ip"
	"errors"
	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"github.com/sirupsen/logrus"
	"strings"
)

var searcher *xdb.Searcher

func InitIPDB() {
	var dbPath = "init/ip2region.xdb"
	_searcher, err := xdb.NewWithFileOnly(dbPath)
	if err != nil {
		logrus.Fatalln("failed to create IP searcher: %s\n", err.Error())
	}

	searcher = _searcher
}

func GetAddress(ip string) (string, error) {
	// 先判断是否是内网
	isPrivate, err := IPutil.IsPrivateIP(ip)
	if err != nil {
		logrus.Errorf("failed to check IP (%s): %s\n", ip, err)
		return "", err
	}
	if isPrivate {
		return "Private IP", nil
	}

	// 再判断公网地址
	region, err := searcher.SearchByStr(ip)
	if err != nil {
		logrus.Errorf("failed to SearchIP(%s): %s\n", ip, err)
		return "", err
	}

	_addrList := strings.Split(region, "|")
	// 输出的格式是 国家 0 省 市 运营商
	country, province, city := _addrList[0], _addrList[2], _addrList[3]

	if province != "0" && city != "0" {
		return province + "-" + city, nil
	}
	if country != "0" && province != "0" {
		return country + "-" + province, nil
	}
	if country != "0" && city != "0" {
		return country + "-" + city, nil
	}
	if country != "0" {
		return country, nil
	}
	if province != "0" {
		return province, nil
	}
	if city != "0" {
		return city, nil
	}
	logrus.Errorf("failed to get address: %s\n", region)
	return "", errors.New("failed to get address")
}
