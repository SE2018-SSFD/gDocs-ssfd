package service

import (
	"backend/dao"
	"backend/lib/cache"
	"backend/model"
	"backend/utils"
	"backend/utils/config"
	"fmt"
)

// TODO: LOG is not Implemented

var sheetCache = cache.NewSheetCache(config.MaxSheetCache)

func NewSheet(params utils.NewSheetParams) (success bool, msg int, data uint) {
	uid := CheckToken(params.Token)
	if uid != 0 {
		path := fmt.Sprintf("sheet_%s_uid_%d.txt", params.Name, uid)


		fid := dao.CreateSheet(model.Sheet{
			Name: params.Name,
			Path: path,
		})

		memSheet := cache.NewMemSheet(int(params.InitRows), int(params.InitColumns))
		//if err := dao.FileCreate(path, 0); err != nil {
		//	panic(err)
		//}
		sheetCache.Add(fid, memSheet)

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

func GetSheet(params utils.GetSheetParams) (success bool, msg int, data model.Sheet) {
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
			// TODO: read in dfs
			memSheet := sheetCache.Get(sheet.Fid)
			if memSheet == nil {
				// TODO: load latest checkpoint from dfs
				panic("not really a exception")
			}

			sheet.Content = memSheet.ToStringSlice()
			_, sheet.Columns = memSheet.Shape()

			success, msg, data = true, utils.SheetGetSuccess, sheet
		}
	} else {
		success, msg = false, utils.InvalidToken
	}

	return
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
					dao.DeleteSheet(sheet.Fid)

					sheetCache.Del(sheet.Fid)
					// TODO: delete checkpoints and log in dfs

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