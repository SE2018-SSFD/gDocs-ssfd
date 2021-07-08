package service

import (
	"backend/dao"
	"backend/lib/cache"
	"backend/lib/cluster"
	"backend/lib/gdocFS"
	"backend/model"
	"backend/utils"
	"backend/utils/config"
	"backend/utils/logger"
	"github.com/kataras/iris/v12"
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

			// create initial checkpoint directory
			if err := sheetCreateCheckPointDir(fid); err != nil {
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
		}


		dao.SetSheet(sheet)
		dao.AddSheetToUser(uid, fid)

		success, msg, data = true, utils.SheetNewSuccess, fid
	} else {
		success, msg, data = false, utils.InvalidToken, 0
	}

	return success, msg, data
}

func GetSheet(params utils.GetSheetParams) (success bool, msg int, data *model.Sheet, redirect string) {
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
				return true, utils.SheetIsInTrashBin, &sheet, ""
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

				inCache := true
				memSheet := getSheetCache().Get(sheet.Fid)
				if memSheet == nil {
					if memSheet, inCache = recoverSheetFromLog(sheet.Fid); memSheet == nil {
						panic("recoverSheetFromLog fails")
					}
				}
				sheet.CheckPointNum = sheetGetCheckPointNum(params.Fid)
				sheet.Content = memSheet.ToStringSlice()
				_, sheet.Columns = memSheet.Shape()
				if inCache {
					keys, evicted := getSheetCache().Put(sheet.Fid)
					commitSheetsWithCache(utils.InterfaceSliceToUintSlice(keys), evicted)
				}

				for i := 1; i < sheet.CheckPointNum; i += 1 {
					curCid := uint(i)
					filePickled, err := sheetGetPickledCheckPointFromDfs(params.Fid, curCid)
					if err != nil {
						logger.Errorf("[fid(%d)\tcid(%d)\tuid(%d)] GetSheet: fail to pickle checkpoint\n%+v",
							params.Fid, curCid, uid, err)
						continue
					}

					sheet.CheckPointBrief = append(sheet.CheckPointBrief, model.ChkpBrief{
						Cid: curCid,
						TimeStamp: filePickled.Timestamp,
					})
				}
			}

			success, msg, data = true, utils.SheetGetSuccess, &sheet
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
						sheetRoot := gdocFS.GetRootPath("sheet", sheet.Fid)
						dao.DeleteSheet(sheet.Fid)
						if err := dao.RemoveAll(sheetRoot); err != nil {
							panic(err)
						}
					}
				} else {
					if addr, isMine := cluster.FileBelongsTo(sheet.Name, sheet.Fid); !isMine {
						return success, msg, addr
					}

					if !sheet.IsDeleted {
						sheet.IsDeleted = true
						dao.SetSheet(sheet)
					} else {
						sheetRoot := gdocFS.GetRootPath("sheet", sheet.Fid)
						dao.DeleteSheet(sheet.Fid)
						if err := dao.RemoveAll(sheetRoot); err != nil {
							panic(err)
						}
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

func RecoverSheet(params utils.RecoverSheetParams) (success bool, msg int) {
	uid := CheckToken(params.Token)
	if uid != 0 {
		ownedFids := dao.GetSheetFidsByUid(uid)
		if !utils.UintListContains(ownedFids, params.Fid) {
			success, msg = false, utils.SheetNoPermission
		} else {
			sheet := dao.GetSheetByFid(params.Fid)
			if sheet.IsDeleted {
				sheet.IsDeleted = false
				dao.SetSheet(sheet)

				success, msg = true, utils.SheetRecoverSuccess
			} else {
				success, msg = false, utils.SheetAlreadyRecovered
			}
		}
	} else {
		success, msg = false, utils.InvalidToken
	}

	return success, msg
}

func CommitSheet(params utils.CommitSheetParams) (success bool, msg int, data uint, redirect string) {
	uid := CheckToken(params.Token)
	if uid != 0 {
		ownedFids := dao.GetSheetFidsByUid(uid)
		if !utils.UintListContains(ownedFids, params.Fid) {
			success, msg, data, redirect = false, utils.SheetNoPermission, 0, ""
		} else {
			sheet := dao.GetSheetByFid(params.Fid)
			if sheet.IsDeleted {
				success, msg, data, redirect = false, utils.SheetIsInTrashBin, 0, ""
			} else {
				if addr, isMine := cluster.FileBelongsTo(sheet.Name, sheet.Fid); !isMine {
					return false, msg, data, addr
				}
				memSheet := getSheetCache().Get(sheet.Fid)
				if memSheet == nil {
					return false, utils.SheetNotInCache, 0, ""
				} else {
					cid := commitOneSheetWithCache(sheet.Fid, memSheet)
					keys, evicted := getSheetCache().Put(sheet.Fid)
					commitSheetsWithCache(utils.InterfaceSliceToUintSlice(keys), evicted)
					if cid == 0 {
						return false, utils.SheetNothingToCommit, 0, ""
					} else {
						return true, utils.SheetCommitSuccess, cid, ""
					}
				}
			}
		}
	} else {
		success, msg, data, redirect = false, utils.InvalidToken, 0, ""
	}

	return success, msg, data, redirect
}

func GetSheetCheckPoint(params utils.GetSheetCheckPointParams) (success bool, msg int, data *gdocFS.SheetCheckPointPickle) {
	uid := CheckToken(params.Token)
	if uid != 0 {
		ownedFids := dao.GetSheetFidsByUid(uid)
		if !utils.UintListContains(ownedFids, params.Fid) {
			success, msg, data = false, utils.SheetNoPermission, nil
		} else {
			if params.Cid > uint(sheetGetCheckPointNum(params.Fid)) || params.Cid <= 0 {
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
			if params.Lid > uint(sheetGetCheckPointNum(params.Fid)) || params.Lid <= 0 {
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

func RollbackSheet(params utils.RollbackSheetParams) (success bool, msg int) {
	uid := CheckToken(params.Token)
	if uid != 0 {
		ownedFids := dao.GetSheetFidsByUid(uid)
		if !utils.UintListContains(ownedFids, params.Fid) {
			success, msg = false, utils.SheetNoPermission
		} else {
			if params.Cid > uint(sheetGetCheckPointNum(params.Fid)) || params.Cid <= 0 {
				return false, utils.SheetChkpDoNotExist
			} else {
				chkp, err := sheetGetPickledCheckPointFromDfs(params.Fid, params.Cid)
				if err != nil {
					panic(err)
				}

				memSheet := cache.NewMemSheetFromStringSlice(chkp.Content, chkp.Columns)
				if ms, keys, evicted := getSheetCache().Add(params.Fid, memSheet); ms != nil {
					commitSheetsWithCache(utils.InterfaceSliceToUintSlice(keys), evicted)
				}

				keys, evicted := getSheetCache().Put(params.Fid)
				commitSheetsWithCache(utils.InterfaceSliceToUintSlice(keys), evicted)

				chkpNum := sheetGetCheckPointNum(params.Fid)
				for i := 1; i <= chkpNum; i += 1 {
					curCid := params.Cid + uint(i)
					if err := sheetDeleteCheckPointFile(params.Fid, curCid); err != nil {
						logger.Errorf("[fid(%d)\tcid(%d)\tuid(%d)] Rollback Sheet: fail to delete checkpoint file\n%+v",
							params.Fid, curCid, uid, err)
					}
					if err := sheetDeleteLogFile(params.Fid, curCid); err != nil {
						logger.Errorf("[fid(%d)\tlid(%d)\tuid(%d)] Rollback Sheet: fail to delete log file\n%+v",
							params.Fid, curCid, uid, err)
					}
				}
				lastLid := params.Cid + uint(chkpNum + 1)
				if err := sheetDeleteLogFile(params.Fid, lastLid); err != nil {
					logger.Errorf("[fid(%d)\tlid(%d)\tuid(%d)] Rollback Sheet: fail to delete log file\n%+v",
						params.Fid, lastLid, uid, err)
				}

				if err := sheetCreateLogFile(params.Fid, params.Cid + 1); err != nil {
					logger.Errorf("[fid(%d)\tlid(%d)\tuid(%d)] Rollback Sheet: fail to create empty log file\n%+v",
						params.Fid, params.Cid + 1, uid, err)
				}

				success, msg = true, utils.SheetRollbackSuccess
			}
		}
	} else {
		success, msg = false, utils.InvalidToken
	}

	return success, msg
}

func GetChunk(ctx iris.Context, params utils.GetChunkParams) (success bool, msg int) {
	return
}

func UploadChunk(ctx iris.Context, params utils.UploadChunkParams) (success bool, msg int) {
	ctx.SetMaxRequestBodySize(ctx.GetContentLength() + 1 << 20)
	return
}