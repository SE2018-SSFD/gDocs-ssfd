package dao

import (
	"backend/dfs"
	"backend/utils"
	"backend/utils/logger"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func writeAll(fd int, off int64, data []byte) (err error) {
	written, total := int64(0), int64(len(data))
	for written < total {
		n, err := dfs.Write(fd, off, data[written:])
		if err != nil {
			return err
		}
		written += n
		off += n
	}

	if written != total {
		return fmt.Errorf("expect to write %d bytes, actually it is %d", total, written)
	}

	return nil
}

func appendAll(fd int, data []byte) (err error) {
	appended, total := int64(0), int64(len(data))
	for appended < total {
		n, err := dfs.Append(fd, data[appended:])
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
		err = writeAll(fd, 0, utils.Zeros(initSize))
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

func FileGetAll(path string, filterPad ...bool) (data []byte, err error) {
	fileInfo, err := dfs.Stat(path)
	if err != nil {
		return nil, err
	}

	if fileInfo.IsDir {
		return nil, errors.New("cannot get all the content of a directory")
	}

	data, err = dfs.ReadAll(path, filterPad...)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func FileAppend(path string, data []byte) (err error) {
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

	err = appendAll(fd, data)
	if err != nil {
		return err
	}

	err = dfs.Close(fd)
	return err
}

func FileOverwriteAll(path string, data []byte) error {
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

	err = writeAll(fd, 0, data)
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

	for i := 0; i < len(fileInfos); i += 1 {
		filenames = append(filenames, fileInfos[i].Name)
	}

	return filenames, err
}

// DirFilenameIndexesAllSorted returns indexes names of all files in the directory in increasing order
func DirFilenameIndexesAllSorted(path string) (indexes []int, err error) {
	filenames, err := DirFileNamesAll(path)
	if err != nil {
		return nil, err
	}

	for _, filename := range filenames {
		if index, err := strconv.Atoi(strings.Split(filename, ".")[0]); err == nil {
			indexes = append(indexes, index)
		} else {
			logger.Errorf("[filename(%s)] Filename is not int!", filename)
		}
	}

	sort.Ints(indexes)
	return indexes, nil
}

func RemoveAll(path string) (err error) {
	return dfs.Delete(path)
}

func Remove(path string) (err error) {
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
