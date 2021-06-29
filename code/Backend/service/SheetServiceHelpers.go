package service

import (
	"backend/dao"
	"backend/lib/cache"
	"backend/lib/gdocFS"
	"backend/model"
	"backend/utils"
	"backend/utils/logger"
	"encoding/json"
	"errors"
	"strconv"
	"time"
)

const (
	minRows = 10
	minCols = 10
)

var (
	logCommitEntry = gdocFS.SheetLogPickle{
		Timestamp: time.Now(),
		Row: -1,
		Col: -1,
	}
)

var (
	SheetFSUnrecoverableErr = errors.New("sheet filesystem is not consistent and unrecoverable")
)

// SheetFSCheck checks the consistency of sheet filesystem (fullChk: THOROUGHLY, !fullChk: BRIEFLY)
//   and make best efforts to recover it.
// A file used to be handled by a crashed server should be checked THOROUGHLY, otherwise use SheetBriefFSCheck instead.
// If the sheet filesystem is consistent, which means -- (lid == cid + 1) && (log end with commit entry),
//   current maximum cid and lid are returned.
// Or if the sheet filesystem is inconsistent and cannot not be recovered, error SheetFSUnrecoverableErr is returned.
func SheetFSCheck(fid uint, fullChk bool) (cid uint, lid uint, err error) {
	logRoot := gdocFS.GetLogRootPath("sheet", fid)
	chkpRoot := gdocFS.GetCheckPointRootPath("sheet", fid)

	// check log-only consistency
	logFileNames, err := dao.DirFilenamesAllSorted(logRoot)
	if err != nil {
		return 0, 0, err
	}

	expectLid := uint(len(logFileNames))
	for expect, actual := range logFileNames {
		curLid := uint(expect + 1)
		// check name == curLid without holes
		if strconv.Itoa(int(curLid)) != actual {
			// TODO: recover - hole in log files
			return 0, 0, SheetFSUnrecoverableErr
		}

		if fullChk {	// fullChk: check log is valid and committed
			logPath := gdocFS.GetLogPath("sheet", fid, curLid)
			if content, err := dao.FileGetAll(logPath); err != nil {
				return 0, 0, err
			} else {
				if logs, err := gdocFS.PickleSheetLogsFromContent(content); err != nil {
					// TODO: recover - log is invalid
					return 0, 0, SheetFSUnrecoverableErr
				} else {
					if lastLog := logs[len(logs) - 1]; lastLog != logCommitEntry {
						if curLid == expectLid { // last log uncommitted can be recovered by simply committing it
							// TODO: !!! recover last uncommitted log !!!
						} else {
							// TODO: recover - log is uncommitted
							return 0, 0, SheetFSUnrecoverableErr
						}
					}

					for _, log := range logs {
						if log.Lid != curLid || log.Row <= 0 || log.Col <= 0 {
							// TODO: recover - log is invalid
							return 0, 0, SheetFSUnrecoverableErr
						}
					}
				}
			}
		}
	}
	if !fullChk {	// !fullChk: check last log is committed
		logPath := gdocFS.GetLogPath("sheet", fid, expectLid)
		if content, err := dao.FileGetAll(logPath); err != nil {
			return 0, 0, err
		} else {
			if logs, err := gdocFS.PickleSheetLogsFromContent(content); err != nil {
				// TODO: recover - log is invalid
				return 0, 0, SheetFSUnrecoverableErr
			} else {
				if lastLog := logs[len(logs)-1]; lastLog != logCommitEntry {
					// TODO: !!! recover last uncommitted log !!!
				}
			}
		}
	}

	// check checkpoint-only consistency
	chkpFileNames, err := dao.DirFilenamesAllSorted(chkpRoot)
	if err != nil {
		return 0, 0, err
	}

	expectCid := uint(len(chkpFileNames))
	for expect, actual := range chkpFileNames {
		curCid := uint(expect) + 1
		// check name == curCid without holes
		if strconv.Itoa(int(curCid)) != actual {
			// TODO: recover - hole in checkpoint files
			return 0, 0, SheetFSUnrecoverableErr
		}

		if fullChk {	// fullChk: check checkpoint is valid
			chkpPath := gdocFS.GetCheckPointPath("sheet", fid, curCid)
			if content, err := dao.FileGetAll(chkpPath); err != nil {
				return 0, 0, err
			} else {
				if chkp, err := gdocFS.PickleCheckPointFromContent(content); err != nil ||
					chkp.Cid != curCid || chkp.Rows <= 0 || chkp.Columns <= 0 {
					// TODO: recover - checkpoint is invalid
					return 0, 0, SheetFSUnrecoverableErr
				}
			}
		}
	}

	// check consistency between log and checkpoint
	if expectCid + 1 != expectLid {
		// TODO: recover - cid + 1 != lid
		return 0, 0, SheetFSUnrecoverableErr
	}

	return expectCid, expectLid, nil
}

func commitOneSheetCheckPoint(fid uint, cid uint, checkpoint *gdocFS.SheetCheckPointPickle) {
	path := gdocFS.GetCheckPointPath("sheet", fid, cid)
	fileRawByte, _ := json.Marshal(*checkpoint)
	fileRaw := string(fileRawByte)
	if err := dao.FileCreate(path, 0); err != nil {
		logger.Errorf("[%s] Checkpoint file create fails!", path)
		return
	}
	if err := dao.FileOverwriteAll(path, fileRaw); err != nil {
		logger.Errorf("[%s] Checkpoint file write fails!", path)
		return
	}
}

func appendOneSheetLog(fid uint, lid uint, log *gdocFS.SheetLogPickle) {
	path := gdocFS.GetLogPath("sheet", fid, lid)
	fileRawByte, _ := json.Marshal(*log)
	fileRaw := string(fileRawByte)
	if err := dao.FileAppend(path, fileRaw); err != nil {
		logger.Errorf("[%s] Log file append fails!", path)
		return
	}
}

func commitSheetsWithCache(fids []uint, memSheets []*cache.MemSheet) {
	for ei := 0; ei < len(fids); ei += 1 {
		fid, memSheet := fids[ei], memSheets[ei]

		// update model Sheet
		sheet := dao.GetSheetByFid(fid)
		curCid := sheet.CheckPointNum
		sheet.CheckPointNum = curCid + 1
		dao.SetSheet(sheet)

		// write checkpoint to curCid+1
		cid := uint(curCid + 1)
		rows, cols := memSheet.Shape()
		commitOneSheetCheckPoint(fid, cid, &gdocFS.SheetCheckPointPickle{
			Cid: cid,
			Timestamp: time.Now(),
			Rows: rows,
			Columns: cols,
			Content: memSheet.ToStringSlice(),
		})

		// write commit entry to log with lid=curCid+1
		lid := uint(curCid + 1)
		appendOneSheetLog(fid, lid, &logCommitEntry)

		// create log with lid=curCid+2
		path := gdocFS.GetLogPath("sheet", fid, lid + 1)
		if err := dao.FileCreate(path, 0); err != nil {
			logger.Errorf("[%s] Cannot create log file!", path)
		}
	}
}

// When calling recoverSheetFromLog, log file must end with commit entry because log would be committed automatically
//   when all users quit editing or sheet is evicted from memCache.
// BUT log can be *UNCOMMITTED* if the server it belonged to crashed, for which we need to thoroughly handle
//   all possible circumstances here in order to achieve crash consistency.
func recoverSheetFromLog(sheet *model.Sheet) (memSheet *cache.MemSheet) {
	// TODO: !!! return error !!!
	fid := sheet.Fid
	curCid := uint(sheet.CheckPointNum)
	logPath := gdocFS.GetLogPath("sheet", fid, curCid + 1)

	// TODO: determine whether sheet is from crashed server and call SheetFSCheck
	// SheetFSCheck(fid, isFromCrashServer)

	if content, err := dao.FileGetAll(logPath); err == nil {
		if logs, err := gdocFS.PickleSheetLogsFromContent(content); err == nil {
			memSheet = cache.NewMemSheet(minRows, minCols)
			for li := 0; li < len(logs); li += 1 {
				log := &logs[li]
				memSheet.Set(log.Row, log.Col, log.New)
			}

			keys, evicted := getSheetCache().Add(fid, memSheet)
			commitSheetsWithCache(utils.InterfaceSliceToUintSlice(keys), evicted)
		} else {
			logger.Errorf("[fid: %d] Sheet logs cannot be pickled")
			return nil
		}
	} else {
		logger.Errorf("[%s] Log file read fails!", logPath)
		return nil
	}

	if _, cols := memSheet.Shape(); cols != sheet.Columns {
		logger.Errorf("[%s] Log columns not equal to model columns!", logPath)
	}

	return memSheet
}