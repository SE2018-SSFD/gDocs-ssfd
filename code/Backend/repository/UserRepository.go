package repository

import (
	"backend/model"
)

func GetUserByUsername(username string) (ret model.User) {
	db.First(&ret, "username = ?", username)
	return
}

func GetUserByUid(uid uint) (ret model.User) {
	db.Preload("Projects").First(&ret, "uid = ?", uid)
	return
}

func CreateUser(user model.User) uint {
	db.Create(&user)
	return user.Uid
}

func SaveUser(user model.User) {
	db.Save(&user)
	return
}

func RemoveUser(uid uint) {
	db.Delete(&model.User{
		Uid: uid,
	})
}