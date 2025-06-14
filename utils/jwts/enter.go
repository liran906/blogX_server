// Path: ./utils/jwts/enter.go

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
	UserID    uint          `json:"userID"`
	Username  string        `json:"username"`
	Role      enum.RoleType `json:"role"`
	CreatedAt time.Time     `json:"createdAt"`
}

type MyClaims struct {
	Claims
	jwt.StandardClaims
}

func (m MyClaims) GetUserFromClaims() (user *models.UserModel, err error) {
	err = global.DB.Take(&user, m.UserID).Error
	return
}

func (m MyClaims) MustGetUserFromClaims() (user *models.UserModel) {
	global.DB.Take(&user, m.UserID)
	if user == nil {
		panic("user is nil")
	}
	return
}

// GenerateToken 服务器生成 token
func GenerateToken(claims Claims) (string, error) {
	cla := MyClaims{
		Claims: Claims{
			UserID:   claims.UserID,
			Username: claims.Username,
			Role:     claims.Role,
		},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Duration(global.Config.Jwt.Expire) * time.Hour).Unix(), // 过期时间
			Issuer:    global.Config.Jwt.Issuer,                                                   // 签发人
			IssuedAt:  time.Now().Unix(),                                                          // 签发时间
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

func ParseTokenFromRequest(c *gin.Context) (*MyClaims, error) {
	token, err := GetTokenFromRequest(c)
	if err != nil {
		return nil, err
	}
	return ParseToken(token)
}

func GetTokenFromRequest(c *gin.Context) (string, error) {
	token := c.GetHeader("token")
	if token == "" {
		token = c.Query("token")
	}
	if token == "" {
		return "", errors.New("no token found")
	}
	return token, nil
}

func GetClaimsFromRequest(c *gin.Context) (claims *MyClaims, ok bool) {
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

func MustGetClaimsFromRequest(c *gin.Context) (claims *MyClaims) {
	claims, _ = GetClaimsFromRequest(c)
	if claims == nil {
		panic("claims is nil")
	}
	return
}
