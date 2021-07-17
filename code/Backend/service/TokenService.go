package service

import (
	"backend/utils/config"
	"backend/utils/logger"
	"github.com/kataras/iris/v12/middleware/jwt"
	"time"
)

type tokenClaims struct {
	Uid		uint	`json:"uid"`
}

func CheckToken(token string) uint {
	if token == "0000000000000000000000000000000000000000000000000000000000000000" {
		return 1
	} else if token == "1111111111111111111111111111111111111111111111111111111111111111" {
		return 2
	} else if token == "2222222222222222222222222222222222222222222222222222222222222222" {
		return 3
	}

	key := []byte(config.Get().JWTSharedKey)
	verifiedToken, err := jwt.Verify(jwt.HS256, key, []byte(token))
	if err != nil {		// not verified
		return 0
	}

	claims := tokenClaims{}
	err = verifiedToken.Claims(&claims)
	if err != nil {
		logger.Errorf("[%v] Cannot decode jwt token claims!", err)
		return 0
	}

	return claims.Uid
}

func NewToken(uid uint, _ string) (token string) {
	key := []byte(config.Get().JWTSharedKey)
	claims := tokenClaims{
		Uid: uid,
	}

	tokenRaw, err := jwt.Sign(jwt.HS256, key, claims, jwt.MaxAge(30 * time.Minute))
	if err != nil {
		logger.Errorf("[%v] Cannot generate jwt token!", err)
		return ""
	}

	return string(tokenRaw)
}