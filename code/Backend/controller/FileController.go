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

func GetChunk(ctx iris.Context) {
	var params utils.GetChunkParams
	if !utils.GetContextParams(ctx, &params) {
		return
	}

	success, msg:= service.GetChunk(ctx, params)

	utils.SendResponse(ctx, success, msg, nil)
}

func UploadChunk(ctx iris.Context) {
	var params utils.UploadChunkParams
	if !utils.GetContextParams(ctx, &params) {
		return
	}

	success, msg:= service.UploadChunk(ctx, params)

	utils.SendResponse(ctx, success, msg, nil)
}