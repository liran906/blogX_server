// Path: ./blogX_server/utils/jwts/enter.go

package jwts

import (
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"strings"
	"time"
)

type Claims struct {
	UserID   uint          `json:"userID"`
	Username string        `json:"username"`
	Role     enum.RoleType `json:"role"`
}

type MyClaims struct {
	Claims
	jwt.StandardClaims
}

func (m MyClaims) GetUser() (user models.UserModel, err error) {
	err = global.DB.Take(&user, m.UserID).Error
	return
}

// GetToken 服务器生成 token
func GetToken(claims Claims) (string, error) {
	cla := MyClaims{
		Claims: claims,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Duration(global.Config.Jwt.Expire) * time.Hour).Unix(), // 过期时间
			Issuer:    global.Config.Jwt.Issuer,                                                   // 签发人
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, cla)
	return token.SignedString([]byte(global.Config.Jwt.Secret)) // 进行签名生成对应的token
}

// ParseToken 解析（客户端发来的）token
func ParseToken(tokenString string) (*MyClaims, error) {
	if tokenString == "" {
		return nil, errors.New("请登录")
	}
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(global.Config.Jwt.Secret), nil
	})
	if err != nil {
		if strings.Contains(err.Error(), "token is expired") {
			return nil, errors.New("token过期")
		}
		if strings.Contains(err.Error(), "signature is invalid") {
			return nil, errors.New("token无效")
		}
		if strings.Contains(err.Error(), "invalid number of segments") {
			return nil, errors.New("token非法")
		}
		return nil, err
	}
	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

func ParseTokenFromGin(c *gin.Context) (*MyClaims, error) {
	token := c.GetHeader("token")
	if token == "" {
		token = c.Query("token")
	}

	return ParseToken(token)
}

func GetClaimsFromGin(c *gin.Context) (claims *MyClaims) {
	_claims, ok := c.Get("claims")
	if !ok {
		return
	}
	claims, ok = _claims.(*MyClaims)
	if !ok {
		return
	}
	return
}
