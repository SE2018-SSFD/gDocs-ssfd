package service

import (
	"backend/dao"
	"backend/lib/cache"
	"backend/lib/cluster"
	"backend/model"
	"backend/utils"
	"backend/utils/config"
	"encoding/json"
	"time"
)

// TODO: LOG is not Implemented

var sheetCache = cache.NewSheetCache(config.Get().MaxSheetCache)

func NewSheet(params utils.NewSheetParams) (success bool, msg int, data uint) {
	uid := CheckToken(params.Token)
	if uid != 0 {
		fid := dao.CreateSheet(model.Sheet{
			Name: params.Name,
		})
		var path string
		if config.Get().WriteThrough {
			path = utils.GetCheckPointPath("sheet", fid, 0)
		} else {
			// TODO
			path = ""
		}
		sheet := dao.GetSheetByFid(fid)
		sheet.Path = path
		dao.SetSheet(sheet)
		dao.AddSheetToUser(uid, fid)

		if config.Get().WriteThrough {
			if err := dao.FileCreate(path, 0); err != nil {
				panic(err)
			}

			initFile := utils.CheckPointPickle{
				Cid: 0,
				Timestamp: time.Now(),
				Rows: int(params.InitRows),
				Columns: int(params.InitColumns),
				Content: make([]string, params.InitRows * params.InitColumns),
			}

			initFileRaw, _ := json.Marshal(initFile)
			if err := dao.FileOverwriteAll(path, string(initFileRaw)); err != nil {
				panic(err)
			}
		} else {
			memSheet := cache.NewMemSheet(int(params.InitRows), int(params.InitColumns))
			sheetCache.Add(fid, memSheet)
		}

		success, msg, data = true, utils.SheetNewSuccess, fid
	} else {
		success, msg, data = false, utils.InvalidToken, 0
	}

	return
}

func ModifySheet(params utils.ModifySheetParams) (success bool, msg int) {
	uid := CheckToken(params.Token)
	if uid != 0 {
		ownedFids := dao.GetSheetFidsByUid(uid)
		if !utils.UintListContains(ownedFids, params.Fid) {
			success, msg = false, utils.SheetNoPermission
		} else {
			sheet := dao.GetSheetByFid(params.Fid)
			if sheet.Fid == 0 {
				success, msg = false, utils.SheetDoNotExist
			} else {
				memSheet := sheetCache.Get(sheet.Fid)
				memSheet.Set(int(params.Row), int(params.Col), params.Content)
				// TODO: save log
				success, msg = true, utils.SheetModifySuccess
			}
		}
	} else {
		success, msg = false, utils.InvalidToken
	}

	return
}

func ModifySheetCache(fid uint, row int, col int, content string) {
	memSheet := sheetCache.Get(fid)
	memSheet.Set(row, col, content)
}

func GetSheet(params utils.GetSheetParams) (success bool, msg int, data model.Sheet, redirect string) {
	uid := CheckToken(params.Token)
	redirect = ""
	if uid != 0 {
		ownedFids := dao.GetSheetFidsByUid(uid)
		sheet := dao.GetSheetByFid(params.Fid)

		if sheet.Fid == 0 {
			success, msg = false, utils.SheetDoNotExist
		} else {
			if !utils.UintListContains(ownedFids, params.Fid) {
				dao.AddSheetToUser(uid, params.Fid)
			}

			if config.Get().WriteThrough {
				path := utils.GetCheckPointPath("sheet", params.Fid, 0)
				fileRaw, err := dao.FileGetAll(path)
				if err != nil {
					panic(err)
				}
				filePickled := utils.CheckPointPickle{}
				if err := json.Unmarshal([]byte(fileRaw), &filePickled); err != nil {
					panic("cannot pickle checkpoint from file")
				}
				sheet.Columns = filePickled.Columns
				sheet.Content = filePickled.Content
			} else {
				if addr, isMine := cluster.FileBelongsTo(sheet.Name, sheet.Fid); !isMine {
					return success, msg, data, addr
				}

				memSheet := sheetCache.Get(sheet.Fid)
				if memSheet == nil {
					// TODO
					panic("not really a exception")
				}
				sheet.Content = memSheet.ToStringSlice()
				_, sheet.Columns = memSheet.Shape()
			}

			success, msg, data = true, utils.SheetGetSuccess, sheet
		}
	} else {
		success, msg = false, utils.InvalidToken
	}

	return success, msg, data, redirect
}

func DeleteSheet(params utils.DeleteSheetParams) (success bool, msg int) {
	uid := CheckToken(params.Token)
	if uid != 0 {
		ownedFids := dao.GetSheetFidsByUid(uid)
		if !utils.UintListContains(ownedFids, params.Fid) {
			success, msg = false, utils.SheetNoPermission
		} else {
			sheet := dao.GetSheetByFid(params.Fid)
			if sheet.Fid == 0 {
				success, msg = false, utils.SheetDoNotExist
			} else {
				if sheet.IsDeleted == false {
					sheet.IsDeleted = true
					dao.SetSheet(sheet)
				} else {
					if config.Get().WriteThrough {

					} else {
						dao.DeleteSheet(sheet.Fid)

						sheetCache.Del(sheet.Fid)
						// TODO: delete checkpoints and log in dfs
					}

				}


				success, msg = true, utils.SheetDeleteSuccess
			}
		}
	} else {
		success, msg = false, utils.InvalidToken
	}

	return
}

func CommitSheet(params utils.CommitSheetParams) (success bool, msg int) {
	return
}

func GetChunk(params utils.GetChunkParams) (success bool, msg int) {
	return
}