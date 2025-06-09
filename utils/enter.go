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

func Unique[T comparable](data []T) []T {
	if len(data) == 0 {
		return []T{}
	}
	set := make(map[T]struct{})
	unique := make([]T, 0, len(data))
	for _, item := range data {
		if _, exists := set[item]; !exists {
			set[item] = struct{}{}
			unique = append(unique, item)
		}
	}
	return unique
}

func ExtractContent(content string, length int) string {
	runes := []rune(content)
	if len(runes) > length {
		return string(runes[:length]) + "..."
	}
	return content
}
