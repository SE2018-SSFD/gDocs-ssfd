package master

import (
	"DFS/util"
	"fmt"
	"path"

	"github.com/sirupsen/logrus"
)

// CreateRPC is called by client to create a new file
func (m *Master) CreateRPC(args util.CreateArg, reply *util.CreateRet) error {
	logrus.Debugf("RPC create, File Path : %s", args.Path)

	// Write ahead log
	err := m.AppendLog(MasterLog{OpType: util.CREATEOPS, Path: args.Path})
	if err != nil {
		logrus.Warnf("RPC delete failed : %s", err)
		return err
	}

	// Modified metadata
	err = m.ns.Mknod(args.Path, false)
	if err != nil {
		logrus.Warnf("RPC create failed : %s\n", err)
		return err
	}
	err = m.cs.NewFile(args.Path)
	if err != nil {
		logrus.Warnf("RPC create failed : %s", err)
		return err
	}
	return nil
}

// MkdirRPC is called by client to create a new dir
func (m *Master) MkdirRPC(args util.MkdirArg, reply *util.MkdirRet) error {
	logrus.Debugf("RPC mkdir, Dir Path : %s", args.Path)

	// Write ahead log
	err := m.AppendLog(MasterLog{OpType: util.MKDIROPS, Path: args.Path})
	if err != nil {
		logrus.Warnf("RPC mkdir failed : %s", err)
		return err
	}

	// Modified metadata
	err = m.ns.Mknod(args.Path, true)
	if err != nil {
		logrus.Warnf("RPC mkdir failed : %s", err)
		return err
	}
	return nil
}

// DeleteRPC is called by client to lazily delete a dir or file
func (m *Master) DeleteRPC(args util.DeleteArg, reply *util.DeleteRet) error {
	logrus.Debugf("RPC delete, Dir Path : %s", args.Path)

	// Write ahead log
	err := m.AppendLog(MasterLog{OpType: util.DELETEOPS, Path: args.Path})
	if err != nil {
		logrus.Warnf("RPC delete failed : %s", err)
		return err
	}

	// Modified metadata
	err = m.cs.Delete(args.Path)
	if err != nil {
		logrus.Warnf("RPC delete failed : %s", err)
		return err
	}
	err = m.ns.Delete(args.Path)
	if err != nil {
		logrus.Warnf("RPC delete failed : %s", err)
		return err
	}
	return nil
}

// ListRPC is called by client to list content of a dir
func (m *Master) ListRPC(args util.ListArg, reply *util.ListRet) (err error) {
	logrus.Debugf("RPC list, Dir Path : %s", args.Path)
	reply.Files, err = m.ns.List(args.Path)
	if err != nil {
		logrus.Warnf("RPC list failed : %s", err)
	}
	return err
}

// ScanRPC is called by client to scan all file info of a dir
func (m *Master) ScanRPC(args util.ScanArg, reply *util.ScanRet) (err error) {
	logrus.Debugf("RPC list, Dir Path : %s", args.Path)
	files, err := m.ns.List(args.Path)
	if err != nil {
		logrus.Warnf("RPC list failed : %s", err)
	}
	reply.FileInfos = make([]util.GetFileMetaRet, 0)
	for _, file := range files {
		var ret util.GetFileMetaRet
		fullPath := path.Join(string(args.Path), file)
		err := m.GetFileMetaRPC(util.GetFileMetaArg{Path: util.DFSPath(fullPath)}, &ret)
		logrus.Info("scan:", ret)
		if err != nil {
			return err
		}
		reply.FileInfos = append(reply.FileInfos, ret)
	}
	return err
}

// GetFileMetaRPC retrieve the file metadata by path
func (m *Master) GetFileMetaRPC(args util.GetFileMetaArg, reply *util.GetFileMetaRet) error {
	logrus.Debugf("RPC getFileMeta, File Path : %s", args.Path)
	node, err := m.ns.GetNode(args.Path)
	_, filename := path.Split(string(args.Path))

	if err != nil {
		logrus.Warnf("RPC getFileMeta failed : %s", err)
		*reply = util.GetFileMetaRet{
			Exist:    false,
			IsDir:    false,
			ChunkNum: 0,
			Filename: filename,
			// Size: -1,
		}
		return nil
	}
	reply.Exist = true
	reply.IsDir = node.isDir
	reply.Filename = filename
	if node.isDir {
		reply.ChunkNum = 0
	} else {
		reply.ChunkNum = m.cs.getChunkNum(args.Path)
	}
	return nil
}

// SetFileMetaRPC set the file metadata by path
// func (m *Master) SetFileMetaRPC(args util.SetFileMetaArg, reply *util.SetFileMetaRet) error {
// 	logrus.Debugf("RPC setFileMeta, File Path : %s", args.Path)

// 	// Write ahead log
// 	err := m.AppendLog(MasterLog{OpType: util.SETFILEMETAOPS, Path: args.Path, Size: args.Size})
// 	if err != nil {
// 		logrus.Warnf("RPC SetFileMeta failed : %s", err)
// 		return err
// 	}

// 	// Modified metadata
// 	m.cs.file[args.Path].Lock()
// 	defer m.cs.file[args.Path].Unlock()

// 	return nil
// }

// GetReplicasRPC get a chunk handle by file path and offset
// as well as the addresses of servers which store the chunk (and its replicas)
// if offset == -1,return the last one
// TODO : add lease
func (m *Master) GetReplicasRPC(args util.GetReplicasArg, reply *util.GetReplicasRet) (err error) {
	// Check if file exist
	logrus.Debugf("RPC getReplica, file path : %s, chunk index : %d", args.Path, args.ChunkIndex)
	m.cs.RLock()
	fs, exist := m.cs.file[args.Path]
	if !exist {
		m.cs.RUnlock()
		err = fmt.Errorf("FileNotExistsError : the requested DFS path %s is non-existing", string(args.Path))
		return err
	}
	fs.Lock()
	m.cs.RUnlock()

	//if offset == -1 , return the last one
	if args.ChunkIndex == -1 {
		if len(fs.chunks) == 0 {
			args.ChunkIndex = 0
		} else {
			args.ChunkIndex = len(fs.chunks) - 1
		}
		reply.ChunkIndex = len(fs.chunks) - 1
	}

	// Find the target chunk, if not exists, create one
	// Note that ChunkIndex <= len(fs.chunks) should be checked by client
	var targetChunk *chunkState
	// when args.ChunkIndex == -1 at this stage, the append is requested on a empty file
	if args.ChunkIndex == len(fs.chunks) || args.ChunkIndex == -1 {
		// randomly choose servers to store chunk replica
		var addrs []util.Address
		//addrs, err = m.css.randomServers(util.REPLICATIONTIMES)
		addrs, err = m.css.balanceServers(util.REPLICATIONTIMES)
		if err != nil {
			return err
		}

		// Write ahead log
		// err = m.AppendLog(MasterLog{OpType: util.GETREPLICASOPS, Path: args.Path, Addrs: addrs})
		err = m.AppendLog(MasterLog{OpType: util.GETREPLICASOPS, Path: args.Path})

		if err != nil {
			logrus.Warnf("RPC GetReplicas append log failed : %s\n", err)
			return err
		}

		// enter the function with write lock of fs
		targetChunk, err = m.cs.CreateChunkAndReplica(fs, addrs)
		if err != nil {
			return err
		}

		for _, addr := range addrs {
			err = m.css.addChunk(addr, targetChunk.Handle)
			if err != nil {
				return err
			}
		}

		//m.css.xxx
	} else {
		targetChunk = fs.chunks[args.ChunkIndex]
		// maybe modify UpdateFlag
		targetChunk.Lock()
		defer targetChunk.Unlock()
		fs.Unlock()
	}
	logrus.Debugf("targetchunk handle :%v, Locations : %v ", targetChunk.Handle, targetChunk.Locations)
	// Get target servers which store the replicate
	reply.ChunkServerAddrs = make([]util.Address, 0)
	for _, addr := range targetChunk.Locations {
		reply.ChunkServerAddrs = append(reply.ChunkServerAddrs, addr)
	}
	if targetChunk.UpdateFlag == true {
		err = util.CallAll(reply.ChunkServerAddrs,"ChunkServer.UpdateVersionRPC",util.UpdateVersionArg{Version: targetChunk.Version,Handle: targetChunk.Handle})
		if err != nil {
			logrus.Error(err)
		}
		targetChunk.Version++
		targetChunk.UpdateFlag = false
	}
	reply.ChunkHandle = targetChunk.Handle
	reply.ChunkIndex = args.ChunkIndex
	return nil
}
