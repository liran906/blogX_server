// Path: ./api/image_api/image_cache.go

package image_api

import (
	"blogX_server/common/res"
	"blogX_server/service/log_service"
	"blogX_server/utils/hash"
	"blogX_server/utils/jwts"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

type ImageCacheReq struct {
	URL string `json:"url"`
}

func (ImageApi) ImageCacheView(c *gin.Context) {
	req := c.MustGet("bindReq").(ImageCacheReq)

	claims := jwts.MustGetClaimsFromGin(c)

	resp, err := http.Get(req.URL)
	if err != nil {
		res.FailWithMsg("图片请求失败", c)
		return
	}
	byteData, err := io.ReadAll(resp.Body)
	if err != nil {
		res.FailWithMsg("图片读取失败", c)
		return
	}

	log := log_service.GetActionLog(c)
	log.ShowAll()
	log.SetTitle("转存图片")

	var suffix string
	switch resp.Header.Get("Content-Type") {
	case "image/jpeg", "image/jpg":
		suffix = "jpg"
	case "image/png":
		suffix = "png"
	case "image/gif":
		suffix = "gif"
	case "image/webp":
		suffix = "webp"
	case "image/avif":
		suffix = "avif"
	case "image/svg+xml":
		suffix = "svg"
	case "image/bmp":
		suffix = "bmp"
	case "image/tiff":
		suffix = "tiff"
	default:
		suffix = ""
	}

	hashString := hash.Md5(byteData)
	filename := fmt.Sprintf("%s.%s", hashString, suffix)

	rCode, msg := uploadImage(byteData, filename, req.URL, claims.UserID)
	switch rCode {
	case respCodeFail:
		res.FailWithMsg(msg, c)
	case respCodeSuccess:
		res.SuccessWithMsg(msg, c)
	case respCodeDupe:
		res.FailWithMsg(msg, c)
	}
}
