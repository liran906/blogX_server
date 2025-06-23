// Path: ./service/cloud_service/qny_cloud_service/gen_token.go

package qny_cloud_service

import (
	"blogX_server/global"
	"context"
	"github.com/qiniu/go-sdk/v7/storagev2/credentials"
	"github.com/qiniu/go-sdk/v7/storagev2/uptoken"
	"time"
)

func GenToken() (token string, err error) {
	mac := credentials.NewCredentials(global.Config.Cloud.QNY.AccessKey, global.Config.Cloud.QNY.SecretKey)
	putPolicy, err := uptoken.NewPutPolicy(global.Config.Cloud.QNY.Bucket, time.Now().Add(time.Duration(global.Config.Cloud.QNY.Expiry)*time.Second))
	if err != nil {
		return
	}
	token, err = uptoken.NewSigner(putPolicy, mac).GetUpToken(context.Background())
	if err != nil {
		return
	}
	return
}
