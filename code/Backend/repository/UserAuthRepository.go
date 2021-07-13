package repository

import (
	"backend/model"
)

func GetUserAuthByUsername(username string) (ret model.UserAuth) {
	db.First(&ret, "username = ?", username)
	return
}

func GetUserAuthByUid(uid uint) (ret model.UserAuth) {
	db.First(&ret, "uid = ?", uid)
	return
}

func CreateUserAuth(userauth model.UserAuth) uint {
	db.Create(&userauth)
	return userauth.Uid
}

func SaveUserAuth(userauth model.UserAuth) {
	db.Save(&userauth)
	return
}

func RemoveUserAuth(uid uint) {
	db.Delete(&model.UserAuth{
		Uid: uid,
	})
}
