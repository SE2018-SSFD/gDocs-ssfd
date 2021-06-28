package service

import (
	"backend/dao"
	"backend/lib/cache"
	"backend/lib/cluster"
	"backend/lib/gdocFS"
	"backend/model"
	"backend/utils"
	"backend/utils/config"
	"encoding/json"
	"time"
)

// TODO: LOG is not Implemented

var sheetCache *cache.SheetCache = nil

func getSheetCache() *cache.SheetCache {
	if sheetCache == nil {
		sheetCache = cache.NewSheetCache(config.Get().MaxSheetCache)
	}
	return sheetCache
}

func NewSheet(params utils.NewSheetParams) (success bool, msg int, data uint) {
	uid := CheckToken(params.Token)
	if uid != 0 {
		fid := dao.CreateSheet(model.Sheet{
			Name: params.Name,
		})
		sheet := dao.GetSheetByFid(fid)
		path := gdocFS.GetCheckPointPath("sheet", fid, 0)
		sheet.Path = path
		dao.SetSheet(sheet)
		dao.AddSheetToUser(uid, fid)

		if config.Get().WriteThrough {
			if err := dao.FileCreate(path, 0); err != nil {
				panic(err)
			}

			initFile := gdocFS.SheetCheckPointPickle{
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
			// TODO: record log

			memSheet := cache.NewMemSheet(int(params.InitRows), int(params.InitColumns))
			getSheetCache().Add(fid, memSheet)
		}

		success, msg, data = true, utils.SheetNewSuccess, fid
	} else {
		success, msg, data = false, utils.InvalidToken, 0
	}

	return success, msg, data
}

func GetSheet(params utils.GetSheetParams) (success bool, msg int, data model.Sheet, redirect string) {
	redirect = ""

	uid := CheckToken(params.Token)
	if uid != 0 {
		ownedFids := dao.GetSheetFidsByUid(uid)
		sheet := dao.GetSheetByFid(params.Fid)

		if sheet.Fid == 0 {
			success, msg = false, utils.SheetDoNotExist
		} else {
			if !utils.UintListContains(ownedFids, params.Fid) {
				dao.AddSheetToUser(uid, params.Fid)
			}

			if sheet.IsDeleted {
				return true, utils.SheetIsInTrashBin, sheet, ""
			}

			if config.Get().WriteThrough {
				path := gdocFS.GetCheckPointPath("sheet", params.Fid, 0)
				fileRaw, err := dao.FileGetAll(path)
				if err != nil {
					panic(err)
				}
				filePickled := gdocFS.SheetCheckPointPickle{}
				if err := json.Unmarshal([]byte(fileRaw), &filePickled); err != nil {
					panic("cannot pickle checkpoint from file")
				}
				sheet.Columns = filePickled.Columns
				sheet.Content = filePickled.Content
			} else {
				if addr, isMine := cluster.FileBelongsTo(sheet.Name, sheet.Fid); !isMine {
					return success, msg, data, addr
				}

				memSheet := getSheetCache().Get(sheet.Fid)
				if memSheet == nil {
					// TODO: load from dfs
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

func DeleteSheet(params utils.DeleteSheetParams) (success bool, msg int, redirect string) {
	redirect = ""

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
				if config.Get().WriteThrough {
					if !sheet.IsDeleted {
						sheet.IsDeleted = true
						dao.SetSheet(sheet)
					} else {
						// TODO: delete in dfs
					}
				} else {
					if addr, isMine := cluster.FileBelongsTo(sheet.Name, sheet.Fid); !isMine {
						return success, msg, addr
					}

					if !sheet.IsDeleted {
						sheet.IsDeleted = true
						dao.SetSheet(sheet)
						getSheetCache().Del(sheet.Fid)
					} else {
						// TODO: delete in dfs
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

func GetChunk(params utils.GetChunkParams) (success bool, msg int) {
	return
}