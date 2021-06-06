package master

import "DFS/util"

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

func (*NamespaceState) Mknod(path util.DFSPath,isDir bool)error{

	return nil
}