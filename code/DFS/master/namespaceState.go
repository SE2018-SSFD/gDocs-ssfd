package master

import (
	"DFS/util"
	"fmt"
	"strings"
)

/* The state of the file system namespace, maintained by the master */

type NamespaceState struct {
	root *treeNode
}

type treeNode struct {
	isDir bool
	// dir
	treeNodes map[string]*treeNode
	//file
	length int64
	chunks int64
}
func newNamespaceState()*NamespaceState{
	ns := &NamespaceState{
		root: &treeNode{
			isDir: true,
			treeNodes: make(map[string]*treeNode),
		},
	}
	return ns
}

// checkValidPath check if a DFS path is valid
func checkValidPath(path util.DFSPath) bool{
	if len(path)==0 || path[0]!='/' || path[len(path)-1]=='/'{
		return false
	}else{
		return true
	}
}

// parsePath is a helper method to parse a path string into parent and file
func parsePath(path util.DFSPath)(parent util.DFSPath,filename string,err error){
	// Check invalid path
	if !checkValidPath(path) {
		err = fmt.Errorf("InvalidPathError : the requested DFS path %s is invalid!\n",string(path))
		return
	}
	pos := strings.LastIndexByte(string(path),'/')
	parent = path[:pos]
	filename = string(path[pos+1:])
	return
}

// Mknod create a directory if isDir is true, else a file
func (ns *NamespaceState) Mknod(path util.DFSPath,isDir bool)error{
	parent,filename,err := parsePath(path)
	if err!=nil{
		return err
	}

	// Get parent node
	// TODO: lock the parents for concurrency control
	parentNode,err := ns.GetDir(parent)
	if err != nil{
		return err
	}

	// Create node
	if _, exist := parentNode.treeNodes[filename];exist{
		err = fmt.Errorf("FileExistsError : the requested DFS path %s has been created before!\n",string(path))
		return err
	}
	if isDir{
		parentNode.treeNodes[filename] = &treeNode{
			isDir: true,
			treeNodes: make(map[string]*treeNode),
		}
	}else{
		parentNode.treeNodes[filename] = nil
	}
	return nil
}

// List list all files in a given dir
func (ns *NamespaceState) List(path util.DFSPath)(files []string,err error){
	// Check invalid path
	if !checkValidPath(path) {
		err = fmt.Errorf("InvalidPathError : the requested DFS path %s is invalid!\n",string(path))
		return
	}
	// Get given dir
	dir,err := ns.GetDir(path)
	if err != nil{
		return nil,err
	}
	// List files
	files = make([]string,0)
	for file,_ := range dir.treeNodes{
		files = append(files,file)
	}
	return files,nil
}

// GetDir get a directory from DFS namespace
func (ns *NamespaceState) GetDir(path util.DFSPath)(*treeNode,error){
	symbols := strings.FieldsFunc(string(path),func(c rune) bool {return c=='/'})
	curNode := ns.root
	for _,symbol := range symbols {
		var found bool
		curNode,found = curNode.treeNodes[symbol]
		if !found || curNode.isDir == false{
			return nil,fmt.Errorf("ParentNotExistsError : the requested DFS path %s has non-existing parent!\n",string(path))
		}
	}
	return curNode,nil
}

// GetNode get a directory or file from DFS namespace
func (ns *NamespaceState) GetNode(path util.DFSPath)(*treeNode,error){
	symbols := strings.FieldsFunc(string(path),func(c rune) bool {return c=='/'})
	curNode := ns.root
	for _,symbol := range symbols {
		var found bool
		curNode,found = curNode.treeNodes[symbol]
		if !found{
			return nil,fmt.Errorf("FileNotExistsError : the requested DFS path %s is non-existing!\n",string(path))
		}
	}
	return curNode,nil
}
