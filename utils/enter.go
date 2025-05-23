// Path: ./blogX_server/utils/enter.go

package utils

func InList[T comparable](suffix T, validSuffixList []T) bool {
	for _, s := range validSuffixList {
		if s == suffix {
			return true
		}
	}
	return false
}
