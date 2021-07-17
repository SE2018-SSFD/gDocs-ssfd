package dao

import (
	"backend/model"
	"backend/repository"
)

func GetUserAuthByUsername(username string) model.UserAuth {
	return repository.GetUserAuthByUsername(username)
}

func GetUserAuthByUid(uid uint) model.UserAuth {
	return repository.GetUserAuthByUid(uid)
}

func CreateUserAuth(userauth model.UserAuth) uint {
	return repository.CreateUserAuth(userauth)
}

func SetUserAuth(userauth model.UserAuth) {
	repository.SaveUserAuth(userauth)
	return
}

func GetUserByUsername(username string) model.User {
	return repository.GetUserByUsername(username)
}

func GetUserByUid(uid uint) model.User {
	if uid == 0 {
		return model.User{}
	} else {
		return repository.GetUserByUid(uid)
	}
}

func CreateUser(user model.User) uint {
	return repository.CreateUser(user)
}

func SetUser(user model.User) {
	repository.SaveUser(user)
	return
}

func RemoveUser(uid uint) {
	repository.RemoveUser(uid)
}

func RemoveUserAuth(uid uint) {
	repository.RemoveUserAuth(uid)
}