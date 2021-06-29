package gdocFS

import (
	"strconv"
	"time"
)

type SheetLogPickle struct {
	Lid			uint		`json:"lid"`
	Timestamp	time.Time	`json:"timestamp"`
	Row			int 		`json:"row"`
	Col			int			`json:"col"`
	Old			string		`json:"old"`
	New			string		`json:"new"`
	Uid			uint		`json:"uid"`
	Username	string		`json:"username"`
}

func GetLogRootPath(fileType string, fid uint) (path string) {
	return "/" + fileType + "/" + strconv.FormatUint(uint64(fid), 10) + "/log"
}

func GetLogPath(fileType string, fid uint, lid uint) string {
	return "/" + fileType + "/" + strconv.FormatUint(uint64(fid), 10) + "/log/" +
		strconv.FormatUint(uint64(lid), 10) + ".txt"
}