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
	http.HandleFunc("/open", c.open)
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