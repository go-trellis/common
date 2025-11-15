package jwt

import (
	"github.com/golang-jwt/jwt/v5"
)

type JWT struct {
	secureKeys []byte
}

func NewJWT(secureKeys []byte) *JWT {
	return &JWT{
		secureKeys: secureKeys,
	}
}

// GenToken generates a JWT token
func (p *JWT) GenToken(claims jwt.Claims) (string, error) {
	if err := checkClaims(claims); err != nil {
		return "", err
	}
	// Create a token object with the specified signing method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Sign the token with the specified secret and get the complete encoded token string
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
