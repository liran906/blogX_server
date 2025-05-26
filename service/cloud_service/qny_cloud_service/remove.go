// Path: ./blogX_server/service/cloud_service/qny_cloud_service/remove.go

package qny_cloud_service

import (
	"blogX_server/global"
	"context"
	"fmt"
	"github.com/qiniu/go-sdk/v7/storagev2/credentials"
	"github.com/qiniu/go-sdk/v7/storagev2/http_client"
	"github.com/qiniu/go-sdk/v7/storagev2/objects"
	"strings"
)

func RemoveFile(url string) error {
	mac := credentials.NewCredentials(global.Config.Cloud.QNY.AccessKey, global.Config.Cloud.QNY.SecretKey)

	objectsManager := objects.NewObjectsManager(&objects.ObjectsManagerOptions{
		Options: http_client.Options{Credentials: mac},
	})

	bucketName := global.Config.Cloud.QNY.Bucket
	bucket := objectsManager.Bucket(bucketName)
	key := strings.TrimPrefix(url, global.Config.Cloud.QNY.Uri+"/")

	fmt.Println(bucketName, bucket, key)

	err := bucket.Object(key).Delete().Call(context.Background())
	return err
}
