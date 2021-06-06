package dao

import (
	"backend/dfs"
	"backend/utils"
	"fmt"
)


// TODO: Find a Way to Get the Real Size of Cell to Fix Bugs of FileInsert etc.
func writeAll(fd int, off int64, content string) error {
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

func FileCreate(path string) error {
	fd, err := dfs.Create(path)
	if err != nil {
		return err
	}

	err = writeAll(fd, 0, string(utils.Zeros(utils.CELL_SIZE)))

	return err
}

func FileInsert(path string, off int64, content string, maxsize int64) error {
	toWrite := int64(len(content))
	if off + toWrite > maxsize {
		toWrite = maxsize - off
	}

	fd, err := dfs.Open(path)
	if err != nil {
		return err
	}

	block, err := dfs.Read(fd, off, toWrite)
	if err != nil {
		return err
	}

	block = block[:off] + content + block[off:]
	err = writeAll(fd, off, block[:toWrite])

	return err
}

func FileDelete(path string, off int64, length int64, maxsize int64) error {
	fd, err := dfs.Open(path)
	if err != nil {
		return err
	}

	block, err := dfs.Read(fd, off, length)
	if err != nil {
		return err
	}

	if off + length > maxsize {
		length = maxsize - off
	}
	block = block[:off] + block[off + length:]
	err = writeAll(fd, off, block)

	return err
}

func FileOverwrite(path string, off int64, content string, maxsize int64) error {
	toWrite := int64(len(content))
	fd, err := dfs.Open(path)
	if err != nil {
		return err
	}

	block, err := dfs.Read(fd, off, toWrite)
	if err != nil {
		return err
	}

	originSize := int64(len(block))
	block = block[:off] + content + block[off:]
	if toWrite + originSize > maxsize {
		toWrite = maxsize - originSize
	}
	err = writeAll(fd, off, block[:toWrite])

	return err
}