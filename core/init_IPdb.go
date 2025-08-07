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
	var dbPath = "ip2region.xdb"
	_searcher, err := xdb.NewWithFileOnly(dbPath)
	if err != nil {
		logrus.Fatalf("failed to create IP searcher: %s\n", err.Error())
	}

	searcher = _searcher
}

func GetLocationFromIP(ip string) (string, error) {
	if ip == "" {
		return "", nil
	}
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
	//fmt.Println(_addrList)

	// 输出的格式是 国家 0 省 市 运营商
	// 去掉末尾的“省”、“市”
	country, province, city := _addrList[0], _addrList[2], _addrList[3]
	if strings.HasSuffix(province, "省") {
		province = strings.TrimSuffix(province, "省")
	}
	if strings.HasSuffix(city, "市") {
		city = strings.TrimSuffix(city, "市")
	}
	if strings.HasSuffix(province, "县") {
		province = strings.TrimSuffix(province, "县")
	}
	if strings.HasSuffix(city, "县") {
		city = strings.TrimSuffix(city, "县")
	}
	if city == province || city == country {
		city = "0"
	}
	//fmt.Println(country, province, city)

	if province == "台湾" {
		return country + "-" + province, nil
	}
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
