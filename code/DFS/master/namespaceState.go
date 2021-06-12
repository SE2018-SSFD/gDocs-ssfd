package master

import (
	"DFS/util"
	"fmt"
	"strings"
	"sync"
)

/* The state of the file system namespace, maintained by the master */

type NamespaceState struct {
	root *treeNode
}

type treeNode struct {
	sync.RWMutex
	isDir bool
	// dir
	treeNodes map[string]*treeNode
	//file
	length int64
	chunks int64
}

func newNamespaceState() *NamespaceState {
	ns := &NamespaceState{
		root: &treeNode{
			isDir:     true,
			treeNodes: make(map[string]*treeNode),
		},
	}
	return ns
}

// checkValidPath check if a DFS path is valid
func checkValidPath(path util.DFSPath) bool {
	if len(path) == 0 || path[0] != '/' || path[len(path)-1] == '/' {
		return false
	} else {
		return true
	}
}

// parsePath is a helper method to parse a path string into parent and file
func parsePath(path util.DFSPath) (parent util.DFSPath, filename string, err error) {
	// Check invalid path
	if !checkValidPath(path) {
		err = fmt.Errorf("InvalidPathError : the requested DFS path %s is invalid!\n", string(path))
		return
	}
	pos := strings.LastIndexByte(string(path), '/')
	parent = path[:pos]
	filename = string(path[pos+1:])
	return
}

// Mknod create a directory if isDir is true, else a file
func (ns *NamespaceState) Mknod(path util.DFSPath, isDir bool) error {
	parent, filename, err := parsePath(path)
	if err != nil {
		return err
	}

	// Get parent node
	// TODO: lock the parents for concurrency control
	parentNode, err := ns.GetDir(parent,true)
	defer ns.GetDir(parent,false)
	if err != nil {
		return err
	}
	parentNode.Lock()
	defer parentNode.Unlock()

	// Create node
	if _, exist := parentNode.treeNodes[filename]; exist {
		err = fmt.Errorf("FileExistsError : the requested DFS path %s has been created before!\n", string(path))
		return err
	}
	parentNode.treeNodes[filename] = &treeNode{
		isDir:     isDir,
		treeNodes: make(map[string]*treeNode),
	}
	return nil
}

// List list all files in a given dir
func (ns *NamespaceState) List(path util.DFSPath) (files []string, err error) {
	// Check invalid path
	if !checkValidPath(path) {
		err = fmt.Errorf("InvalidPathError : the requested DFS path %s is invalid\n", string(path))
		return
	}
	// Get given dir and RLock
	dir, err := ns.GetDir(path,true)
	defer ns.GetDir(path,false)
	if err != nil {
		return nil, err
	}
	if dir.isDir == false{
		return nil,fmt.Errorf("ListError : the requested DFS path %s is not a directory\n",string(path))
	}
	dir.RLock()
	defer dir.RUnlock()
	// List files
	files = make([]string, 0)
	for file, _ := range dir.treeNodes {
		files = append(files, file)
	}
	return files, nil
}

// Delete delete a file or empty directory from DFS namespace
func (ns *NamespaceState) Delete(path util.DFSPath) (err error) {
	// Check path
	parent, filename, err := parsePath(path)
	if err != nil {
		return err
	}
	if path == "/"{
		err = fmt.Errorf("DeleteError : Cannot delete root directory\n")
		return
	}

	// Get parent node and WLock
	parentNode, err := ns.GetDir(parent,true)
	defer ns.GetDir(parent,false)
	if err != nil {
		return err
	}
	parentNode.Lock()
	defer parentNode.Unlock()

	// Delete node (in namespace only)
	node, exist := parentNode.treeNodes[filename]
	if !exist {
		err = fmt.Errorf("DeleteError : the requested DFS path %s is not exist\n", string(path))
		return err
	}
	if node.isDir == true && len(node.treeNodes)!=0{
		err = fmt.Errorf("DeleteError : the requested DFS dir %s is not empty\n", string(path))
		return err
	}
	delete(parentNode.treeNodes,filename)
	return nil
}

// GetDir get a directory from DFS namespace
// if lock is true, then all intermediate dir will be RLocked,
// if lock is false, then they will be RUnlocked
func (ns *NamespaceState) GetDir(path util.DFSPath,lock bool) (*treeNode, error) {
	symbols := strings.FieldsFunc(string(path), func(c rune) bool { return c == '/' })
	curNode := ns.root
	for _, symbol := range symbols {
		if lock==true{
			curNode.RLock()
		}else if lock==false{
			curNode.RUnlock()
		}
		var found bool
		curNode, found = curNode.treeNodes[symbol]
		if !found || curNode.isDir == false {
			return nil, fmt.Errorf("ParentNotExistsError : the requested DFS path %s has non-existing parent!\n", string(path))
		}
	}
	return curNode, nil
}

// GetNode get a directory or file from DFS namespace
func (ns *NamespaceState) GetNode(path util.DFSPath) (*treeNode, error) {
	symbols := strings.FieldsFunc(string(path), func(c rune) bool { return c == '/' })
	curNode := ns.root
	for _, symbol := range symbols {
		// Logrus.Debugln("symbol : ", symbol)
		var found bool
		curNode, found = curNode.treeNodes[symbol]
		if !found {
			return nil, fmt.Errorf("FileNotExistsError : the requested DFS path %s is non-existing!\n", string(path))
		}
	}
	return curNode, nil
}
