// Path: ./utils/hash/enter.go

package hash

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"mime/multipart"
	"os"
)

func Md5(data []byte) string {
	md5New := md5.New()
	md5New.Write(data)
	return hex.EncodeToString(md5New.Sum(nil))
}

func Md5FromFilePath(filePath string) (hash string, err error) {
	byteData, err := os.ReadFile(filePath)
	if err != nil {
		return
	}
	return Md5(byteData), nil
}

func Md5FromMultipartFile(file multipart.File) (hash string, err error) {
	byteData, err := io.ReadAll(file)
	if err != nil {
		return
	}
	return Md5(byteData), nil
}
