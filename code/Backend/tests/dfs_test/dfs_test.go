package dfs_test

import (
	"backend/dfs"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"strconv"
	"strings"
	"testing"
)

const (
	basePath	=	"/test"
	testStr		=	"test"
)

func TestDfs(t *testing.T) {
	t.Log(dfs.Delete(basePath))

	var AllFd []int

	for i := 0; i < 10; i += 1 {
		parentName := strconv.Itoa(i)
		parentPath := basePath + "/" + parentName
		for j := 0; j < 20; j += 1 {
			childName := strconv.Itoa(j)
			allPath := parentPath + "/" + childName
			fd, err := dfs.Create(allPath)
			AllFd = append(AllFd, fd)
			if assert.NoError(t, err) {
				fileInfo, err := dfs.Stat(allPath)
				if assert.NoError(t, err) {
					assert.Equal(t, childName, fileInfo.Name)
					assert.Equal(t, false, fileInfo.IsDir)
				}
			}
		}
		fileInfos, err := dfs.Scan(parentPath)
		if assert.NoError(t, err) {
			var actual, expect []string

			for j, info := range fileInfos {
				expect = append(expect, strconv.Itoa(j))
				actual = append(actual, info.Name)
				assert.Equal(t, false, info.IsDir)
			}
			assert.ElementsMatch(t, expect, actual)
		}
	}

	fileInfos, err := dfs.Scan(basePath)
	if assert.NoError(t, err) {
		var actual, expect []string
		for i, info := range fileInfos {
			expect = append(expect, strconv.Itoa(i))
			actual = append(actual, info.Name)
			assert.Equal(t, true, info.IsDir)
		}
		assert.ElementsMatch(t, expect, actual)
	}

	for i := 0; i < 10; i += 1 {
		for j := 0; j < 10; j += 1 {
			path := basePath + "/" + strconv.Itoa(i) + "/" + strconv.Itoa(j)
			fd, err := dfs.Open(path, false)
			AllFd = append(AllFd, fd)
			if assert.NoError(t, err) {
				repeat := rand.Int() % 10 + 1
				toWrite := strings.Repeat(testStr, repeat)
				bytesWritten, err := dfs.Write(fd, 0, toWrite)
				if assert.NoError(t, err) {
					assert.EqualValues(t, len(testStr)*repeat, bytesWritten)
					content, err := dfs.ReadAll(path)
					if assert.NoError(t, err) {
						assert.Equal(t, toWrite, content)
					}
				}
			}
		}

		for j := 10; j < 20; j += 1 {
			path := basePath + "/" + strconv.Itoa(i) + "/" + strconv.Itoa(j)
			fd, err := dfs.Open(path, false)
			AllFd = append(AllFd, fd)
			if assert.NoError(t, err) {
				repeat := rand.Int() % 10 + 1
				toAppend := strings.Repeat(testStr, repeat)
				bytesWritten, err := dfs.Append(fd, toAppend)
				if assert.NoError(t, err) {
					assert.EqualValues(t, len(testStr)*repeat, bytesWritten)
					content, err := dfs.ReadAll(path)
					if assert.NoError(t, err) {
						assert.Equal(t, toAppend, content)
					}
				}
			}
		}
	}

	t.Log(AllFd)
	for _, fd := range AllFd {
		err := dfs.Close(fd)
		assert.NoError(t, err)
	}
}
