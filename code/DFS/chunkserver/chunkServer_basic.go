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

func (cs *ChunkServer) PadChunk(handle util.Handle) error {

	// if off+len(buf) > util.MAXCHUNKSIZE {
	// log.Panic("chunk size cannot be larger than maxchunksize\n")
	// }

	filename := cs.GetFileName(handle)

	fd, err := os.OpenFile(filename, os.O_WRONLY, 0644)

	if err != nil {
		return err
	}

	defer fd.Close()
	fileinfo, err := fd.Stat()
	if err != nil {
		log.Panic(err)
		return err
	}
	if fileinfo.Size() == util.MAXCHUNKSIZE {
		return nil
	} else {
		buf := []byte{0}
		_, err = fd.WriteAt(buf, util.MAXCHUNKSIZE-1)
	}
	return err
}

func (cs *ChunkServer) CreateAndSetChunk(handle util.Handle, off int, buf []byte) (int, error) {

	if off+len(buf) > util.MAXCHUNKSIZE {
		log.Panic("chunk size cannot be larger than maxchunksize\n")
	}

	filename := cs.GetFileName(handle)

	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		logrus.Panic(err)
		return 0, err
	}

	defer f.Close()
	// err = f.Truncate(util.MAXCHUNKSIZE)
	err = f.Truncate(0)
	if err != nil {
		logrus.Panic(err)
		return 0, err
	}

	return f.WriteAt(buf, int64(off))
}

func (cs *ChunkServer) RemoveChunk(handle util.Handle) error {
	filename := cs.GetFileName(handle)
	err := os.Remove(filename)
	if err != nil {
		logrus.Error(err)
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
	// err = f.Truncate(util.MAXCHUNKSIZE)
	err = f.Truncate(0)
	if err != nil {
		logrus.Panic(err)
		return err
	}
	return nil
}

// off: offset before write,if len(buf) + filesize > MAXCHUNKSIZE, return MAXCHUNKSIZE
func (cs *ChunkServer) AppendChunk(handle util.Handle, buf []byte) (off int, err error) {
	// if off+len(buf) > util.MAXCHUNKSIZE {
	// 	log.Panic("chunk size cannot be larger than maxchunksize\n")
	// }
	filename := cs.GetFileName(handle)

	fd, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND, 0644)

	if err != nil {
		return 0, err
	}

	defer fd.Close()
	fileInfo, err := fd.Stat()
	//NO SPACE
	if fileInfo.Size()+int64(len(buf)) > util.MAXCHUNKSIZE {
		if fileInfo.Size() == util.MAXCHUNKSIZE {
			return util.MAXCHUNKSIZE, nil
		}
		offset := util.MAXCHUNKSIZE - 1
		data := []byte{0}
		f, err := os.OpenFile(filename, os.O_WRONLY, 0644)
		if err != nil {
			logrus.Panic(err)
			return 0, err
		}
		defer f.Close()
		_, err = f.WriteAt(data, int64(offset))
		return util.MAXCHUNKSIZE, err
	}

	writeByte, err := fd.Write(buf)
	if err != nil {
		logrus.Panic(err)
	}

	fileInfo, err2 := fd.Stat()
	if err2 != nil {
		err = err2
	}

	off = int(fileInfo.Size()) - writeByte
	return
}
