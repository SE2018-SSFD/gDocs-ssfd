package dao

import (
	"backend/dfs"
	"backend/utils"
	"errors"
	"fmt"
	"sort"
)


func writeAll(fd int, off int64, content string) (err error) {
	toWrite := int64(len(content))
	for toWrite > 0 {
		n, err := dfs.Write(fd, off, content[:toWrite])
		if err != nil {
			return err
		}
		toWrite = toWrite - n
		off = off + n
	}
	
	if toWrite != 0 {
		return fmt.Errorf("expect to write %d bytes, actually it is %d", len(content),
			int64(len(content)) - toWrite)
	}

	return nil
}

func appendAll(fd int, content string) (err error) {
	appended := int64(0)
	total := int64(len(content))
	for appended < total {
		n, err := dfs.Append(fd, content[appended:])
		if err != nil {
			return err
		}
		appended += n
	}

	if appended != total {
		return fmt.Errorf("expect to append %d bytes, actually it is %d", total, appended)
	}

	return nil
}

func FileCreate(path string, initSize int64) (err error) {
	fd, err := dfs.Create(path)
	if err != nil {
		return err
	}

	if initSize != 0 {
		err = writeAll(fd, 0, string(utils.Zeros(initSize)))
		if err != nil {
			return err
		}
	}

	err = dfs.Close(fd)
	return err
}

func DirCreate(path string) (err error) {
	err = dfs.Mkdir(path)
	if err != nil {
		return err
	}

	return nil
}

func FileGetAll(path string) (content string, err error) {
	fileInfo, err := dfs.Stat(path)
	if err != nil {
		return "", err
	}

	if fileInfo.IsDir {
		return "", errors.New("cannot get all the content of a directory")
	}

	fd, err := dfs.Open(path, false)
	if err != nil {
		return "", err
	}

	content, err = dfs.ReadAll(fd)
	if err != nil {
		return "", err
	}

	return content, nil
}

func FileAppend(path string, content string) (err error) {
	fileInfo, err := dfs.Stat(path)
	if err != nil {
		return err
	}

	if fileInfo.IsDir {
		return errors.New("cannot append to a directory")
	}

	fd, err := dfs.Open(path, true)
	if err != nil {
		return err
	}

	err = appendAll(fd, content)
	if err != nil {
		return err
	}

	err = dfs.Close(fd)
	return err
}

func FileOverwriteAll(path string, content string) error {
	fileInfo, err := dfs.Stat(path)
	if err != nil {
		return err
	}

	if fileInfo.IsDir {
		return errors.New("cannot write a directory")
	}

	fd, err := dfs.Open(path, false)
	if err != nil {
		return err
	}

	err = writeAll(fd, 0, content)
	if err != nil {
		return err
	}

	err = dfs.Close(fd)
	return err
}



func DirFileNamesAll(path string) (filenames []string, err error) {
	fileInfos, err := dfs.Scan(path)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(filenames); i += 1 {
		filenames = append(filenames, fileInfos[i].Name)
	}

	return filenames, err
}

// DirFilenamesAllSorted returns names of all files in the directory in increasing order
func DirFilenamesAllSorted(path string) (filenames []string, err error) {
	filenames, err = DirFileNamesAll(path)
	if err != nil {
		return nil, err
	}

	sort.Strings(filenames)
	return filenames, nil
}

func RemoveAll(path string) (err error) {
	return dfs.Delete(path)
}

//func FileInsert(path string, off int64, content string, maxsize int64) error {
//	toWrite := int64(len(content))
//	if off + toWrite > maxsize {
//		toWrite = maxsize - off
//	}
//
//	fd, err := dfs.Open(path)
//	if err != nil {
//		return err
//	}
//
//	block, err := dfs.Read(fd, off, toWrite)
//	if err != nil {
//		return err
//	}
//
//	block = block[:off] + content + block[off:]
//	err = writeAll(fd, off, block[:toWrite])
//
//	return err
//}
//
//func FileDelete(path string, off int64, length int64, maxsize int64) error {
//	fd, err := dfs.Open(path)
//	if err != nil {
//		return err
//	}
//
//	block, err := dfs.Read(fd, off, length)
//	if err != nil {
//		return err
//	}
//
//	if off + length > maxsize {
//		length = maxsize - off
//	}
//	block = block[:off] + block[off + length:]
//	err = writeAll(fd, off, block)
//
//	return err
//}
//
//func FileOverwrite(path string, off int64, content string, maxsize int64) error {
//	toWrite := int64(len(content))
//	fd, err := dfs.Open(path)
//	if err != nil {
//		return err
//	}
//
//	block, err := dfs.Read(fd, off, toWrite)
//	if err != nil {
//		return err
//	}
//
//	originSize := int64(len(block))
//	block = block[:off] + content + block[off:]
//	if toWrite + originSize > maxsize {
//		toWrite = maxsize - originSize
//	}
//	err = writeAll(fd, off, block[:toWrite])
//
//	return err
//}