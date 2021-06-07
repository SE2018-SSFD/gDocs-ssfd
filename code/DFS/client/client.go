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
	masterAddr string
	fdTable map[int]util.DFSPath
	//TODO:add lease
}

// InitClient initClient init a new client and return.
func InitClient(masterAddr string) *Client {
	c := &Client{
		masterAddr:   masterAddr,
		fdTable: make(map[int]util.DFSPath),
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

func (c *Client)Serve(addr string){
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
	err = util.Call(c.masterAddr, "Master.CreateRPC", arg, &ret)
	if err != nil {
		logrus.Fatalln("CreateRPC failed:",err)
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
	err = util.Call(c.masterAddr, "Master.MkdirRPC", arg, &ret)
	if err != nil {
		logrus.Fatalln("MkdirRPC failed:",err)
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
	for i:=0;i<util.MAXFD;i++{
		_,exist := c.fdTable[i]
		if !exist{
			logrus.Debugf("Client open : assign %d",i)
			c.fdTable[i] = arg.Path
			io.WriteString(w,strconv.Itoa(i))
			return
		}
	}
	w.WriteHeader(400)
	io.WriteString(w,strconv.Itoa(-1))
}

// close a file.
func (c *Client) close(w http.ResponseWriter, r *http.Request) {
	var arg util.CloseArg
	err := json.NewDecoder(r.Body).Decode(&arg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_,exist := c.fdTable[arg.Fd]
	if !exist{
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
	// out of range
	//err = fmt.Errorf("OutOfRangeError : the file %s has only %d chunks, requested the %d!\n",string(args.Path),len(fileState.chunks),args.ChunkIndex)
	//return err

}