package chunkserver

import (
	"DFS/util"
	"log"
	"os"

	"github.com/sirupsen/logrus"
)

func (cs *ChunkServer) GetChunk(handle util.Handle, off int, buf []byte) (int, error) {
	filename := cs.GetFileName(handle)

	fd, err := os.Open(filename)

	if err != nil {
		return 0, err
	}
	defer fd.Close()

	return fd.ReadAt(buf, int64(off))
}

func (cs *ChunkServer) SetChunk(handle util.Handle, off int, buf []byte) (int, error) {

	if off+len(buf) > util.MAXCHUNKSIZE {
		log.Panic("chunk size cannot be larger than maxchunksize\n")
	}

	filename := cs.GetFileName(handle)

	fd, err := os.OpenFile(filename, os.O_WRONLY, 0644)

	if err != nil {
		return 0, err
	}

	defer fd.Close()

	return fd.WriteAt(buf, int64(off))
}

func (cs *ChunkServer) RemoveChunk(handle util.Handle) error {
	filename := cs.GetFileName(handle)
	err := os.Remove(filename)
	if err != nil {
		log.Panic(err)
		return err
	}
	return nil
}

func (cs *ChunkServer) CreateChunk(handle util.Handle) error {
	filename := cs.GetFileName(handle)
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		logrus.Panic(err)
		return err
	}

	defer f.Close()
	err = f.Truncate(util.MAXCHUNKSIZE)
	if err != nil {
		logrus.Panic(err)
		return err
	}
	return nil
}
