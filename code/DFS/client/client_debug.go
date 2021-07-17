package client

import "github.com/sirupsen/logrus"

func (c *Client) PrintMasterAddr() {
	logrus.Print("Master Addr : ", c.masterAddr)
}
