package client

import (
	"DFS/util"
	json "encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

type Client struct {
	clientAddr util.Address
	masterAddr util.Address
	fdTable    map[int]util.DFSPath
	s          *http.Server
	//TODO:add lease
}

// InitClient initClient init a new client and return.
func InitClient(clientAddr util.Address, masterAddr util.Address) *Client {
	c := &Client{
		clientAddr: clientAddr,
		masterAddr: masterAddr,
		fdTable:    make(map[int]util.DFSPath),
	}
	return c
}

func (c *Client) Serve() {
	mux := http.NewServeMux()
	mux.HandleFunc("/create", c.Create)
	mux.HandleFunc("/mkdir", c.Mkdir)
	mux.HandleFunc("/delete", c.Delete)
	mux.HandleFunc("/read", c.Read)
	mux.HandleFunc("/write", c.Write)
	mux.HandleFunc("/open", c.Open)
	mux.HandleFunc("/close", c.Close)
	mux.HandleFunc("/append", c.Append)
	mux.HandleFunc("/fileInfo", c.GetFileInfo)
	c.s = &http.Server{
		Addr:           util.CLIENTADDR,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err := c.s.ListenAndServe()
	if err != nil {
		logrus.Debug("Client server shutdown!\n")
	}
	//logrus.Fatalln("stop!")
}

// Exit Directly
func (c *Client) Exit() {
	err := c.s.Close()
	if err != nil {
		return
	}
}

// TODO:client should not return error due to DFS failure
// Create a file.
func (c *Client) Create(w http.ResponseWriter, r *http.Request) {
	var arg util.CreateArg
	var ret util.CreateRet
	err := json.NewDecoder(r.Body).Decode(&arg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = util.Call(string(c.masterAddr), "Master.CreateRPC", arg, &ret)
	if err != nil {
		logrus.Fatalln("CreateRPC failed:", err)
		return
	}
	io.WriteString(w, "0")
	return
}

// Mkdir a dir.
func (c *Client) Mkdir(w http.ResponseWriter, r *http.Request) {
	var arg util.MkdirArg
	var ret util.MkdirRet
	err := json.NewDecoder(r.Body).Decode(&arg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = util.Call(string(c.masterAddr), "Master.MkdirRPC", arg, &ret)
	if err != nil {
		logrus.Fatalln("MkdirRPC failed:", err)
		return
	}
	io.WriteString(w, "0")
	return
}

// Open a file.
// If fd is depleted, return -1
func (c *Client) Open(w http.ResponseWriter, r *http.Request) {
	var arg util.OpenArg
	var ret util.OpenRet
	err := json.NewDecoder(r.Body).Decode(&arg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	for i := 0; i < util.MAXFD; i++ {
		_, exist := c.fdTable[i]
		if !exist {
			logrus.Debugf("Client open : assign %d", i)
			c.fdTable[i] = arg.Path
			io.WriteString(w, strconv.Itoa(i))
			return
		}
	}
	w.WriteHeader(400)
	msg, _ := json.Marshal(ret)
	w.Write(msg)
}

// Close a file.
func (c *Client) Close(w http.ResponseWriter, r *http.Request) {
	var arg util.CloseArg
	err := json.NewDecoder(r.Body).Decode(&arg)
	defer func(err error) {
		logrus.Warn(err, "!\n")
		if err != nil {
			w.Write([]byte(err.Error()))
		}
	}(err)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, exist := c.fdTable[arg.Fd]
	if !exist {
		err = fmt.Errorf("FileClosedError : file has been closed\n")
		print(err.Error())
		http.Error(w, "Fd Nonexist", http.StatusBadRequest)
		return
	}
	logrus.Debugf("Client close : free %d", arg.Fd)
	delete(c.fdTable, arg.Fd)
	return
}

func (c *Client) _Read(path util.DFSPath, offset int, len int, fileSize int) (readBytes int, buf []byte, err error) {
	var argR util.GetReplicasArg
	var retR util.GetReplicasRet
	var argRCK util.ReadChunkArgs
	var retRCK util.ReadChunkReply
	argR.Path = path
	for readBytes < len {
		roundOff := (offset + readBytes) % util.MAXCHUNKSIZE
		roundReadBytes := int(math.Min(float64(util.MAXCHUNKSIZE-roundOff), float64(len-readBytes)))

		argR.ChunkIndex = (offset + readBytes) / util.MAXCHUNKSIZE
		err = util.Call(string(c.masterAddr), "Master.GetReplicasRPC", argR, &retR)
		if err != nil {
			return
		}
		logrus.Debugf(" ChunkHandle : %d Addresses : %s %s %s\n", retR.ChunkHandle, retR.ChunkServerAddrs[0], retR.ChunkServerAddrs[1], retR.ChunkServerAddrs[2])
		//TODO : make it random
		argRCK.Handle = retR.ChunkHandle
		argRCK.Len = roundReadBytes
		argRCK.Off = roundOff
		err = util.Call(string(retR.ChunkServerAddrs[0]), "ChunkServer.ReadChunkRPC", argRCK, &retRCK)
		if err != nil {
			logrus.Panicln("Client read failed :", err)
			return
		}
		if retRCK.Len != roundReadBytes {
			logrus.Panicln("Client should read %v,buf only read %v", roundReadBytes, retRCK.Len)
			return
		}
		buf = append(buf, retRCK.Buf...)
		readBytes += roundReadBytes
		logrus.Debugf(" Read %d bytes from chunkserver %s, bytes read %d\n", roundReadBytes, string(retR.ChunkServerAddrs[0]), readBytes)
	}
	return
}

// Read a file.
// should contact the master first, then get the data directly from chunkserver
func (c *Client) Read(w http.ResponseWriter, r *http.Request) {
	var argR util.ReadArg
	var retR util.ReadRet
	var argF util.GetFileMetaArg
	var retF util.GetFileMetaRet
	// Decode the params
	err := json.NewDecoder(r.Body).Decode(&argR)
	if err != nil {
		logrus.Fatalln("Client read failed :", err)
		w.WriteHeader(400)
		return
	}

	// Get the file metadata and check
	path := c.fdTable[argR.Fd]
	if path == "" {
		err = fmt.Errorf("Client read failed : fd %d is not opened\n", argR.Fd)
		return
	}

	argF.Path = path
	err = util.Call(string(c.masterAddr), "Master.GetFileMetaRPC", argF, &retF)
	if !retF.Exist {
		logrus.Fatalln("Client read failed :", err)
		return
	}
	fileSize := retF.Size
	if argR.Offset+argR.Len > fileSize {
		err = fmt.Errorf("Client read failed : read offset + read len exceed file size\n")
		return
	}

	// Read to chunk
	readBytes, buf, err := c._Read(path, argR.Offset, argR.Len, fileSize)
	if err != nil {
		logrus.Fatalln("Client read failed :", err)
		w.WriteHeader(400)
		return
	}
	retR.Data = buf
	retR.Len = readBytes
	msg, _ := json.Marshal(retR)
	w.Write(msg)
	return

}

// Delete a file.
func (c *Client) Delete(w http.ResponseWriter, r *http.Request) {
	var argR util.GetReplicasArg
	var retR util.GetReplicasRet
	var argF util.GetFileMetaArg
	var retF util.GetFileMetaRet
	var argD util.DeleteArg
	var retD util.DeleteRet

	// Decode the params
	err := json.NewDecoder(r.Body).Decode(&argD)
	if err != nil {
		logrus.Fatalln("Client delete failed :", err)
		w.WriteHeader(400)
		return
	}

	// Get the file metadata and check
	argF.Path = argD.Path
	err = util.Call(string(c.masterAddr), "Master.GetFileMetaRPC", argF, &retF)
	if !retF.Exist {
		logrus.Fatalln("Client delete failed :", err)
		w.WriteHeader(400)
		return
	}

	// Delete the master metadata first
	err = util.Call(string(c.masterAddr), "Master.DeleteRPC", argD, &retD)
	if err != nil {
		logrus.Fatalln("Client delete failed :", err)
		w.WriteHeader(400)
		return
	}

	// Delete the chunk one by one
	// By default, the first entry int retR.ChunkServerAddr is the primary
	chunkIndex := 0
	for chunkIndex*util.MAXCHUNKSIZE < retF.Size {
		argR.ChunkIndex = chunkIndex
		err = util.Call(string(c.masterAddr), "Master.GetReplicasRPC", argR, &retR)
		if err != nil {
			logrus.Fatalln("Client delete failed :", err)
			w.WriteHeader(400)
			return
		}
		// TODO delete file in chunkServer
		chunkIndex += 1
	}
	w.WriteHeader(200)
	return
}

// Append to a file
func (c *Client) Append(w http.ResponseWriter, r *http.Request) {
	var argA util.AppendArg
	var argF util.GetFileMetaArg
	var retF util.GetFileMetaRet

	// Decode the params
	err := json.NewDecoder(r.Body).Decode(&argA)
	if err != nil {
		logrus.Fatalln("Client append failed :", err)
		w.WriteHeader(400)
		return
	}

	// Get the file metadata and check
	path := c.fdTable[argA.Fd]
	if path == "" {
		err = fmt.Errorf("Client write failed : fd %d is not opened\n", argA.Fd)
		return
	}
	argF.Path = path
	err = util.Call(string(c.masterAddr), "Master.GetFileMetaRPC", argF, &retF)
	if !retF.Exist {
		logrus.Fatalln("Client write failed :", err)
		return
	}
	fileSize := retF.Size

	// Write to file
	writtenBytes, err := c._Write(path, fileSize, argA.Data, fileSize)
	if err != nil {
		logrus.Fatalln("Client write failed :", err)
		w.WriteHeader(400)
		return
	}
	msg, _ := json.Marshal(writtenBytes)
	w.Write(msg)
	w.WriteHeader(200)
	return

}

func (c *Client) _Write(path util.DFSPath, offset int, data []byte, fileSize int) (writtenBytes int, err error) {
	var argR util.GetReplicasArg
	var retR util.GetReplicasRet
	var argL util.LoadDataArgs
	var retL util.LoadDataReply
	var argS util.SetFileMetaArg
	var retS util.SetFileMetaRet
	var argC util.SyncArgs
	var retC util.SyncReply

	// Write the chunk (may add chunks)
	// By default, the first entry int retR.ChunkServerAddr is the primary
	argR.Path = path
	for writtenBytes < len(data) {
		argR.ChunkIndex = (offset + writtenBytes) / util.MAXCHUNKSIZE
		err = util.Call(string(c.masterAddr), "Master.GetReplicasRPC", argR, &retR)
		if err != nil {
			return
		}
		logrus.Debugf(" ChunkHandle : %d Addresses : %s %s %s\n", retR.ChunkHandle, retR.ChunkServerAddrs[0], retR.ChunkServerAddrs[1], retR.ChunkServerAddrs[2])
		roundWrittenBytes := int(math.Min(float64(util.MAXCHUNKSIZE-(offset+writtenBytes)%util.MAXCHUNKSIZE), float64(len(data)-writtenBytes)))
		var cid = util.CacheID{
			Handle:     retR.ChunkHandle,
			ClientAddr: c.clientAddr,
		}
		argL.CID = cid
		argL.Data = data[writtenBytes:(writtenBytes + roundWrittenBytes)]
		argL.Addrs = retR.ChunkServerAddrs
		//TODO: make it random
		//argL.Addrs = make([]util.Address,0)
		//for _,index := range rand.Perm(len(retR.ChunkServerAddrs)){
		//	argL.Addrs = append(argL.Addrs,retR.ChunkServerAddrs[index])
		//}
		// Send to Master now
		err = util.Call(string(argL.Addrs[0]), "ChunkServer.LoadDataRPC", argL, &retL)
		if err != nil {
			logrus.Fatalln("Client write failed :", err)
			return
		}
		argC = util.SyncArgs{
			CID:   cid,
			Off:   (offset + writtenBytes) % util.MAXCHUNKSIZE,
			Addrs: retR.ChunkServerAddrs[1:],
		}
		err = util.Call(string(argL.Addrs[0]), "ChunkServer.SyncRPC", argC, &retC)
		if err != nil {
			logrus.Fatalln("Client write failed :", err)
			return
		}
		writtenBytes += roundWrittenBytes
		logrus.Debugf(" Write %d bytes to chunkserver %s, bytes written %d\n", roundWrittenBytes, argL.Addrs[0], writtenBytes)
	}
	// Set new file metadata back to master
	if offset+writtenBytes > fileSize {
		fileSize = offset + writtenBytes
	}
	argS = util.SetFileMetaArg{
		Path: path,
		Size: fileSize,
	}
	err = util.Call(string(c.masterAddr), "Master.SetFileMetaRPC", argS, &retS)
	if err != nil {
		return
	}
	return
}

// Write a file.
// should contact the master first, then write the data directly to chunkserver
func (c *Client) Write(w http.ResponseWriter, r *http.Request) {
	var argF util.GetFileMetaArg
	var retF util.GetFileMetaRet
	var argW util.WriteArg

	// Decode the params
	err := json.NewDecoder(r.Body).Decode(&argW)
	if err != nil {
		logrus.Fatalln("Client write failed :", err)
		w.WriteHeader(400)
		return
	}

	// Get the file metadata and check
	path := c.fdTable[argW.Fd]
	if path == "" {
		err = fmt.Errorf("Client write failed : fd %d is not opened\n", argW.Fd)
		return
	}
	argF.Path = path
	err = util.Call(string(c.masterAddr), "Master.GetFileMetaRPC", argF, &retF)
	if !retF.Exist {
		logrus.Fatalln("Client write failed :", err)
		return
	}
	fileSize := retF.Size
	if argW.Offset > fileSize {
		err = fmt.Errorf("Client write failed : write offset exceed file size\n")
		return
	}

	// Write to chunk
	writtenBytes, err := c._Write(path, argW.Offset, argW.Data, fileSize)
	if err != nil {
		logrus.Fatalln("Client write failed :", err)
		w.WriteHeader(400)
		return
	}
	msg, _ := json.Marshal(writtenBytes)
	w.Write(msg)
	return
}

func (c *Client) GetFileInfo(w http.ResponseWriter, r *http.Request) {
	var arg util.GetFileMetaArg
	var ret util.GetFileMetaRet
	// Decode the params
	err := json.NewDecoder(r.Body).Decode(&arg)
	if err != nil {
		logrus.Fatalln("Client getFileInfo failed :", err)
		w.WriteHeader(400)
		return
	}
	err = util.Call(string(c.masterAddr), "Master.GetFileMetaRPC", arg, &ret)
	msg, _ := json.Marshal(ret)
	w.Write(msg)
}
