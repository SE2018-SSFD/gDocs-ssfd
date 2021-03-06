package controller

import (
	"backend/service"
	"backend/utils"
	"github.com/kataras/iris/v12"
)

func NewSheet(ctx iris.Context) {
	var params utils.NewSheetParams
	if !utils.GetContextParams(ctx, &params) {
		return
	}

	success, msg, data:= service.NewSheet(params)

	utils.SendResponse(ctx, success, msg, data)
}

func GetSheet(ctx iris.Context) {
	var params utils.GetSheetParams
	if !utils.GetContextParams(ctx, &params) {
		return
	}

	success, msg, data, addr := service.GetSheet(params)
	if addr != "" {
		utils.RequestRedirectTo(ctx, "http://", addr, "/getsheet")
	} else {
		utils.SendResponse(ctx, success, msg, data)
	}
}

func DeleteSheet(ctx iris.Context) {
	var params utils.DeleteSheetParams
	if !utils.GetContextParams(ctx, &params) {
		return
	}

	success, msg, addr := service.DeleteSheet(params)
	if addr != "" {
		utils.RequestRedirectTo(ctx, "http://", addr, "/deletesheet")
	} else {
		utils.SendResponse(ctx, success, msg, nil)
	}
}

func RecoverSheet(ctx iris.Context) {
	var params utils.RecoverSheetParams
	if !utils.GetContextParams(ctx, &params) {
		return
	}

	success, msg:= service.RecoverSheet(params)

	utils.SendResponse(ctx, success, msg, nil)
}

func CommitSheet(ctx iris.Context) {
	var params utils.CommitSheetParams
	if !utils.GetContextParams(ctx, &params) {
		return
	}

	success, msg, data, addr := service.CommitSheet(params)
	if addr != "" {
		utils.RequestRedirectTo(ctx, "http://", addr, "/commitsheet")
	} else {
		utils.SendResponse(ctx, success, msg, data)
	}
}

func GetSheetCheckPoint(ctx iris.Context) {
	var params utils.GetSheetCheckPointParams
	if !utils.GetContextParams(ctx, &params) {
		return
	}

	success, msg, data := service.GetSheetCheckPoint(params)

	utils.SendResponse(ctx, success, msg, data)
}

func GetSheetLog(ctx iris.Context) {
	var params utils.GetSheetLogParams
	if !utils.GetContextParams(ctx, &params) {
		return
	}

	success, msg, data := service.GetSheetLog(params)

	utils.SendResponse(ctx, success, msg, data)
}

func RollbackSheet(ctx iris.Context) {
	var params utils.RollbackSheetParams
	if !utils.GetContextParams(ctx, &params) {
		return
	}

	success, msg, addr := service.RollbackSheet(params)

	if addr != "" {
		utils.RequestRedirectTo(ctx, "http://", addr, "/rollbacksheet")
	} else {
		utils.SendResponse(ctx, success, msg, nil)
	}
}

func GetChunk(ctx iris.Context) {
	service.GetChunk(ctx)
}

func GetAllChunks(ctx iris.Context) {
	var params utils.GetAllChunksParams
	if !utils.GetContextParams(ctx, &params) {
		return
	}

	success, msg, data := service.GetAllChunks(params)

	utils.SendResponse(ctx, success, msg, data)
}

func UploadChunk(ctx iris.Context) {
	success, msg, data := service.UploadChunk(ctx)

	utils.SendResponse(ctx, success, msg, data)
}