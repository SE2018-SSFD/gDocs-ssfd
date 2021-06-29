package gdocFS

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
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

func PickleSheetLogsFromContent(content string) (logs []SheetLogPickle, err error) {
	reader := strings.NewReader(content)
	decoder := json.NewDecoder(reader)

	pickled := SheetLogPickle{}
	for decoder.More() {
		if err := decoder.Decode(&pickled); err == nil {
			logs = append(logs, pickled)
		} else {
			return nil, err
		}
	}

	if reader.Len() != 0 {
		return nil, errors.New("remained content that cannot be pickled")
	}

	return logs, nil
}