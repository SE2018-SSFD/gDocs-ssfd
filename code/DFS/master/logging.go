package master

import (
	"DFS/util"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"time"
)

type MasterCKP struct {
	// manage the state of chunkserver node
	servers []serialChunkServerStates
	cs SerialChunkStates
	namespace []SerialTreeNode
}
// Warning : must be uppercase to be identified by JSON package
type MasterLog struct{
	OpType OperationType
	Path util.DFSPath
	Size int // for setFileMetaRPC
	Addrs []util.Address // for GetReplicasRPC
}

// AppendLog appends a log structure to persistent file
func (m *Master) AppendLog(ml MasterLog)error{
	m.logLock.Lock()
	defer m.logLock.Unlock()
	logrus.Debugf("AppendLog : %d\n", ml.OpType)
	filename := path.Join(string(m.metaPath),"log.dat")
	fd, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		logrus.Warnf("AppendLogError :%s\n", err)
		return err
	}
	defer fd.Close()
	enc := json.NewEncoder(fd)
	logrus.Infoln(ml)
	err = enc.Encode(ml)
	if err != nil {
		logrus.Warnf("AppendLogError :%s\n",err)
		return err
	}
	return nil
}

// RecoverLog recovers a set of logs into master metadata
func (m *Master) RecoverLog() error {
	m.logLock.Lock()
	defer m.logLock.Unlock()
	var cret *util.CreateRet
	var mret *util.MkdirRet
	var dret *util.DeleteRet
	var sret *util.SetFileMetaRet

	filename := path.Join(string(m.metaPath),"log.dat")
	fd, err := os.Open(filename)
	if err != nil {
		// logrus.Printf("chunkserver %v: open file error\n")
		return err
	}
	defer fd.Close()
	dec := json.NewDecoder(fd)

	for dec.More(){
		var log MasterLog
		err = dec.Decode(&log)
		if err != nil {
			logrus.Warnf("RecoverLog Failed : Decode failed : %s",err)
			return err
		}

		logrus.Debugf("RecoverLog : %d\n", log.OpType)
		switch log.OpType {
		case util.CREATEOPS:
			err := m.CreateRPC(util.CreateArg{Path: log.Path},cret)
			if err != nil {
				logrus.Warnf("RecoverLog Failed : Create failed : %s\n", err)
				return err
			}
		case util.MKDIROPS:
			err := m.MkdirRPC(util.MkdirArg{Path: log.Path},mret)
			if err != nil {
				logrus.Warnf("RecoverLog Failed : Mkdir failed : %s\n", err)
				return err
			}
		case util.DELETEOPS:
			err := m.DeleteRPC(util.DeleteArg{Path: log.Path},dret)
			if err != nil {
				logrus.Warnf("RecoverLog Failed : Delete failed : %s\n", err)
				return err
			}
		case util.SETFILEMETAOPS:
			err := m.SetFileMetaRPC(util.SetFileMetaArg{Path: log.Path,Size:log.Size},sret)
			if err != nil {
				logrus.Warnf("RecoverLog Failed : Delete failed : %s\n", err)
				return err
			}
		case util.GETREPLICASOPS:
			// when this operation is in the log, there must be new chunk created
			// increment handle
			newHandle := m.cs.handle.curHandle+1
			m.cs.handle.curHandle+=1
			// add chunk to file
			newChunk := &chunkState{
				Locations: make([]util.Address,0),
				Handle: newHandle,
				expire: time.Now(),
			}
			fs, exist := m.cs.file[log.Path]
			if !exist {
				err := fmt.Errorf("FileNotExistsError : the requested DFS path %s is non-existing!\n", string(log.Path))
				return err
			}
			fs.chunks = append(fs.chunks,newChunk)
			for _ , addr := range log.Addrs{
				newChunk.Locations = append(newChunk.Locations,addr)
			}
		}

	}
	return nil
}

// LoadCheckPoint loads master metadata from disk
func (m *Master) LoadCheckPoint() error{
	filename := path.Join(string(m.metaPath),"checkpoint.dat")
	fd, err := os.Open(filename)
	if err != nil {
		logrus.Printf("LoadCheckPoint failed : open file %v error\n",filename)
		return err
	}
	var ckcp MasterCKP
	dec := gob.NewDecoder(fd)
	err = dec.Decode(&ckcp)
	if err != nil {
		logrus.Printf("LoadCheckPoint failed : decode error\n")
		return err
	}
	err = m.cs.Deserialize(ckcp.cs)
	if err != nil {
		logrus.Printf("LoadCheckPoint failed : deserialize chunkStates error\n")
		return err
	}
	err = m.css.Deserialize(ckcp.servers)
	if err != nil {
		logrus.Printf("LoadCheckPoint failed : deserialize chunkserverStates error\n")
		return err
	}
	err = m.ns.Deserialize(ckcp.namespace)
	if err != nil {
		logrus.Printf("LoadCheckPoint failed : deserialize namespace error\n")
		return err
	}
	logrus.Infoln("Master LoadCheckPoint success")
	return nil
}

// StoreCheckPoint store master metadata to disk
func (m *Master) StoreCheckPoint() error {
	m.RLock()
	defer m.RUnlock()

	filename := path.Join(string(m.metaPath),"checkpoint.dat")
	fd, err := os.OpenFile(filename, os.O_WRONLY |os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer fd.Close()
	ckcp := MasterCKP{
		// manage the state of chunkserver node
		servers:m.css.Serialize(),
		cs:m.cs.Serialize() ,
		namespace:m.ns.Serialize(),
	}
	enc := gob.NewEncoder(fd)
	err = enc.Encode(ckcp)
	if err != nil {
		logrus.Warnf("StoreCheckPointError : %s\n",err)
		return err
	}
	//TODO : truncate the log
	filename = path.Join(string(m.metaPath),"log.dat")
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		logrus.Panic(err)
		return err
	}
	defer f.Close()
	err = f.Truncate(0) //clear log
	if err != nil {
		logrus.Warnf("StoreCheckPointError : %s\n",err)
		return err
	}
	return nil
}

func (m *Master) TryRecover() error{
	// Check if a checkpoint exists
	_,err := os.Stat(path.Join(string(m.metaPath),"checkpoint.dat"))
	if os.IsNotExist(err){
		logrus.Infof("No checkpoint, start master directly\n")
		return nil
	}
	logrus.Infof("Checkpoint found, start recover\n")

	err = m.LoadCheckPoint()
	if err!=nil{
		return err
	}
	err = m.RecoverLog()
	return err
}