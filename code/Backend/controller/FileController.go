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
		utils.RequestRedirectTo(ctx, addr, "/getsheet")
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
		utils.RequestRedirectTo(ctx, addr, "/deletesheet")
	} else {
		utils.SendResponse(ctx, success, msg, nil)
	}
}

func GetChunk(ctx iris.Context) {
	var params utils.GetChunkParams
	if !utils.GetContextParams(ctx, &params) {
		return
	}

	success, msg:= service.GetChunk(params)

	utils.SendResponse(ctx, success, msg, nil)
}