package gdocFS

import "strconv"

func GetCheckPointPath(fileType string, fid uint, cid uint) string {
	return "/" + fileType + "/" + strconv.FormatUint(uint64(fid), 10) + "/checkpoint/" +
		strconv.FormatUint(uint64(cid), 10) + ".txt"
}
