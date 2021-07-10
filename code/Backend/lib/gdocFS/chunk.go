package gdocFS

import "strconv"

func GetChunkRootPath(fid uint) (path string) {
	return "/chunk/" + strconv.FormatUint(uint64(fid), 10)
}

func GetChunkPath(fid uint, name string) (path string) {
	return "/chunk/" + strconv.FormatUint(uint64(fid), 10) + "/" + name
}
