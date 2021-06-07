package client

import (
	"DFS/util"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strconv"
)

type Client struct {
	clientAddr util.Address
	masterAddr util.Address
	fdTable    map[int]util.DFSPath
	//TODO:add lease
}

// InitClient initClient init a new client and return.
func InitClient(clientAddr util.Address,masterAddr util.Address) *Client {
	c := &Client{
		clientAddr : clientAddr,
		masterAddr: masterAddr,
		fdTable:    make(map[int]util.DFSPath),
	}
	http.HandleFunc("/create", c.create)
	http.HandleFunc("/mkdir", c.mkdir)
	http.HandleFunc("/delete", c.delete)
	http.HandleFunc("/read", c.read)
	http.HandleFunc("/write", c.write)
	http.HandleFunc("/open", c.open)
	http.HandleFunc("/close", c.close)
	return c
}

func (c *Client) Serve(addr string) {
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		logrus.Fatal("Client server shutdown!\n")
	}

}

// create a file.
func (c *Client) create(w http.ResponseWriter, r *http.Request) {
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
	return
}

// mkdir a dir.
func (c *Client) mkdir(w http.ResponseWriter, r *http.Request) {
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
	return
}

// delete a file.
func (c *Client) delete(w http.ResponseWriter, r *http.Request) {
}

// open a file.
// if fd is not enough, return -1
func (c *Client) open(w http.ResponseWriter, r *http.Request) {
	var arg util.OpenArg
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
	io.WriteString(w, strconv.Itoa(-1))
}

// close a file.
func (c *Client) close(w http.ResponseWriter, r *http.Request) {
	var arg util.CloseArg
	err := json.NewDecoder(r.Body).Decode(&arg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, exist := c.fdTable[arg.Fd]
	if !exist {
		w.WriteHeader(400)
		return
	}
	delete(c.fdTable, arg.Fd)
	return
}

// read a file.
// should contact the master first, then get the data directly from chunkserver
func (c *Client) read(w http.ResponseWriter, r *http.Request) {
}

// write a file.
// should contact the master first, then write the data directly to chunkserver
func (c *Client) write(w http.ResponseWriter, r *http.Request) {
	var argR util.GetReplicasArg
	var retR util.GetReplicasRet
	var argF util.GetFileMetaArg
	var retF util.GetFileMetaRet
	var argW util.WriteArg
	var argL util.LoadDataArgs
	var retL util.LoadDataReply
	w.WriteHeader(400)

	// Decode the params
	err := json.NewDecoder(r.Body).Decode(&argW)
	if err != nil {
		logrus.Fatalln("GetFileMetaRPC failed :", err)
		return
	}
	// Get the file metadata
	path := c.fdTable[argW.Fd]
	if path == "" {
		logrus.Fatalf("GetFileMetaRPC failed : fd %d is not opened\n", argW.Fd)
		return
	}
	argF.Path = path
	err = util.Call(string(c.masterAddr), "Master.GetFileMetaRPC", argF, &retF)
	if !retF.Exist {
		logrus.Fatalln("GetFileMetaRPC failed :", err)
		return
	}
	fileSize := retF.Size
	remainSize := len(argW.Data)
	writtenBytes := 0
	if argW.Offset > fileSize {
		logrus.Fatalln("GetFileMetaRPC failed : write offset exceed file size")
		return
	}
	// Write the chunk (may add chunks)
	// By default, the first entry int retR.ChunkServerAddr is the primary
	argR.Path = path
	for remainSize > 0 {
		argR.ChunkIndex = (argW.Offset + writtenBytes) / util.MAXCHUNKSIZE
		err = util.Call(string(c.masterAddr), "Master.GetReplicasRPC", argR, &retR)
		roundWrittenBytes := util.MAXCHUNKSIZE -  (argW.Offset + writtenBytes) % util.MAXCHUNKSIZE
		argL.CID = util.CacheID{
			Handle: retR.ChunkHandle,
			ClientAddr: c.clientAddr,
		}
		argL.Data = nil
		argL.Addrs = retR.ChunkServerAddrs
		//TODO: make it random
		//argL.Addrs = make([]util.Address,0)
		//for _,index := range rand.Perm(len(retR.ChunkServerAddrs)){
		//	argL.Addrs = append(argL.Addrs,retR.ChunkServerAddrs[index])
		//}
		err = util.Call(string(argL.Addrs[0]), "ChunkServer.LoadDataRPC", argL, &retL)
		writtenBytes += roundWrittenBytes
	}
	w.WriteHeader(200)
	return
}
