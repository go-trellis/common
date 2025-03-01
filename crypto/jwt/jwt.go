package jwt

import (
	"github.com/golang-jwt/jwt/v5"
)

type JWT struct {
	secureKeys []byte
}

func NewJWT(secureKeys []byte) *JWT {
	return &JWT{}
}

// GenToken 生成JWT
func (p *JWT) GenToken(claims jwt.Claims) (string, error) {
	if err := checkClaims(claims); err != nil {
		return "", err
	}
	// 使用指定的签名方法创建签名对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 使用指定的secret签名并获得完整的编码后的字符串token
	return token.SignedString(p.secureKeys)
}

func (p *JWT) ParseJWT(tokenString string) (*jwt.Token, error) {
	tc := &TrellisClaims{}
	return p.ParseJWTWithClaims(tokenString, tc)
}

func (p *JWT) ParseJWTWithClaims(tokenString string, claims jwt.Claims) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		return p.secureKeys, nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}
