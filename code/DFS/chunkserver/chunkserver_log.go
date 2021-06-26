package chunkserver

import (
	"encoding/gob"
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/sirupsen/logrus"
)

func (cs *ChunkServer) AppendLog(ckl ChunkInfoLog) error {
	cs.logLock.Lock()
	defer cs.logLock.Unlock()
	filename := cs.GetLogName()
	fd, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		cs.Printf("\"AppendLog\" Cannot open file %s!\n", filename)
		return err
	}
	defer fd.Close()

	enc := json.NewEncoder(fd)
	err = enc.Encode(ckl)
	if err != nil {
		cs.Printf("Append log error\n")
		return err
	}
	return nil
}

func (cs *ChunkServer) StoreCheckPoint() error {
	cs.RLock()
	defer cs.RUnlock()

	filename := cs.GetCPName()
	fd, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer fd.Close()

	var ckcp []ChunkInfoCP
	for h, ck := range cs.chunks {
		ckcp = append(ckcp, ChunkInfoCP{Handle: h, VerNum: ck.verNum})
	}
	enc := gob.NewEncoder(fd)
	err = enc.Encode(ckcp)
	if err != nil {
		logrus.Print(err)
		logrus.Panic("set checkpoint error\n")
		return err
	}

	filename = cs.GetLogName()
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		logrus.Panic(err)
		return err
	}

	defer f.Close()
	err = f.Truncate(0) //clear log
	if err != nil {
		log.Panic(err)
		return err
	}
	cs.Printf("store checkpoint success\n")
	return nil
}

// call by RecoverChunkInfo
func (cs *ChunkServer) LoadCheckPoint() error {

	filename := cs.GetCPName()
	fd, err := os.Open(filename)
	if err != nil {
		// logrus.Printf("chunkserver %v: open file error\n")
		return err
	}
	defer fd.Close()
	var ckcps []ChunkInfoCP
	dec := gob.NewDecoder(fd)
	err = dec.Decode(&ckcps)
	if err == io.EOF {
		return nil
	} else if err != nil {
		return err
	}
	for _, ckcp := range ckcps {
		cs.chunks[ckcp.Handle] = &ChunkInfo{
			verNum:  ckcp.VerNum,
			isStale: false,
		}
	}

	cs.Printf("load checkpoint success\n")
	return nil
}

func (cs *ChunkServer) LoadLog() error {
	filename := cs.GetLogName()
	fd, err := os.Open(filename)
	if err != nil {
		// logrus.Printf("chunkserver %v: open file error\n")
		return err
	}
	defer fd.Close()

	dec := json.NewDecoder(fd)
	for dec.More() {
		var ckl ChunkInfoLog
		err = dec.Decode(&ckl)
		if err != nil {
			logrus.Print(err)
			return err
		}
		if ckl.Operation == Operation_Delete {
			delete(cs.chunks, ckl.Handle)
		} else {
			cs.chunks[ckl.Handle] = &ChunkInfo{
				verNum:  ckl.VerNum,
				isStale: false,
			}
		}
	}

	cs.Printf("load log success\n")
	return nil
}

// no need to get lock
func (cs *ChunkServer) RecoverChunkInfo() error {
	// cs.Lock()
	// defer cs.Unlock()

	err := cs.LoadCheckPoint()
	if err != nil {
		logrus.Print(err)
		// return err
	}
	err = cs.LoadLog()
	if err != nil {
		logrus.Print(err)
	}
	return err
}
