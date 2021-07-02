package service

import (
	"backend/dao"
	"backend/lib/cache"
	"backend/lib/cluster"
	"backend/lib/gdocFS"
	"backend/model"
	"backend/utils"
	"backend/utils/config"
	"time"
)

var sheetCache *cache.SheetCache

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
		user := dao.GetUserByUid(uid)
		sheet.Path = gdocFS.GetRootPath("sheet", fid)
		sheet.Owner = user.Username

		if config.Get().WriteThrough {
			if err := sheetCreatePickledCheckPointInDfs(fid, 0, &gdocFS.SheetCheckPointPickle{
				Cid: 0,
				Timestamp: time.Now(),
				Rows: int(params.InitRows),
				Columns: int(params.InitColumns),
				Content: make([]string, params.InitRows * params.InitColumns),
			}); err != nil {
				panic(err)
			}
		} else {
			// create initial log file
			if err := sheetCreateLogFile(fid, 1); err != nil {
				panic(err)
			}

			// write one log at (rows - 1, cols - 1) to initialize a rows x cols sheet
			rows, cols := int(params.InitRows), int(params.InitColumns)
			if rows < minRows {
				rows = minRows
			}
			if cols < minCols {
				cols = minCols
			}
			appendOneSheetLog(fid, 1, &gdocFS.SheetLogPickle{
				Lid: 1,
				Timestamp: time.Now(),
				Row: rows - 1,
				Col: cols - 1,
				Old: "",
				New: "",
				Uid: uid,
				Username: user.Username,
			})

			sheet.Columns = cols
		}


		dao.SetSheet(sheet)
		dao.AddSheetToUser(uid, fid)

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
				filePickled, err := sheetGetPickledCheckPointFromDfs(params.Fid, 0)
				if err != nil {
					panic(err)
				}
				sheet.Columns = filePickled.Columns
				sheet.Content = filePickled.Content
			} else {
				if addr, isMine := cluster.FileBelongsTo(sheet.Name, sheet.Fid); !isMine {
					return success, msg, data, addr
				}

				memSheet := getSheetCache().Get(sheet.Fid)
				if memSheet == nil {
					if memSheet = recoverSheetFromLog(&sheet); memSheet == nil {
						panic("recoverSheetFromLog fails")
					}
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

	return success, msg, redirect
}

func GetSheetCheckPoint(params utils.GetSheetCheckPointParams) (success bool, msg int, data *gdocFS.SheetCheckPointPickle) {
	uid := CheckToken(params.Token)
	if uid != 0 {
		ownedFids := dao.GetSheetFidsByUid(uid)
		if !utils.UintListContains(ownedFids, params.Fid) {
			success, msg, data = false, utils.SheetNoPermission, nil
		} else {
			sheet := dao.GetSheetByFid(params.Fid)
			if params.Cid > uint(sheet.CheckPointNum) || params.Cid == 0 {
				return false, utils.SheetChkpDoNotExist, nil
			} else {
				chkp, err := sheetGetPickledCheckPointFromDfs(params.Fid, params.Cid)
				if err != nil {
					panic(err)
				}
				success, msg, data = true, utils.SheetGetChkpSuccess, chkp
			}
		}
	} else {
		success, msg, data = false, utils.InvalidToken, nil
	}

	return success, msg, data
}

func GetSheetLog(params utils.GetSheetLogParams) (success bool, msg int, data []gdocFS.SheetLogPickle) {
	uid := CheckToken(params.Token)
	if uid != 0 {
		ownedFids := dao.GetSheetFidsByUid(uid)
		if !utils.UintListContains(ownedFids, params.Fid) {
			success, msg, data = false, utils.SheetNoPermission, nil
		} else {
			sheet := dao.GetSheetByFid(params.Fid)
			if params.Lid > uint(sheet.CheckPointNum) || params.Lid == 0 {
				return false, utils.SheetLogDoNotExist, nil
			} else {
				log, err := sheetGetPickledLogFromDfs(params.Fid, params.Lid)
				if err != nil {
					panic(err)
				}
				success, msg, data = true, utils.SheetGetLogSuccess, log
			}
		}
	} else {
		success, msg, data = false, utils.InvalidToken, nil
	}

	return success, msg, data
}

func GetChunk(params utils.GetChunkParams) (success bool, msg int) {
	return
}