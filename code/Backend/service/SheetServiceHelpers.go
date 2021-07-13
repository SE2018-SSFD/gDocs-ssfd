package service

import (
	"backend/dao"
	"backend/lib/cache"
	"backend/lib/gdocFS"
	"backend/lib/reentrantMutex"
	"backend/utils"
	"backend/utils/logger"
	"encoding/json"
	"github.com/pkg/errors"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

const (
	minRows = 10
	minCols = 10
)

var (
	logCommitEntry = gdocFS.SheetLogPickle{
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
	lockOnFid(fid)
	defer unlockOnFid(fid)

	defer func() {
		err = errors.WithStack(err)
	}()

	logRoot := gdocFS.GetLogRootPath("sheet", fid)
	chkpRoot := gdocFS.GetCheckPointRootPath("sheet", fid)

	logIndexes, err := dao.DirFilenameIndexesAllSorted(logRoot)
	if err != nil {
		return 0, 0, err
	}

	chkpFileNames, err := dao.DirFilenameIndexesAllSorted(chkpRoot)
	if err != nil {
		return 0, 0, err
	}

	expectLid := uint(len(logIndexes))
	expectCid := uint(len(chkpFileNames))

	// check consistency between log and checkpoint
	if expectCid+1 < expectLid {
		// miss checkpoints - recover from the log and make a checkpoint
		if chkp, err := sheetGetPickledCheckPointFromDfs(fid, expectCid); err != nil {
			return 0, 0, SheetFSUnrecoverableErr
		} else {
			// starting from loading the last checkpoint in checkpoint root path
			memSheet := cache.NewMemSheetFromStringSlice(chkp.Content, chkp.Columns)

			// for every missing checkpoint
			for curCid := expectCid + 1; curCid < expectLid; curCid += 1 {
				logger.Warnf("[expectCid(%d)\texpectLid(%d)\tcurCid(%d)] recover missing checkpoints",
					expectCid, expectLid, curCid)
				// redo with pickled logs
				logs, err := sheetGetPickledLogFromDfs(fid, curCid)
				if err != nil {
					return 0, 0, SheetFSUnrecoverableErr
				}
				for _, log := range logs {
					if log == logCommitEntry { // skip commit log entry
						continue
					} else {
						memSheet.Set(log.Row, log.Col, log.New)
					}
				}

				// make checkpoint
				rows, cols := memSheet.Shape()
				if err := sheetCreatePickledCheckPointInDfs(fid, curCid, &gdocFS.SheetCheckPointPickle{
					Cid:       cid,
					Timestamp: time.Now(),
					Rows:      rows,
					Columns:   cols,
					Content:   memSheet.ToStringSlice(),
				}); err != nil {
					logger.Errorf("%+v", err)
				}
			}

			expectCid = expectLid - 1
		}
	} else if expectCid+1 > expectLid {
		// miss one log, should not happen because new log is created before checkpoint
		return 0, 0, SheetFSUnrecoverableErr
	}

	// check log-only consistency
	for expect, actual := range logIndexes {
		curLid := uint(expect + 1)
		// check name == curLid without holes
		if int(curLid) != actual {
			// TODO: recover - hole in log files
			return 0, 0, SheetFSUnrecoverableErr
		}

		if fullChk { // fullChk: check log is valid and committed
			if logs, err := sheetGetPickledLogFromDfs(fid, curLid); err != nil {
				// TODO: recover - log is invalid
				return 0, 0, SheetFSUnrecoverableErr
			} else if lastLog := logs[len(logs)-1]; lastLog != logCommitEntry {
				if curLid == expectLid { // last log uncommitted can be recovered by simply committing it
					// TODO: !!! recover last uncommitted log !!!
				} else { // middle log uncommitted can be recovered?
					// TODO: recover - log is uncommitted
					return 0, 0, SheetFSUnrecoverableErr
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
	if !fullChk { // !fullChk: check last log is committed
		if logs, err := sheetGetPickledLogFromDfs(fid, expectLid); err != nil {
			// TODO: recover - log is invalid
			return 0, 0, err
		} else if len(logs) != 0 {
			if lastLog := logs[len(logs)-1]; lastLog != logCommitEntry {
				// recover last uncommitted log - just commit it
				appendOneSheetLog(fid, expectLid, &logCommitEntry)
			}
		}
	}

	// check checkpoint-only consistency
	for expect, actual := range chkpFileNames {
		curCid := uint(expect) + 1
		// check name == curCid without holes
		if int(curCid) != actual {
			// TODO: recover - hole in checkpoint files
			return 0, 0, SheetFSUnrecoverableErr
		}

		if fullChk { // fullChk: check checkpoint is valid
			if chkp, err := sheetGetPickledCheckPointFromDfs(fid, curCid); err != nil ||
				chkp.Cid != curCid || chkp.Rows <= 0 || chkp.Columns <= 0 {
				// TODO: recover - checkpoint is invalid
				return 0, 0, SheetFSUnrecoverableErr
			}
		}
	}

	return expectCid, expectLid, nil
}

func appendOneSheetLog(fid uint, lid uint, log *gdocFS.SheetLogPickle) {
	path := gdocFS.GetLogPath("sheet", fid, lid)
	fileRaw, _ := json.Marshal(*log)
	if err := dao.FileAppend(path, fileRaw); err != nil {
		logger.Errorf("[%s] Log file append fails!\n%+v", path, err)
		return
	}
}

func commitOneSheetWithCache(fid uint, memSheet *cache.MemSheet) (cid uint) {
	lockOnFid(fid)
	defer unlockOnFid(fid)
	memSheet.Lock()
	defer memSheet.Unlock()

	// update model Sheet
	curCid := uint(sheetGetCheckPointNum(fid))

	// empty log
	lid := curCid + 1
	if logs, err := sheetGetPickledLogFromDfs(fid, lid); err != nil {
		if len(logs) == 0 {
			return curCid
		}
	}

	rows, cols := memSheet.Shape()

	// 1. commit log: write commit entry to log with lid=curCid+1
	appendOneSheetLog(fid, lid, &logCommitEntry)

	// 2. create new log file: create log with lid=curCid+2
	if err := sheetCreateLogFile(fid, lid+1); err != nil {
		logger.Errorf("%+v", err)
	}

	// 3. make checkpoint: write checkpoint to curCid+1
	cid = curCid + 1
	if err := sheetCreatePickledCheckPointInDfs(fid, cid, &gdocFS.SheetCheckPointPickle{
		Cid:       cid,
		Timestamp: time.Now(),
		Rows:      rows,
		Columns:   cols,
		Content:   memSheet.ToStringSlice(),
	}); err != nil {
		logger.Errorf("%+v", err)
	}

	return cid
}

func commitSheetsWithCache(fids []uint, memSheets []*cache.MemSheet) {
	for ei := 0; ei < len(fids); ei += 1 {
		fid, memSheet := fids[ei], memSheets[ei]
		commitOneSheetWithCache(fid, memSheet)
	}
}

// When calling recoverSheetFromLog, log file must end with commit entry because log would be committed automatically
//   when all users quit editing or sheet is evicted from memCache.
// BUT log can be *UNCOMMITTED* if the server it belonged to crashed, for which we need to thoroughly handle
//   all possible circumstances here in order to achieve crash consistency.
func recoverSheetFromLog(fid uint) (memSheet *cache.MemSheet, inCache bool) {
	// TODO: determine whether sheet is from crashed server and call SheetFSCheck
	curCid := uint(sheetGetCheckPointNum(fid))
	//curCid, _, err := SheetFSCheck(fid, false)
	//if err != nil {
	//	logger.Errorf("[fid(%d)] FSCheck not passed!\n%+v", fid, err)
	//	return nil, false
	//}

	// get memSheet from scratch or latest checkpoint
	if curCid == 0 {
		memSheet = cache.NewMemSheet(minRows, minCols)
	} else {
		if chkp, err := sheetGetPickledCheckPointFromDfs(fid, curCid); err != nil {
			logger.Errorf("[%d] %+v", fid, err)
			return nil, false
		} else {
			memSheet = cache.NewMemSheetFromStringSlice(chkp.Content, chkp.Columns)
		}
	}

	// redo with latest log
	if logs, err := sheetGetPickledLogFromDfs(fid, curCid+1); err != nil {
		logger.Errorf("%+v", err)
		return nil, false
	} else {
		for li := 0; li < len(logs); li += 1 { // without logCommitEntry, which is in the end
			if logs[li] == logCommitEntry {
				continue
			} else {
				log := &logs[li]
				memSheet.Set(log.Row, log.Col, log.New)
			}
		}

		// do eviction
		if ms, keys, evicted := getSheetCache().Add(fid, memSheet); ms != nil {
			commitSheetsWithCache(utils.InterfaceSliceToUintSlice(keys), evicted)
			return ms, true
		} else {
			return memSheet, false
		}
	}
}

// sheetGetPickledCheckPointFromDfs pickles a CheckPoint from dfs with fid and cid
func sheetGetPickledCheckPointFromDfs(fid uint, cid uint) (chkp *gdocFS.SheetCheckPointPickle, err error) {
	lockOnFid(fid)
	defer unlockOnFid(fid)
	if cid == 0 {
		return &gdocFS.SheetCheckPointPickle{
			Columns: minCols,
			Content: make([]string, minRows*minCols),
		}, nil
	}
	path := gdocFS.GetCheckPointPath("sheet", fid, cid)
	if fileRaw, err := dao.FileGetAll(path); err != nil {
		return nil, errors.WithStack(err)
	} else {
		chkp, err = gdocFS.PickleSheetCheckPointFromContent(string(fileRaw))
		return chkp, errors.WithStack(err)
	}
}

// sheetWritePickledCheckPointToDfs writes a CheckPoint to a EXISTENT file in dfs with fid and cid
func sheetWritePickledCheckPointToDfs(fid uint, cid uint, chkp *gdocFS.SheetCheckPointPickle) (err error) {
	lockOnFid(fid)
	defer unlockOnFid(fid)
	path := gdocFS.GetCheckPointPath("sheet", fid, cid)
	fileRaw, _ := json.Marshal(*chkp)
	if err = dao.FileOverwriteAll(path, fileRaw); err != nil {
		return errors.WithStack(err)
	} else {
		return nil
	}
}

// sheetCreatePickledCheckPointInDfs create a CheckPoint in a NONEXISTENT file in dfs with fid and cid
func sheetCreatePickledCheckPointInDfs(fid uint, cid uint, chkp *gdocFS.SheetCheckPointPickle) (err error) {
	lockOnFid(fid)
	defer unlockOnFid(fid)
	path := gdocFS.GetCheckPointPath("sheet", fid, cid)
	if err := dao.FileCreate(path, 0); err != nil {
		return errors.WithStack(err)
	} else {
		return sheetWritePickledCheckPointToDfs(fid, cid, chkp)
	}
}

// sheetCreateCheckPointDir create a empty checkpoint directory in dfs with fid
func sheetCreateCheckPointDir(fid uint) (err error) {
	chkpRoot := gdocFS.GetCheckPointRootPath("sheet", fid)
	if err := dao.DirCreate(chkpRoot); err != nil {
		return err
	} else {
		return nil
	}
}

// sheetDeleteCheckPointFile delete a checkpoint file in dfs with fid and cid
func sheetDeleteCheckPointFile(fid uint, cid uint) (err error) {
	lockOnFid(fid)
	defer unlockOnFid(fid)
	chkpPath := gdocFS.GetCheckPointPath("sheet", fid, cid)
	if err := dao.Remove(chkpPath); err != nil {
		return err
	} else {
		return nil
	}
}

func sheetGetCheckPointNum(fid uint) (chkpNum int) {
	//lockOnFid(fid); defer unlockOnFid(fid)	// for accuracy, but would dramatically slow down in-cache performance
	path := gdocFS.GetCheckPointRootPath("sheet", fid)
	fileNames, err := dao.DirFilenameIndexesAllSorted(path)
	if err != nil {
		panic(err)
	}

	if len(fileNames) != 0 {
		chkpNum := fileNames[len(fileNames)-1]

		if len(fileNames) != chkpNum {
			logger.Errorf("[chkpNum(%d)] len(children) != latest checkpoint's name, use the latter", chkpNum)
		}

		return chkpNum
	} else {
		return 0
	}
}

// sheetGetPickledLogFromDfs pickles a Log from dfs with fid and lid
func sheetGetPickledLogFromDfs(fid uint, lid uint) (logs []gdocFS.SheetLogPickle, err error) {
	lockOnFid(fid)
	defer unlockOnFid(fid)
	path := gdocFS.GetLogPath("sheet", fid, lid)
	if fileRaw, err := dao.FileGetAll(path, true); err != nil {
		return nil, errors.WithStack(err)
	} else {
		logs, err = gdocFS.PickleSheetLogsFromContent(string(fileRaw))
		if err != nil {
			logger.Infof("Pickle Sheet Log Fails\n%s\n%v", string(fileRaw), fileRaw)
			return nil, errors.WithStack(err)
		}
		return logs, nil
	}
}

// sheetCreateLogFile create a empty Log in dfs with fid and lid
func sheetCreateLogFile(fid uint, lid uint) (err error) {
	lockOnFid(fid)
	defer unlockOnFid(fid)
	logPath := gdocFS.GetLogPath("sheet", fid, lid)
	if err := dao.FileCreate(logPath, 0); err != nil {
		return err
	} else {
		return nil
	}
}

// sheetDeleteLogFile delete a log file in dfs with fid and lid
func sheetDeleteLogFile(fid uint, lid uint) (err error) {
	lockOnFid(fid)
	defer unlockOnFid(fid)
	logPath := gdocFS.GetLogPath("sheet", fid, lid)
	if err := dao.Remove(logPath); err != nil {
		return err
	} else {
		return nil
	}
}

// chunkCreateDir create a empty chunk directory in dfs with fid
func chunkCreateDir(fid uint) (err error) {
	chunkRoot := gdocFS.GetChunkRootPath(fid)
	if err := dao.DirCreate(chunkRoot); err != nil {
		return err
	} else {
		return nil
	}
}

// reentrant lock on fid
var fidLockMap sync.Map

func lockOnFid(fid uint) {
	logger.Debugf("[fid(%d)\tgoId(%d)] Trying to Lock on fid!",
		fid, reentrantMutex.GetGoroutineId())
	actual, _ := fidLockMap.LoadOrStore(fid, reentrantMutex.NewReentrantMutex())
	mutex := actual.(*reentrantMutex.ReentrantMutex)
	mutex.Lock()
	logger.Debugf("[fid(%d)\tgoId(%d)\tcnt(%d)] Successfully Lock on fid!\n%s",
		fid, reentrantMutex.GetGoroutineId(), atomic.LoadInt32(&mutex.HoldCount), debug.Stack())
}

func unlockOnFid(fid uint) {
	logger.Debugf("[fid(%d)\tgoId(%d)] Trying to Unlock on fid!", fid, reentrantMutex.GetGoroutineId())
	if v, ok := fidLockMap.Load(fid); ok {
		mutex := v.(*reentrantMutex.ReentrantMutex)
		mutex.Unlock()
		logger.Debugf("[fid(%d)\tgoId(%d)\tcnt(%d)] Successfully Unlock on fid!\n%s",
			fid, reentrantMutex.GetGoroutineId(), atomic.LoadInt32(&mutex.HoldCount), debug.Stack())
	} else {
		logger.Errorf("[fid(%d)] Trying to unlock a fid without lock")
	}
}
