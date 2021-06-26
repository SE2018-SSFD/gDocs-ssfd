package master

import (
	"encoding/gob"
	"github.com/sirupsen/logrus"
	"os"
	"path"
)

type MasterCKP struct {
	// manage the state of chunkserver node
	servers []serialChunkServerStates
	cs SerialChunkStates
	namespace []SerialTreeNode
}
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
		logrus.Panic("set checkpoint error\n")
		return err
	}
	//TODO : truncate the log

	return nil
}