package client

import (
	"DFS/util"
	"DFS/util/zkWrap"
	json "encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type Client struct {
	sync.RWMutex
	fdLock  sync.Mutex
	clientAddr      util.Address
	masterAddr      util.Address
	fdTable         map[int]util.DFSPath
	s               *http.Server
	LeaderHeartbeat *zkWrap.Heartbeat // only one master(leader) and some clients in this room
	cidLock         sync.Mutex
	backupRead 		bool
	//TODO:add lease
}

func (c *Client) GetClientAddr() util.Address {
	return c.clientAddr
}

// InitClient initClient init a new client and return.
func InitClient(clientAddr util.Address, masterAddr util.Address) *Client {
	c := &Client{
		clientAddr: clientAddr,
		masterAddr: masterAddr, //TODO: we should not use this arg
		fdTable:    make(map[int]util.DFSPath),
		backupRead: false,
	}
	//to find master leader
	err := zkWrap.Chroot("/DFS")
	if err != nil {
		logrus.Fatalln(err)
		return nil
	}
	c.RegisterNodes()
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
	// mux.HandleFunc("/cappend", c.ConcurrentAppend)
	mux.HandleFunc("/list", c.List)
	mux.HandleFunc("/scan", c.Scan)
	mux.HandleFunc("/fileInfo", c.GetFileInfo)
	c.s = &http.Server{
		Addr:           string(c.clientAddr),
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err := c.s.ListenAndServe()
	if err != nil {
		logrus.Debug("Client server shutdown:", err)
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
	logrus.Debugf("Client Create : path %v",arg.Path)

	err = util.Call(string(c.masterAddr), "Master.CreateRPC", arg, &ret)
	if err != nil {
		logrus.Warnln("CreateRPC failed:", err)
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
	logrus.Debugf("Client Mkdir : path %v",arg.Path)
	err = util.Call(string(c.masterAddr), "Master.MkdirRPC", arg, &ret)
	if err != nil {
		logrus.Warnln("MkdirRPC failed:", err)
		return
	}
	io.WriteString(w, "0")
	return
}

// Open a file.
// If fd is depleted, return -1
func (c *Client) Open(w http.ResponseWriter, r *http.Request) {
	var argO util.OpenArg
	var retO util.OpenRet
	var argF util.GetFileMetaArg
	var retF util.GetFileMetaRet
	err := json.NewDecoder(r.Body).Decode(&argO)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	logrus.Debugf("Client Open : path %v",argO.Path)
	argF.Path = argO.Path
	err = util.Call(string(c.masterAddr), "Master.GetFileMetaRPC", argF, &retF)
	if err != nil || !retF.Exist {
		logrus.Warnln("Client open failed :", err)
		retO.Fd = -1
		msg, _ := json.Marshal(retO)
		w.Write(msg)
		return
	}
	c.fdLock.Lock()
	defer c.fdLock.Unlock()
	for i := util.MINFD; i < util.MAXFD; i++ {

		_, exist := c.fdTable[i]
		if !exist {
			logrus.Debugf("Client open : assign %d", i)
			c.fdTable[i] = argO.Path
			retO.Fd = i
			msg, _ := json.Marshal(retO)
			w.Write(msg)
			return
		}
	}
	w.WriteHeader(400)
	msg, _ := json.Marshal(retO)
	w.Write(msg)
}

// Close a file.
func (c *Client) Close(w http.ResponseWriter, r *http.Request) {
	var arg util.CloseArg
	err := json.NewDecoder(r.Body).Decode(&arg)
	defer func(err error) {
		if err != nil {
			w.Write([]byte(err.Error()))
		}
	}(err)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	logrus.Debugf("Client close : fd %v",arg.Fd)
	c.fdLock.Lock()
	defer c.fdLock.Unlock()
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

// func (c *Client) _Read(path util.DFSPath, offset int, len int, fileSize int) (readBytes int, buf []byte, err error) {
func (c *Client) _Read(path util.DFSPath, offset int, len int) (realReadBytes int, buf []byte, err error) {
	var argR util.GetReplicasArg
	var retR util.GetReplicasRet
	var argRCK util.ReadChunkArgs
	var retRCK util.ReadChunkReply
	var readBytes int
	argR.Path = path
	for readBytes < len {
		roundOff := (offset + readBytes) % util.MAXCHUNKSIZE
		roundReadBytes := int(math.Min(float64(util.MAXCHUNKSIZE-roundOff), float64(len-readBytes)))

		argR.ChunkIndex = (offset + readBytes) / util.MAXCHUNKSIZE
		err = util.Call(string(c.masterAddr), "Master.GetReplicasRPC", argR, &retR)
		if err != nil {
			return
		}
		argRCK.Handle = retR.ChunkHandle
		argRCK.Len = roundReadBytes
		argRCK.Off = roundOff
		err = util.Call(string(retR.ChunkServerAddrs[rand.Int()%util.REPLICATIONTIMES]), "ChunkServer.ReadChunkRPC", argRCK, &retRCK)
		realReadBytes += retRCK.Len
		if err != nil {
			logrus.Panicln("Client read chunk failed :", err)
			return
		}
		buf = append(buf,retRCK.Buf...)
		if retRCK.Len != roundReadBytes {
			logrus.Warnf("Client should read %v,buf only read %v", roundReadBytes, retRCK.Len)
			return
		}
		readBytes += roundReadBytes
		logrus.Debugf(" Read %d bytes from chunkserver %s, bytes read %d\n", roundReadBytes, string(retR.ChunkServerAddrs[0]), readBytes)
	}
	return
}

// func (c *Client) _Read(path util.DFSPath, offset int, len int, fileSize int) (readBytes int, buf []byte, err error) {
func (c *Client) _backupRead(path util.DFSPath, offset int, lens int) (realReadBytes int, buf []byte, err error) {
	var argR util.GetReplicasArg
	var retR util.GetReplicasRet
	var argRCK util.ReadChunkArgs
	var retRCK util.ReadChunkReply
	var retRCK1 util.ReadChunkReply
	var retRCK2 util.ReadChunkReply

	var readBytes int
	argR.Path = path
	for readBytes < lens {
		roundOff := (offset + readBytes) % util.MAXCHUNKSIZE
		roundReadBytes := int(math.Min(float64(util.MAXCHUNKSIZE-roundOff), float64(lens-readBytes)))

		argR.ChunkIndex = (offset + readBytes) / util.MAXCHUNKSIZE
		err = util.Call(string(c.masterAddr), "Master.GetReplicasRPC", argR, &retR)
		if err != nil {
			return
		}

		// Start 2-backup read
		argRCK.Handle = retR.ChunkHandle
		argRCK.Len = roundReadBytes
		argRCK.Off = roundOff
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			err = util.Call(string(retR.ChunkServerAddrs[0]), "ChunkServer.ReadChunkRPC", argRCK, &retRCK1)
			wg.Done()
		}()
		go func() {
			err = util.Call(string(retR.ChunkServerAddrs[1]), "ChunkServer.ReadChunkRPC", argRCK, &retRCK2)
			wg.Done()
		}()
		wg.Wait()
		if len(retRCK1.Buf) > len(retRCK2.Buf){
			retRCK = retRCK1
		}else{
			retRCK = retRCK2
		}
		if err != nil {
			logrus.Panicln("Client read failed :", err)
			return
		}
		buf = append(buf,retRCK.Buf...)
		if retRCK.Len != roundReadBytes {
			logrus.Warnf("Client should read %v,buf only read %v", roundReadBytes, retRCK.Len)
			return
		}
		readBytes += roundReadBytes
		logrus.Debugf(" Read %d bytes from chunkserver %s, bytes read %d\n", roundReadBytes, string(retR.ChunkServerAddrs[0]), readBytes)
		//logrus.Debugf("read:%s",retRCK.Buf)
	}
	return
}

// Read a file.
// should contact the master first, then get the data directly from chunkserver
func (c *Client) Read(w http.ResponseWriter, r *http.Request) {
	var argR util.ReadArg
	//var retR util.ReadRet
	var argF util.GetFileMetaArg
	var retF util.GetFileMetaRet
	// Decode the params
	err := json.NewDecoder(r.Body).Decode(&argR)
	if err != nil {
		logrus.Warnln("Client read failed :", err)
		w.WriteHeader(400)
		return
	}
	logrus.Debugf("Client read : fd %v,len %v,offset %v",argR.Fd,argR.Len,argR.Offset)

	// Get the file metadata and check
	//c.fdLock.RLock()
	pathh := c.fdTable[argR.Fd]
	//c.fdLock.RUnlock()
	if pathh == "" {
		err = fmt.Errorf("Client read failed : fd %d is not opened\n", argR.Fd)
		return
	}

	argF.Path = pathh
	err = util.Call(string(c.masterAddr), "Master.GetFileMetaRPC", argF, &retF)
	if err != nil || !retF.Exist {
		logrus.Warnln("Client read failed :", err)
		return
	}

	// fileSize := retF.Size
	// if argR.Offset+argR.Len > fileSize {
	// 	err = fmt.Errorf("Client read failed : read offset + read len  %d exceed file size %d",argR.Offset+argR.Len,fileSize)
	// 	logrus.Debugln(err)

	// 	return
	// }

	//TODO: check whether offset exceed filesize? (offset > chunknum * chunksize ?)

	// Read to chunk

	// readBytes, buf, err := c._Read(path, argR.Offset, argR.Len, fileSize)
	buf := make([]byte,0)
	c.RLock()
	if c.backupRead{
		_, buf, err = c._backupRead(pathh, argR.Offset, argR.Len)
	}else{
		_, buf, err = c._Read(pathh, argR.Offset, argR.Len)
	}
	c.RUnlock()
	if err != nil {
		logrus.Warnln("Client read failed :", err)
		w.WriteHeader(400)
		return
	}
	//retR.Data = buf
	//retR.Len = readBytes
	//msg, _ := json.Marshal(retR)
	//w.Write(msg)
	w.Write(buf) // return unmarshal dadta
	return

}

// Delete a file.
func (c *Client) Delete(w http.ResponseWriter, r *http.Request) {
	// var argR util.GetReplicasArg
	// var retR util.GetReplicasRet
	var argF util.GetFileMetaArg
	var retF util.GetFileMetaRet
	var argD util.DeleteArg
	var retD util.DeleteRet

	// Decode the params
	err := json.NewDecoder(r.Body).Decode(&argD)
	if err != nil {
		logrus.Warnln("Client delete failed :", err)
		w.WriteHeader(400)
		return
	}

	logrus.Debugf("Client delete : path %v",argD.Path)

	// Get the file metadata and check
	argF.Path = argD.Path
	err = util.Call(string(c.masterAddr), "Master.GetFileMetaRPC", argF, &retF)
	if !retF.Exist {
		logrus.Warnln("Client delete failed :", err)
		w.WriteHeader(400)
		return
	}

	// Delete the master metadata last
	err = util.Call(string(c.masterAddr), "Master.DeleteRPC", argD, &retD)
	if err != nil {
		logrus.Warnln("Client delete failed :", err)
		w.WriteHeader(400)
		return
	}

	// we don't need to delete chunk at once

	// Delete the chunk one by one
	// By default, the first entry int retR.ChunkServerAddr is the primary
	// chunkIndex := 0
	// for chunkIndex*util.MAXCHUNKSIZE < retF.Size {
	// 	argR.ChunkIndex = chunkIndex
	// 	err = util.Call(string(c.masterAddr), "Master.GetReplicasRPC", argR, &retR)
	// 	if err != nil {
	// 		logrus.Warnln("Client delete failed :", err)
	// 		w.WriteHeader(400)
	// 		return
	// 	}
	// 	// TODO delete file in chunkServer
	// 	chunkIndex += 1
	// }

	w.WriteHeader(200)
	return
}

// ConcurrentAppend to a file
// func (c *Client) ConcurrentAppend(w http.ResponseWriter, r *http.Request) {
// 	//var argG util.GetFileMetaArg
// 	var retG util.GetFileMetaRet
// 	var argC util.CAppendArg
// 	var retC util.CAppendRet
// 	// Decode the params
// 	err := json.NewDecoder(r.Body).Decode(&argC)
// 	if err != nil {
// 		logrus.Warnln("Client getFileInfo failed :", err)
// 		w.WriteHeader(400)
// 		return
// 	}
// 	// Get the file metadata and check
// 	path := c.fdTable[argC.Fd]
// 	if path == "" {
// 		err = fmt.Errorf("Client read failed : fd %d is not opened\n", argC.Fd)
// 		return
// 	}
// 	err = util.Call(string(c.masterAddr), "Master.GetFileMetaRPC", util.GetFileMetaArg{Path: path}, &retG)
// 	if err != nil {
// 		logrus.Warnln("Client concurrent append failed :", err)
// 		w.WriteHeader(400)
// 		return
// 	}
// 	if len(argC.Data) > util.MAXAPPENDSIZE { // append size cannot exist half a chunk
// 		logrus.Warnln("Client concurrent append failed : append size is too large")
// 		w.WriteHeader(400)
// 		return
// 	}
// 	var offset int
// 	chunkIndex := retG.Size / util.MAXCHUNKSIZE
// 	end := false
// 	// try append to chunk, pad it and retry on next chunk if normal failure
// 	// until success or unexpected error
// 	for !end {
// 		// TODO :finish it
// 		offset, err = c._ConcurrentAppend(chunkIndex, argC.Data)
// 		if err == nil {
// 			end = true
// 		}
// 	}
// 	retC.Offset = offset
// 	msg, _ := json.Marshal(retC)
// 	w.Write(msg)
// 	return
// }

// Append to a file
func (c *Client) Append(w http.ResponseWriter, r *http.Request) {
	var argA util.AppendArg
	var retA util.CAppendRet
	var argF util.GetFileMetaArg
	var retF util.GetFileMetaRet

	// Decode the params
	//err := json.NewDecoder(r.Body).Decode(&argA)
	//if err != nil {
	//	logrus.Warnln("Client append failed :", err)
	//	w.WriteHeader(400)
	//	return
	//}
	//logrus.Debugf("Client append : fd %v",argA.Fd)

	mr,err := r.MultipartReader()
	if err != nil{
		fmt.Println("r.MultipartReader() err,",err)
		return
	}
	for{
		p ,err := mr.NextPart()
		if err == io.EOF{
			break
		}
		if err != nil{
			fmt.Println("mr.NextPart() err,",err)
			break
		}
		//fmt.Println("part header:",p.Header)
		formName := p.FormName()
		fileName := p.FileName()
		if formName == "fd" && fileName == ""{
			formValue,_:= ioutil.ReadAll(p)
			argA.Fd,_ = strconv.Atoi(string(formValue))
		}
		if fileName != "" {
			fileData,_:=ioutil.ReadAll(p)
			argA.Data = fileData
		}
	}



	//Check append length
	if len(argA.Data) > util.MAXAPPENDSIZE { // append size cannot exist half a chunk
		logrus.Warnln("Client append failed : append size is too large")
		w.WriteHeader(400)
		return
	}

	// Get the file metadata and check
	//c.fdLock.RLock()
	pathh := c.fdTable[argA.Fd]
	//c.fdLock.RUnlock()

	if pathh== "" {
		err = fmt.Errorf("Client append failed : fd %d is not opened\n", argA.Fd)
		return
	}

	argF.Path = pathh
	err = util.Call(string(c.masterAddr), "Master.GetFileMetaRPC", argF, &retF)
	if !retF.Exist {
		logrus.Warnln("Client append failed GetFileMetaErr :", err)
		return
	}
	// fileSize := retF.Size

	// Write to file
	// writtenBytes, err := c._Write(path, fileSize, argA.Data, fileSize)

	offset, err := c._Append(pathh, argA.Data)
	if err != nil {
		logrus.Warnln("Client append failed :", err)
		w.WriteHeader(400)
		return
	}
	retA.Offset = offset
	logrus.Info("offset in client:", offset)
	msg, _ := json.Marshal(retA)
	w.Write(msg)
	return

}

func (c *Client) _Write(path util.DFSPath, offset int, data []byte) (writtenBytes int, err error) {
	var argR util.GetReplicasArg
	var retR util.GetReplicasRet
	var argL util.LoadDataArgs
	var retL util.LoadDataReply
	// var argS util.SetFileMetaArg
	// var retS util.SetFileMetaRet
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
		roundWrittenBytes := int(math.Min(float64(util.MAXCHUNKSIZE-(offset+writtenBytes)%util.MAXCHUNKSIZE), float64(len(data)-writtenBytes)))
		//logrus.Infof(" Write ChunkHandle : %d Addresses : %s %s %s, write %v\n", retR.ChunkHandle, retR.ChunkServerAddrs[0], retR.ChunkServerAddrs[1], retR.ChunkServerAddrs[2], roundWrittenBytes)
		var cid = c.newCacheID(retR.ChunkHandle)

		argL.CID = cid
		argL.Data = []byte(data[writtenBytes:(writtenBytes + roundWrittenBytes)])
		// TODO: make it random , now is fixed order
		argL.Addrs = retR.ChunkServerAddrs

		//argL.Addrs = make([]util.Address,0)
		//for _,index := range rand.Perm(len(retR.ChunkServerAddrs)){
		//	argL.Addrs = append(argL.Addrs,retR.ChunkServerAddrs[index])
		//}
		// Send to Master now
		err = util.Call(string(argL.Addrs[0]), "ChunkServer.LoadDataRPC", argL, &retL)
		if err != nil {
			logrus.Warnln("Client write failed LoadData :", err)
			return
		}
		argC = util.SyncArgs{
			CID:      cid,
			Off:      (offset + writtenBytes) % util.MAXCHUNKSIZE,
			Addrs:    retR.ChunkServerAddrs[1:],
			IsAppend: false,
		}
		err = util.Call(string(argL.Addrs[0]), "ChunkServer.SyncRPC", argC, &retC)
		if err != nil {
			logrus.Warnln("Client write failed Sync :", err)
			return
		}

		writtenBytes += roundWrittenBytes
		//logrus.Debugf(" Write %d bytes : %v, bytes written %d offset %d\n", roundWrittenBytes, argL.Data, writtenBytes,argC.Off)
	}
	// Set new file metadata back to master
	// if offset+writtenBytes > fileSize {
	// 	fileSize = offset + writtenBytes
	// }
	// argS = util.SetFileMetaArg{
	// 	Path: path,
	// 	Size: fileSize,
	// }
	// err = util.Call(string(c.masterAddr), "Master.SetFileMetaRPC", argS, &retS)
	// if err != nil {
	// 	return
	// }
	return
}

// Write a file.
// should contact the master first, then write the data directly to chunkserver
func (c *Client) Write(w http.ResponseWriter, r *http.Request) {
	var argF util.GetFileMetaArg
	var retF util.GetFileMetaRet
	var argW util.WriteArg
	var retW util.WriteRet

	mr,err := r.MultipartReader()
	if err != nil{
		fmt.Println("r.MultipartReader() err,",err)
		return
	}
	for{
		p ,err := mr.NextPart()
		if err == io.EOF{
			break
		}
		if err != nil{
			fmt.Println("mr.NextPart() err,",err)
			break
		}

		//fmt.Println("part header:",p.Header)
		formName := p.FormName()
		fileName := p.FileName()
		if formName == "fd" && fileName == ""{

			formValue,_:= ioutil.ReadAll(p)
			argW.Fd,_ = strconv.Atoi(string(formValue))
		}
		if formName == "offset" && fileName == ""{

			formValue,_:= ioutil.ReadAll(p)
			argW.Offset,_ = strconv.Atoi(string(formValue))
		}
		if fileName != "" {
			fileData,_:=ioutil.ReadAll(p)
			argW.Data = fileData
		}

	}
	// Decode the params
	//var inter map[string]interface{}
	//err := json.NewDecoder(r.Body).Decode(&inter)
	//argW.Fd = int(inter["fd"].(float64))
	//argW.Offset = int(inter["offset"].(float64))
	//logrus.Warnln(inter)
	//logrus.Warnln("-----")
	//argW.Data = []byte(strings.TrimSpace(inter["data"].(string)))
	//logrus.Warnln(argW.Data)
	//if err != nil {
	//	logrus.Warnln("Client write failed decode :", err)
	//	w.WriteHeader(400)
	//	return
	//}

	logrus.Debugf("Client write : fd %v, offset %v",argW.Fd,argW.Offset)

	// Get the file metadata and check
	//c.fdLock.RLock()
	pathh := c.fdTable[argW.Fd]
	//c.fdLock.RUnlock()
	if pathh == "" {
		err = fmt.Errorf("Client write failed : fd %d is not opened\n", argW.Fd)
		return
	}
	argF.Path = pathh
	err = util.Call(string(c.masterAddr), "Master.GetFileMetaRPC", argF, &retF)
	if !retF.Exist {
		logrus.Warnln("Client write failed GetFileMeta :", err)
		return
	}
	logrus.Debugf("client write path:%v,offset:%v,datasize:%v", pathh, argW.Offset, len(argW.Data))

	// fileSize := retF.Size
	// if argW.Offset > fileSize {
	// 	err = fmt.Errorf("Client write failed : write offset exceed file size\n")
	// 	return
	// }

	// Write to chunk
	// writtenBytes, err := c._Write(path, argW.Offset, argW.Data, fileSize)
	writtenBytes, err := c._Write(pathh, argW.Offset, argW.Data)
	if err != nil {
		logrus.Warnln("Client write failed _write :", err)
		w.WriteHeader(400)
		return
	}
	retW.BytesWritten = writtenBytes
	msg, _ := json.Marshal(retW)
	w.Write(msg)
	return
}

// GetFileInfo get one file information
func (c *Client) GetFileInfo(w http.ResponseWriter, r *http.Request) {
	var arg util.GetFileMetaArg
	var ret util.GetFileMetaRet
	var ret2 util.GetFileInfoRet
	// Decode the params
	err := json.NewDecoder(r.Body).Decode(&arg)
	if err != nil {
		logrus.Warnln("Client getFileInfo failed :", err)
		w.WriteHeader(400)
		return
	}
	logrus.Debugf("Client getFileInfo : path %v",arg.Path)

	err = util.Call(string(c.masterAddr), "Master.GetFileMetaRPC", arg, &ret)
	if err != nil {
		logrus.Warnln("Client getFileInfo failed :", err)
		w.WriteHeader(400)
		return
	}
	ret2 = util.GetFileInfoRet{
		Exist:         ret.Exist,
		IsDir:         ret.IsDir,
		UpperFileSize: ret.ChunkNum * util.MAXCHUNKSIZE,
	}
	msg, _ := json.Marshal(ret2)
	w.Write(msg)
}

func (c *Client) List(w http.ResponseWriter, r *http.Request) {
	var arg util.ListArg
	var ret util.ListRet
	// Decode the params
	err := json.NewDecoder(r.Body).Decode(&arg)
	if err != nil {
		logrus.Warnln("Client list failed :", err)
		w.WriteHeader(400)
		return
	}
	logrus.Debugf("Client list : path %v",arg.Path)

	err = util.Call(string(c.masterAddr), "Master.ListRPC", arg, &ret)
	msg, _ := json.Marshal(ret)
	w.Write(msg)
}

// Scan scan all files' information in a dir
func (c *Client) Scan(w http.ResponseWriter, r *http.Request) {
	var argS util.ScanArg
	var retS util.ScanRet
	retS.FileInfos = make([]util.GetFileMetaRet, 0)
	// Decode the params
	err := json.NewDecoder(r.Body).Decode(&argS)
	if err != nil {
		logrus.Warnln("Client Scan failed :", err)
		w.WriteHeader(400)
		return
	}
	logrus.Debugf("Client scan : path %v",argS.Path)
	err = util.Call(string(c.masterAddr), "Master.ScanRPC", argS, &retS)
	if err != nil {
		logrus.Warnln("Client Scan failed :", err)
		w.WriteHeader(400)
		return
	}
	msg, _ := json.Marshal(retS)
	w.Write(msg)
}

// helper method for ConcurrentAppend
func (c *Client) _ConcurrentAppend(index int, data string) (int, error) {
	return 0, nil
}

// helper method for Append
func (c *Client) _Append(path util.DFSPath, data []byte) (offset int, err error) {
	var argR util.GetReplicasArg
	var retR util.GetReplicasRet
	var argL util.LoadDataArgs
	var retL util.LoadDataReply
	var argC util.SyncArgs

	// Write the chunk (may add chunks)
	// By default, the first entry int retR.ChunkServerAddr is the primary
	argR.Path = path

	//TODO: 1.get the last chunk from master  (handle, address[])
	//2.append data to chunkserver  (load + sync)
	//3.if chunkserver says that chunk does not have enough room to store data, call getNextChunk in master (if have nextChunk,return newChunk;else, create a newChunk,return newChunk)
	//4.redo 2
	argR.ChunkIndex = -1
	for {
		var retC util.SyncReply

		err = util.Call(string(c.masterAddr), "Master.GetReplicasRPC", argR, &retR)
		if err != nil {
			return
		}
		//logrus.Debugf(" ChunkHandle : %d Addresses : %s %s %s\n", retR.ChunkHandle, retR.ChunkServerAddrs[0], retR.ChunkServerAddrs[1], retR.ChunkServerAddrs[2])
		// roundWrittenBytes := int(math.Min(float64(util.MAXCHUNKSIZE-(offset+writtenBytes)%util.MAXCHUNKSIZE), float64(len(data)-writtenBytes)))
		var cid = c.newCacheID(retR.ChunkHandle)
		argL.CID = cid
		argL.Data = []byte(data)

		// TODO: make it random , now is fixed order
		argL.Addrs = retR.ChunkServerAddrs
		//argL.Addrs = make([]util.Address,0)
		//for _,index := range rand.Perm(len(retR.ChunkServerAddrs)){
		//	argL.Addrs = append(argL.Addrs,retR.ChunkServerAddrs[index])
		//}
		// Send to Master now
		err = util.Call(string(argL.Addrs[0]), "ChunkServer.LoadDataRPC", argL, &retL)
		if err != nil {
			logrus.Warnln("Client append failed LoadFata :", err)
			return
		}
		argC = util.SyncArgs{
			CID:      cid,
			Off:      -1,
			Addrs:    retR.ChunkServerAddrs[1:],
			IsAppend: true,
		}
		err = util.Call(string(argL.Addrs[0]), "ChunkServer.SyncRPC", argC, &retC)
		if err == nil && retC.ErrorCode != util.NOSPACE {
			offset = retC.Off + retR.ChunkIndex*util.MAXCHUNKSIZE
			logrus.Debugf(" Append %d bytes to chunkserver %s, offset %d\n", len(data), argL.Addrs[0], offset)
			return
		} else if retC.ErrorCode != util.NOSPACE {
			//TODO: we should retry append
			logrus.Warnln("Client append failed sync :", err)
			break
		}

		// errorcode == nospace, try append to the next chunk

		logrus.Debugf("Client write file %v chunk %v no space, retry, Errcode: %v ,err: %v", path, retR.ChunkIndex, retC.ErrorCode, err)
		argR.ChunkIndex = retR.ChunkIndex + 1
	}
	return
}

func (c *Client) newCacheID(handle util.Handle) util.CacheID {
	c.cidLock.Lock()
	t := time.Now().UnixNano()
	c.cidLock.Unlock()

	var cid = util.CacheID{
		Handle:     handle,
		ClientAddr: c.clientAddr,
		Timestamp:  t,
	}
	return cid
}
