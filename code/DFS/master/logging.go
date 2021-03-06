package master

import (
	"DFS/util"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
)

type MasterCKP struct {
	// manage the state of chunkserver node
	// Servers   []serialChunkServerStates
	Cs        SerialChunkStates
	Namespace []SerialTreeNode
}

// Warning : must be uppercase to be identified by JSON package
type MasterLog struct {
	OpType OperationType
	Path   util.DFSPath
	Handle util.Handle
	// Size   int            // for setFileMetaRPC
	// Addrs []util.Address // for GetReplicasRPC
	// Addr  util.Address   // for register & unregister RPC
}

func PushMessage(ap *sarama.AsyncProducer, value util.MasterLog) {
	//topic := "my_topic2"
	data, err := json.Marshal(value)
	if err != nil {
		logrus.Error("[kafka_producer][sendMessage]:%s", err.Error())
		return
	}

	msg := &sarama.ProducerMessage{
		Topic: util.MasterTopicName,
		Value: sarama.ByteEncoder(data),
	}

	(*ap).Input() <- msg
	select {
	case suc := <-(*ap).Successes():
		fmt.Printf("offset: %d,  timestamp: %s\n", suc.Offset, suc.Timestamp.String())
	case fail := <-(*ap).Errors():
		fmt.Printf("err: %s\n", fail.Err.Error())
	}
}

// AppendLog appends a log structure to persistent file
func (m *Master) AppendLog(ml MasterLog) error {
	m.logLock.Lock()
	defer m.logLock.Unlock()
	if m.ap == nil {
		logrus.Debug("producer is nil")
		return nil
	}
	uml := util.MasterLog{
		OpType: util.OperationType(ml.OpType),
		Path:   ml.Path,
		Handle: ml.Handle,
		// Size:   ml.Size,
		// Addrs: ml.Addrs,
		// Addr:  ml.Addr,
	}
	PushMessage(m.ap, uml)
	//logrus.Debugf("AppendLog : %d", ml.OpType)
	//filename := path.Join(string(m.metaPath), "log.dat")
	//fd, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	//if err != nil {
	//	logrus.Warnf("AppendLogError :%s\n", err)
	//	return err
	//}
	//defer fd.Close()
	//enc := json.NewEncoder(fd)
	//logrus.Infoln(ml)
	//err = enc.Encode(ml)
	//if err != nil {
	//	logrus.Warnf("AppendLogError :%s", err)
	//	return err
	//}
	//return nil
	return nil
}

// RecoverLog recovers a set of logs into master metadata
func (m *Master) RecoverLog() error {
	m.logLock.Lock()
	defer m.logLock.Unlock()

	filename := path.Join(string(m.metaPath), "log.dat")
	fd, err := os.Open(filename)
	if err != nil {
		// logrus.Printf("chunkserver %v: open file error\n")
		return nil
	}
	defer fd.Close()
	dec := json.NewDecoder(fd)

	for dec.More() {
		var log MasterLog
		err = dec.Decode(&log)
		if err != nil {
			logrus.Warnf("RecoverLog Failed : Decode failed : %s", err)
			return err
		}

		logrus.Debugf("RecoverLog : %d", log.OpType)
		switch log.OpType {
		case util.CREATEOPS:
			err = m.ns.Mknod(log.Path, false)
			if err != nil {
				logrus.Warnf("RPC create failed : %s\n", err)
				return err
			}
			err = m.cs.NewFile(log.Path)
			if err != nil {
				logrus.Warnf("RPC create failed : %s", err)
				return err
			}
		case util.MKDIROPS:
			err = m.ns.Mknod(log.Path, true)
			if err != nil {
				logrus.Warnf("RPC mkdir failed : %s", err)
				return err
			}
		case util.DELETEOPS:
			err = m.cs.Delete(log.Path)
			if err != nil {
				logrus.Warnf("RPC delete failed : %s", err)
				return err
			}
			err = m.ns.Delete(log.Path)
			if err != nil {
				logrus.Warnf("RPC delete failed : %s", err)
				return err
			}
		// case util.SETFILEMETAOPS:
		case util.GETREPLICASOPS:
			// when this operation is in the log, there must be new chunk created
			// increment handle
			newHandle := m.cs.handle.curHandle + 1
			m.cs.handle.curHandle += 1
			// add chunk to file
			newChunk := &chunkState{
				Locations: make([]util.Address, 0),
				Handle:    newHandle,
				Version: util.INITIALVERSION,
			}
			m.cs.chunk[newHandle] = newChunk
			fs, exist := m.cs.file[log.Path]
			if !exist {
				err := fmt.Errorf("FileNotExistsError : the requested DFS path %s is non-existing", string(log.Path))
				return err
			}
			fs.chunks = append(fs.chunks, newChunk)
			// for _, addr := range log.Addrs {
			// 	newChunk.Locations = append(newChunk.Locations, addr)
			// }
		case util.ADDVERSIONOPS:
			m.cs.chunk[log.Handle].UpdateFlag = false
			m.cs.chunk[log.Handle].Version++
		case util.SETUPDATEFLAGTOTRUEOPS:
			m.cs.chunk[log.Handle].UpdateFlag = true
		}

	}
	return nil
}

// LoadCheckPoint loads master metadata from disk
func (m *Master) LoadCheckPoint() error {
	filename := path.Join(string(m.metaPath), "checkpoint.dat")
	fd, err := os.Open(filename)
	if err != nil {
		logrus.Printf("LoadCheckPoint failed : open file %v error\n", filename)
		return err
	}
	var ckcp MasterCKP
	dec := gob.NewDecoder(fd)
	err = dec.Decode(&ckcp)
	if err != nil {
		logrus.Printf("LoadCheckPoint failed : decode error")
		return err
	}
	err = m.cs.Deserialize(ckcp.Cs)
	if err != nil {
		logrus.Printf("LoadCheckPoint failed : deserialize chunkStates error")
		return err
	}
	// err = m.css.Deserialize(ckcp.Servers)
	// if err != nil {
	// 	logrus.Printf("LoadCheckPoint failed : deserialize chunkserverStates error")
	// 	return err
	// }
	err = m.ns.Deserialize(ckcp.Namespace)
	if err != nil {
		logrus.Printf("LoadCheckPoint failed : deserialize namespace error")
		return err
	}
	logrus.Infoln("Master LoadCheckPoint success")
	return nil
}

// StoreCheckPoint store master metadata to disk
func (m *Master) StoreCheckPoint() error {
	m.RLock()
	defer m.RUnlock()

	filename := path.Join(string(m.metaPath), "checkpoint.dat")
	fd, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer fd.Close()
	ckcp := MasterCKP{
		// manage the state of chunkserver node
		// Servers:   m.css.Serialize(),
		Cs:        m.cs.Serialize(),
		Namespace: m.ns.Serialize(),
	}
	logrus.Debugln(ckcp)
	enc := gob.NewEncoder(fd)
	err = enc.Encode(ckcp)
	if err != nil {
		logrus.Warnf("StoreCheckPointError : %s\n", err)
		return err
	}
	//TODO : truncate the log
	filename = path.Join(string(m.metaPath), "log.dat")
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		logrus.Panic(err)
		return err
	}
	defer f.Close()
	err = f.Truncate(0) //clear log
	if err != nil {
		logrus.Warnf("StoreCheckPointError : %s", err)
		return err
	}
	return nil
}

func (m *Master) TryRecover() error {
	// Check if a checkpoint exists
	_, err := os.Stat(path.Join(string(m.metaPath), "checkpoint.dat"))
	if os.IsNotExist(err) {
		logrus.Infof("No checkpoint, start master directly")
		// return nil
	} else {
		logrus.Infof("Checkpoint found, start recover")
		err = m.LoadCheckPoint()
		if err != nil {
			logrus.Debug("load checkpoint err", err)
			return err
		}
	}

	err = m.RecoverLog()
	return err
}
