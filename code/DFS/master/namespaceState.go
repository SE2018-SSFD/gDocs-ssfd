package master

import (
	"DFS/util"
	"fmt"
	"github.com/sirupsen/logrus"
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
func parsePath(path util.DFSPath)(parent util.DFSPath,filename string,err error){
	pos := strings.LastIndexByte(string(path),'/')
	// Check invalid path
	if pos == -1 || path[0]!='/' || path[len(path)-1]=='/' {
		err = fmt.Errorf("InvalidPathError : the requested DFS path %s is invalid!\n",string(path))
		return
	}
	// Parent should be /a/b/
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
	parentNode := ns.GetDir(parent)
	if parentNode == nil{
		err = fmt.Errorf("ParentNotExistsError : the requested DFS path %s has non-existing parent!\n",string(path))
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

func (ns *NamespaceState) GetDir(path util.DFSPath) *treeNode{
	symbols := strings.FieldsFunc(string(path),func(c rune) bool {return c=='/'})
	curNode := ns.root
	for _,symbol := range symbols {
		logrus.Debugln(symbol)
		curNode,found := curNode.treeNodes[symbol]
		if !found || curNode.isDir == false{
			return nil
		}
	}
	return curNode
}