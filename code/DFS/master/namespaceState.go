package master

import (
	"DFS/util"
	"fmt"
	"strings"
	"sync"
)

/* The state of the file system namespace, maintained by the master */

type NamespaceState struct {
	root     *treeNode
	childCount int
}
type SerialTreeNode struct {
	IsDir    bool
	Children map[string]int
}
type treeNode struct {
	sync.RWMutex
	isDir bool
	// dir
	treeNodes map[string]*treeNode
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



// Mknod create a directory if isDir is true, else a file
func (ns *NamespaceState) Mknod(path util.DFSPath, isDir bool) error {
	parent, filename, err := util.ParsePath(path)
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
	if strings.HasPrefix(filename,util.DELETEPREFIX){
		err = fmt.Errorf("UnlawwedPrefixError : the requested DFS path %s has deleted prefix!\n", string(path))
		return err
	}
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
	if string(path)!="/" && !util.CheckValidPath(path) {
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
		if !strings.HasPrefix(file,util.DELETEPREFIX){ // ignore deleted bindings
			files = append(files, file)
		}
	}
	return files, nil
}

// Delete delete a file or empty directory from DFS namespace
func (ns *NamespaceState) Delete(path util.DFSPath) (err error) {
	// Check path
	parent, filename, err := util.ParsePath(path)
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
	parentNode.treeNodes[util.DELETEPREFIX+filename]  = node
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
// Serialize a namespace
func (ns *NamespaceState) Serialize()[]SerialTreeNode{
	// only need to lock the root here since any operation on child node must get the root lock beforehand
	ns.root.RLock()
	defer ns.root.RUnlock()
	nodes := make([]SerialTreeNode,0)
	ns.tree2array(&nodes,ns.root)
	return nodes
}

// tree2array transforms the namespace tree into an array for serialization
func (ns *NamespaceState) tree2array(array *[]SerialTreeNode, node *treeNode) int {
	n := SerialTreeNode{IsDir: node.isDir}
	if node.isDir {
		n.Children = make(map[string]int)
		for k, v := range node.treeNodes {
			n.Children[k] = ns.tree2array(array, v)
		}
	}

	*array = append(*array, n)
	ret := ns.childCount
	ns.childCount++
	return ret
}
// array2tree transforms the an serialized array to namespace tree
func (ns *NamespaceState) array2tree(array []SerialTreeNode, id int) *treeNode {
	n := &treeNode{
		isDir:  array[id].IsDir,
	}

	if array[id].IsDir {
		n.treeNodes = make(map[string]*treeNode)
		for k, v := range array[id].Children {
			n.treeNodes[k] = ns.array2tree(array, v)
		}
	}

	return n
}

// Deserializa the metadata from disk
func (ns *NamespaceState) Deserialize(array []SerialTreeNode) error {
	ns.root.Lock()
	defer ns.root.Unlock()
	ns.root = ns.array2tree(array, len(array)-1)
	return nil
}