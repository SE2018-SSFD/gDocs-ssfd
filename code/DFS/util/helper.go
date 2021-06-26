package util

import (
	"fmt"
	"net/rpc"
	"strings"
)

// Call is RPC call helper
func Call(srv string, rpcname string, args interface{}, reply interface{}) error {
	c, errx := rpc.Dial("tcp", string(srv))
	if errx != nil {
		return errx
	}
	defer func(c *rpc.Client) {
		err := c.Close()
		if err != nil {
			//TODO:handle error
		}
	}(c)
	err := c.Call(rpcname, args, reply)
	return err
}

// CallAll applies the rpc call to all destinations.
func CallAll(dst []Address, rpcname string, args interface{}) error {
	ch := make(chan error)
	for _, d := range dst {
		go func(addr Address) {
			ch <- Call(string(addr), rpcname, args, nil)
		}(d)
	}
	errList := ""
	for _ = range dst {
		if err := <-ch; err != nil {
			errList += err.Error() + ";"
		}
	}

	if errList == "" {
		return nil
	} else {
		return fmt.Errorf(errList)
	}
}

// MakeString make ordered string from a-z repeating size times
func MakeString(size int) string {
	str := ""
	for i:=0;i<size;i++{
		str += string(rune('a' + i%26))
	}
	return str
}

// parsePath is a helper method to parse a path string into parent and file
func ParsePath(path DFSPath) (parent DFSPath, filename string, err error) {
	// Check invalid path
	if !CheckValidPath(path) {
		err = fmt.Errorf("InvalidPathError : the requested DFS path %s is invalid!\n", string(path))
		return
	}
	pos := strings.LastIndexByte(string(path), '/')
	parent = path[:pos]
	filename = string(path[pos+1:])
	return
}

// checkValidPath check if a DFS path is valid
func CheckValidPath(path DFSPath) bool {
	if len(path) == 0 || path[0] != '/' || path[len(path)-1] == '/' {
		return false
	} else {
		return true
	}
}
