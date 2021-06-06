package controller

import (
	"backend/service"
	"backend/utils"
	"github.com/kataras/iris/v12"
)

func Login(ctx iris.Context) {
	var params utils.LoginParams
	if !utils.GetContextParams(ctx, &params) {
		return
	}

	success, msg, data := service.Login(params)

	utils.SendResponse(ctx, success, msg, data)

	return
}

func Register(ctx iris.Context) {
	var params utils.RegisterParams
	if !utils.GetContextParams(ctx, &params) {
		return
	}

	success, msg, data := service.Register(params)

	utils.SendResponse(ctx, success, msg, data)

	return
}

func GetUser(ctx iris.Context) {
	var params utils.GetUserParams
	if !utils.GetContextParams(ctx, &params) {
		return
	}

	success, msg, data := service.GetUser(params)

	if success {
		utils.SendResponse(ctx, success, msg, data)
	} else {
		utils.SendResponse(ctx, success, msg, nil)
	}

	return
}

func ModifyUser(ctx iris.Context) {
	var params utils.ModifyUserParams
	if !utils.GetContextParams(ctx, &params) {
		return
	}

	success, msg := service.ModifyUser(params)

	utils.SendResponse(ctx, success, msg, nil)
}

func ModifyUserAuth(ctx iris.Context) {
	var params utils.ModifyUserAuthParams
	if !utils.GetContextParams(ctx, &params) {
		return
	}

	success, msg := service.ModifyUserAuth(params)

	utils.SendResponse(ctx, success, msg, nil)
}