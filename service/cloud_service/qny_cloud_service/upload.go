// Path: ./service/cloud_service/qny_cloud_service/upload.go

package qny_cloud_service

import (
	"blogX_server/global"
	"blogX_server/utils/file"
	"blogX_server/utils/hash"
	"bytes"
	"context"
	"fmt"
	"github.com/qiniu/go-sdk/v7/storagev2/credentials"
	"github.com/qiniu/go-sdk/v7/storagev2/http_client"
	"github.com/qiniu/go-sdk/v7/storagev2/uploader"
	"io"
)

func UploadFile(filePath string) (QNYUrl string, err error) {
	mac := credentials.NewCredentials(global.Config.Cloud.QNY.AccessKey, global.Config.Cloud.QNY.SecretKey)

	// 获取哈希作为文件名
	hashString, err := hash.Md5FromFilePath(filePath)
	if err != nil {
		return
	}
	// 后缀
	suffix, err := file.ImageSuffix(filePath)
	if err != nil {
		return
	}
	fileName := fmt.Sprintf("%s.%s", hashString, suffix)
	key := fmt.Sprintf("%s/%s", global.Config.Cloud.QNY.Prefix, fileName)

	// 上传
	uploadManager := uploader.NewUploadManager(&uploader.UploadManagerOptions{
		Options: http_client.Options{
			Credentials: mac,
		},
	})
	err = uploadManager.UploadFile(context.Background(), filePath, &uploader.ObjectOptions{
		BucketName: global.Config.Cloud.QNY.Bucket,
		ObjectName: &key,
		FileName:   fileName,
	}, nil)
	if err != nil {
		return
	}
	QNYUrl = fmt.Sprintf("%s/%s", global.Config.Cloud.QNY.Uri, key)
	return
}

func UploadBytes(byteData []byte) (QNYUrl string, err error) {
	mac := credentials.NewCredentials(global.Config.Cloud.QNY.AccessKey, global.Config.Cloud.QNY.SecretKey)

	// 获取哈希和后缀文件格式作为文件名
	hashString, ft, err := file.GetHashAndFileTypeFromBytes(byteData)
	if err != nil || ft == nil {
		return
	}
	suffix := ft.Extension

	fileName := fmt.Sprintf("%s.%s", hashString, suffix)
	key := fmt.Sprintf("%s/%s", global.Config.Cloud.QNY.Prefix, fileName)

	uploadManager := uploader.NewUploadManager(&uploader.UploadManagerOptions{
		Options: http_client.Options{
			Credentials: mac,
		},
	})

	// 将 bytes 写入 reader（reader 中的内容只能读取一次）
	reader := bytes.NewReader(byteData)

	err = uploadManager.UploadReader(context.Background(), reader, &uploader.ObjectOptions{
		BucketName: global.Config.Cloud.QNY.Bucket,
		ObjectName: &key,
		CustomVars: map[string]string{
			"name": "github logo",
		},
		FileName: fileName,
	}, nil)
	if err != nil {
		return
	}
	QNYUrl = fmt.Sprintf("%s/%s", global.Config.Cloud.QNY.Uri, key)
	return
}

func UploadReader(reader io.Reader) (QNYUrl string, err error) {
	// 先把所有内容读到 byteData
	// 注意此时 reader 已经空了
	byteData, err := io.ReadAll(reader)
	if err != nil {
		return
	}

	return UploadBytes(byteData)
}
