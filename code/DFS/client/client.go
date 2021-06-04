package client

import (
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
	http.HandleFunc("/delete", c.delete)
	http.HandleFunc("/read", c.read)
	http.HandleFunc("/write", c.write)
	http.HandleFunc("/open", c.open)
	http.HandleFunc("/close", c.close)
	return c
}

func (c *Client)Serve(){
	err := http.ListenAndServe("localhost:8000", nil)
	if err != nil {
		logrus.Fatal("Client server shutdown!\n")
	}

}

// create a file.
func (c *Client) create(w http.ResponseWriter, r *http.Request) {
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