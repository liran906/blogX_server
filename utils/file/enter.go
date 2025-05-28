// Path: ./blogX_server/utils/file/enter.go

package file

import (
	"blogX_server/global"
	"blogX_server/utils"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"github.com/h2non/filetype"
	"io"
	"strings"
)

func ImageSuffix(filename string) (suffix string, err error) {
	_list := strings.Split(filename, ".")
	if len(_list) <= 1 {
		err = errors.New("非法文件名")
		return
	}

	suffix = _list[len(_list)-1]
	whiteList := global.Config.Filter.ValidImageSuffix

	if !utils.InList(suffix, whiteList) {
		err = errors.New("非法文件格式")
		return
	}
	return
}

// FileTypeInfo 存储文件类型信息
type FileTypeInfo struct {
	Extension string // 文件扩展名（不含点）
	MIMEType  string // MIME 类型
}

// GetHashAndFileTypeFromReader 从 Reader 中获取文件哈希值和类型信息
func GetHashAndFileTypeFromReader(reader io.Reader) (string, *FileTypeInfo, error) {
	// 读取内容到 buffer
	var buf bytes.Buffer
	_, err := buf.ReadFrom(reader)
	if err != nil {
		return "", nil, err
	}
	return GetHashAndFileTypeFromBytes(buf.Bytes())
}

// GetHashAndFileTypeFromBytes 从 []byte 中获取文件哈希值和类型信息
func GetHashAndFileTypeFromBytes(data []byte) (string, *FileTypeInfo, error) {
	// 计算哈希
	hash := md5.Sum(data)
	hashStr := hex.EncodeToString(hash[:])

	// 检测文件类型
	kind, err := filetype.Match(data)
	if err != nil {
		return hashStr, nil, err
	}

	if kind == filetype.Unknown {
		return hashStr, nil, errors.New("unknown file type")
	}

	fileType := &FileTypeInfo{
		Extension: kind.Extension,
		MIMEType:  kind.MIME.Value,
	}

	return hashStr, fileType, nil
}
