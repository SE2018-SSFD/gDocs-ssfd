package gdocFS

import "strconv"

func GetLogPath(fileType string, fid uint, lid uint) string {
	return "/" + fileType + "/" + strconv.FormatUint(uint64(fid), 10) + "/log/" +
		strconv.FormatUint(uint64(lid), 10) + ".txt"
}
