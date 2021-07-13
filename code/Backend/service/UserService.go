package service

import (
	"backend/dao"
	"backend/model"
	"backend/utils"
)

func Login(params utils.LoginParams) (success bool, msg int, data map[string]interface{}) {
	query := dao.GetUserAuthByUsername(params.Username)
	if query.Uid == 0 {
		success, msg, data = false, utils.LoginNoSuchUser, nil
	} else if query.Password != params.Password {
		success, msg, data = false, utils.LoginWrongPassword, nil
	} else {
		data = make(map[string]interface{})
		data["token"] = NewToken(query.Uid, query.Username)
		data["info"] = dao.GetUserByUid(query.Uid)
		success, msg = true, utils.LoginSuccess
	}
	return
}

func Register(params utils.RegisterParams) (success bool, msg int, data string) {
	query := dao.GetUserAuthByUsername(params.Username)
	if query.Uid == 0 {
		dao.CreateUserAuth(model.UserAuth{
			Username: params.Username,
			Password: params.Password,
		})
		uid := dao.CreateUser(model.User{
			Username: params.Username,
		})

		success, msg, data = true, utils.RegisterSuccess, NewToken(uid, params.Username)
	} else {
		success, msg, data = false, utils.RegisterUserExists, ""
	}
	return
}

func GetUser(params utils.GetUserParams) (success bool, msg int, data model.User) {
	uid := CheckToken(params.Token)
	if uid != 0 {
		success, msg, data = true, utils.UserGetSuccess, dao.GetUserByUid(uid)
	} else {
		success, msg, data = false, utils.InvalidToken, model.User{}
	}

	return
}

func ModifyUser(params utils.ModifyUserParams) (success bool, msg int) {
	uid := CheckToken(params.Token)
	if uid != 0 {
		user := dao.GetUserByUid(uid)
		if params.Username != user.Username && dao.GetUserByUsername(params.Username).Uid != 0 {
			success, msg = false, utils.ModifyDupUsername
		} else {
			if params.Username != user.Username {
				userauth := dao.GetUserAuthByUid(uid)
				userauth.Username = params.Username
				dao.SetUserAuth(userauth)
			}
			user.Username = params.Username
			dao.SetUser(user)
			success, msg = true, utils.UserModifySuccess
		}
	} else {
		success, msg = false, utils.InvalidToken
	}

	return
}

func ModifyUserAuth(params utils.ModifyUserAuthParams) (success bool, msg int) {
	uid := CheckToken(params.Token)
	if uid != 0 {
		userauth := dao.GetUserAuthByUid(uid)
		userauth.Password = params.Password
		dao.SetUserAuth(userauth)
		success, msg = true, utils.UserAuthModifySuccess
	} else {
		success, msg = false, utils.InvalidToken
	}

	return
}

func RemoveUserAndUserAuth(username string) {
	uid := dao.GetUserAuthByUsername(username).Uid
	dao.RemoveUser(uid)
	dao.RemoveUserAuth(uid)
}
