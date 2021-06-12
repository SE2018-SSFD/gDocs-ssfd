package dfs

func Open(path string) (int, error) {
	return mockOpen(path)
}

func Create(path string) (int, error) {
	return mockCreate(path)
}

func Delete(path string) error {
	return mockDelete(path)
}

func Read(fd int, off int64, length int64) (string, error) {
	return mockRead(fd, off, length)
}

func Write(fd int, off int64, content string) (int64, error) {
	return mockWrite(fd, off, content)
}

func List(path string) ([]string, error) {
	return mockList(path)
}

type FileInfo struct {
	IsDir		bool
	Size		int64
}

func Stat(path string) (FileInfo, error) {
	return mockStat(path)
}