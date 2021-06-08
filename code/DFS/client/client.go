package client

import (
	"DFS/util"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"math"
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
	http.HandleFunc("/create", c.Create)
	http.HandleFunc("/mkdir", c.Mkdir)
	http.HandleFunc("/delete", c.Delete)
	http.HandleFunc("/read", c.Read)
	http.HandleFunc("/write", c.Write)
	http.HandleFunc("/open", c.Open)
	http.HandleFunc("/close", c.Close)
	return c
}

func (c *Client) Serve() {
	err := http.ListenAndServe(string(c.clientAddr), nil)
	if err != nil {
		logrus.Fatal("Client server shutdown!\n")
	}

}

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
	io.WriteString(w,"0")
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
	io.WriteString(w,"0")
	return
}

// Delete a file.
func (c *Client) Delete(w http.ResponseWriter, r *http.Request) {
}

// Open a file.
// If fd is depleted, return -1
func (c *Client) Open(w http.ResponseWriter, r *http.Request) {
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

// Close a file.
func (c *Client) Close(w http.ResponseWriter, r *http.Request) {
	var arg util.CloseArg
	err := json.NewDecoder(r.Body).Decode(&arg)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, exist := c.fdTable[arg.Fd]
	if !exist {
		w.WriteHeader(400)
		return
	}
	logrus.Debugf("Client close : free %d", arg.Fd)
	delete(c.fdTable, arg.Fd)
	return
}

// Read a file.
// should contact the master first, then get the data directly from chunkserver
func (c *Client) Read(w http.ResponseWriter, r *http.Request) {
}

// write a file.
// should contact the master first, then write the data directly to chunkserver
func (c *Client) Write(w http.ResponseWriter, r *http.Request) {
	var argR util.GetReplicasArg
	var retR util.GetReplicasRet
	var argF util.GetFileMetaArg
	var retF util.GetFileMetaRet
	var argW util.WriteArg
	var argL util.LoadDataArgs
	var argS util.SetFileMetaArg
	var retS util.SetFileMetaRet
	// var retL util.LoadDataReply

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
		logrus.Fatalf("Client write failed : fd %d is not opened\n", argW.Fd)
		w.WriteHeader(400)
		return
	}
	argF.Path = path
	err = util.Call(string(c.masterAddr), "Master.GetFileMetaRPC", argF, &retF)
	if !retF.Exist {
		logrus.Fatalln("Client write failed :", err)
		w.WriteHeader(400)
		return
	}
	fileSize := retF.Size
	writtenBytes := 0
	if argW.Offset > fileSize {
		logrus.Fatalln("Client write failed : write offset exceed file size")
		w.WriteHeader(400)
		return
	}

	// Write the chunk (may add chunks)
	// By default, the first entry int retR.ChunkServerAddr is the primary
	argR.Path = path
	for writtenBytes < len(argW.Data) {
		argR.ChunkIndex = (argW.Offset + writtenBytes) / util.MAXCHUNKSIZE
		err = util.Call(string(c.masterAddr), "Master.GetReplicasRPC", argR, &retR)
		if err!=nil{
			logrus.Fatalln("Client write failed :", err)
			w.WriteHeader(400)
			return
		}
		logrus.Debugf(" ChunkHandle : %d Addresses : %s %s %s\n",retR.ChunkHandle,retR.ChunkServerAddrs[0],retR.ChunkServerAddrs[1],retR.ChunkServerAddrs[2])
		roundWrittenBytes := int(math.Min(float64(util.MAXCHUNKSIZE-(argW.Offset+writtenBytes)%util.MAXCHUNKSIZE), float64(len(argW.Data)-writtenBytes)))
		argL.CID = util.CacheID{
			Handle: retR.ChunkHandle,
			ClientAddr: c.clientAddr,
		}
		argL.Data = argW.Data[(argW.Offset + writtenBytes):(argW.Offset + writtenBytes + roundWrittenBytes)]
		argL.Addrs = retR.ChunkServerAddrs
		//TODO: make it random
		//argL.Addrs = make([]util.Address,0)
		//for _,index := range rand.Perm(len(retR.ChunkServerAddrs)){
		//	argL.Addrs = append(argL.Addrs,retR.ChunkServerAddrs[index])
		//}
		// err = util.Call(string(argL.Addrs[0]), "ChunkServer.LoadDataRPC", argL, &retL)
		writtenBytes += roundWrittenBytes
		logrus.Debugf(" Write %d bytes to chunkserver %s, bytes written %d\n",roundWrittenBytes,argL.Addrs[0],writtenBytes)
	}

	// Set file metadata back to master
	argS = util.SetFileMetaArg{
		Path: path,
		Size: fileSize+writtenBytes,
	}
	err = util.Call(string(c.masterAddr), "Master.SetFileMetaRPC", argS, &retS)
	if err!=nil{
		logrus.Fatalln("Client write failed :", err)
		return
	}
	w.WriteHeader(200)
	return
}
