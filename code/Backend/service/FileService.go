package service

import (
	"backend/dao"
	"backend/model"
	"backend/utils"
	"fmt"
)

// TODO: LOG is not Implemented

func NewSheet(params utils.NewSheetParams) (success bool, msg int, data uint) {
	uid := CheckToken(params.Token)
	if uid != 0 {
		path := fmt.Sprintf("sheet_%s_uid_%d.txt", params.Name, uid)
		if err := dao.FileCreate(path); err != nil {
			panic(err)
		}

		fid := dao.CreateSheet(model.Sheet{
			Name: params.Name,
			Path: path,
		})

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
				switch params.Oper {
				case utils.SheetInsert:
					if err := dao.FileInsert(sheet.Path, int64(params.Offset), params.Content, utils.CELL_SIZE); err != nil {
						panic(err)
					}
				case utils.SheetDelete:
					if err := dao.FileDelete(sheet.Path, int64(params.Offset), params.Length, utils.CELL_SIZE); err != nil {
						panic(err)
					}
				case utils.SheetOverwrite:
					if err := dao.FileOverwrite(sheet.Path, int64(params.Offset), params.Content, utils.CELL_SIZE); err != nil {
						panic(err)
					}
					break
				case utils.SheetModMeta:
					sheet.Name = params.Name
					if params.Columns > sheet.Columns {
						sheet.Columns = params.Columns
					}
					dao.SetSheet(sheet)
					break
				}

				success, msg = true, utils.SheetModifySuccess
			}
		}
	} else {
		success, msg = false, utils.InvalidToken
	}

	return
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
			sheet.Content = []string{""}
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
				dao.DeleteSheet(sheet.Fid)

				// TODO: delete in dfs

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