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

	success, msg, data := service.NewSheet(params)

	utils.SendResponse(ctx, success, msg, data)
}

func ModifySheet(ctx iris.Context) {
	var params utils.ModifySheetParams
	if !utils.GetContextParams(ctx, &params) {
		return
	}

	success, msg := service.ModifySheet(params)

	utils.SendResponse(ctx, success, msg, nil)
}

func GetSheet(ctx iris.Context) {
	var params utils.GetSheetParams
	if !utils.GetContextParams(ctx, &params) {
		return
	}

	success, msg, data := service.GetSheet(params)

	utils.SendResponse(ctx, success, msg, data)
}

func DeleteSheet(ctx iris.Context) {
	var params utils.DeleteSheetParams
	if !utils.GetContextParams(ctx, &params) {
		return
	}

	success, msg := service.DeleteSheet(params)

	utils.SendResponse(ctx, success, msg, nil)
}

func CommitSheet(ctx iris.Context) {
	var params utils.CommitSheetParams
	if !utils.GetContextParams(ctx, &params) {
		return
	}

	success, msg := service.CommitSheet(params)

	utils.SendResponse(ctx, success, msg, nil)
}

func GetChunk(ctx iris.Context) {
	var params utils.GetChunkParams
	if !utils.GetContextParams(ctx, &params) {
		return
	}

	success, msg := service.GetChunk(params)

	utils.SendResponse(ctx, success, msg, nil)
}