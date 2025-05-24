// Path: ./blogX_server/utils/enter.go

package utils

import (
	"crypto/md5"
	"encoding/hex"
)

func InList[T comparable](suffix T, validSuffixList []T) bool {
	for _, s := range validSuffixList {
		if s == suffix {
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
