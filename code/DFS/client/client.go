package client

import (
	"DFS/util"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"net/http"
)

type Client struct {
	masterAddr string
	//TODO:add lease
}

// InitClient initClient init a new client and return.
func InitClient(masterAddr string) *Client {
	c := &Client{
		masterAddr:   masterAddr,
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
func (c *Client) open(w http.ResponseWriter, r *http.Request) {
}

// close a file.
func (c *Client) close(w http.ResponseWriter, r *http.Request) {
}

// read a file.
// should contact the master first, then get the data directly from chunkserver
func (c *Client) read(w http.ResponseWriter, r *http.Request) {
}

// write a file.
// should contact the master first, then write the data directly to chunkserver
func (c *Client) write(w http.ResponseWriter, r *http.Request) {
}