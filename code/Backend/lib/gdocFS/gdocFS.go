package gdocFS

import "strconv"

func GetRootPath(fileType string, fid uint) (path string) {
	return "/" + fileType + "/" + strconv.FormatUint(uint64(fid), 10)
}

