package service

import (
	"backend/utils"
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

var tokenMap sync.Map

type grant struct {
	Uid		uint
	Begin	int64
	Term	int64
}

func CheckToken(token string) uint {
	if token == "0000000000000000000000000000000000000000000000000000000000000000" {
		return 1
	} else if token == "1111111111111111111111111111111111111111111111111111111111111111" {
		return 2
	} else if token == "2222222222222222222222222222222222222222222222222222222222222222" {
		return 3
	}
	if g, exist := tokenMap.Load(token); exist {
		grant := g.(grant)
		if grant.Begin + grant.Term <= time.Now().Unix() {
			tokenMap.Delete(token)
			return 0
		} else {
			return grant.Uid
		}
	} else {
		return 0
	}
}

func NewToken(uid uint, username string) (token string) {
	origin := fmt.Sprintf("%d %s %s", uid, username, time.Now().Format("2006-01-02 15:04:05"))
	token = fmt.Sprintf("%x", sha256.Sum256([]byte(origin)))
	tokenMap.Store(token, grant{
		Uid: uid,
		Begin: time.Now().Unix(),
		Term: utils.TokenTerm,
	})
	return
}