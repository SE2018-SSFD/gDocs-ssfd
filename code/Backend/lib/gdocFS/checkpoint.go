package gdocFS

import (
	"strconv"
	"time"
)

type SheetCheckPointPickle struct {
	Cid			uint				`json:"cid"`
	Timestamp	time.Time			`json:"timestamp"`
	Rows		int					`json:"rows"`
	Columns		int					`json:"columns"`
	Content		[]string			`json:"content"`
}

func GetCheckPointPath(fileType string, fid uint, cid uint) string {
	return "/" + fileType + "/" + strconv.FormatUint(uint64(fid), 10) + "/checkpoint/" +
		strconv.FormatUint(uint64(cid), 10) + ".txt"
}
