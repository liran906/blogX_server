// Path: ./utils/enter.go

package utils

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

func InList[T comparable](item T, list []T) bool {
	for _, s := range list {
		if s == item {
			return true
		}
	}
	return false
}

func InStringList(item string, list []string) bool {
	item = strings.ToLower(item)
	for _, s := range list {
		if strings.Contains(item, s) {
			return true
		}
	}
	return false
}

func Md5(data []byte) string {
	md5New := md5.New()
	md5New.Write(data)
	return hex.EncodeToString(md5New.Sum(nil))
}
