package gdocFS

import (
	"encoding/json"
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

func GetCheckPointRootPath(fileType string, fid uint) (path string) {
	return "/" + fileType + "/" + strconv.FormatUint(uint64(fid), 10) + "/checkpoint"
}

func GetCheckPointPath(fileType string, fid uint, cid uint) string {
	return "/" + fileType + "/" + strconv.FormatUint(uint64(fid), 10) + "/checkpoint/" +
		strconv.FormatUint(uint64(cid), 10) + ".txt"
}

func PickleCheckPointFromContent(content string) (chkp SheetCheckPointPickle, err error) {
	if err = json.Unmarshal([]byte(content), &chkp); err != nil {
		return SheetCheckPointPickle{}, err
	} else {
		return chkp, nil
	}
}
