package dfs

func Open(path string) (fd int, err error) {
	return mockOpen(path)
}

func Close(fd int) (err error) {
	return mockClose(fd)
}

func Create(path string) (fd int, err error) {
	return mockCreate(path)
}

func Delete(path string) (err error) {
	return mockDelete(path)
}

func Read(fd int, off int64, length int64) (content string, err error) {
	return mockRead(fd, off, length)
}

func Write(fd int, off int64, content string) (n int64, err error) {
	return mockWrite(fd, off, content)
}

func Truncate(fd int, length int64) (err error) {
	return mockTruncate(fd, length)
}

func Scan(path string) ([]FileInfo, error) {
	return mockScan(path)
}

type FileInfo struct {
	Name		string
	IsDir		bool
	Size		int64
}

func Stat(path string) (FileInfo, error) {
	return mockStat(path)
}