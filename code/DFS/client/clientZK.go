package client

import (
	"DFS/util"
	"DFS/util/zkWrap"
	"strings"

	"github.com/sirupsen/logrus"
)

func (c *Client) onHeartbeatConn(_ string, who string) {
	logrus.Print("leader heartbeart conn")
	if strings.Compare("master", who[:6]) == 0 {
		c.masterAddr = util.Address(who[6:])
		//TODO: maybe we should clean fdTable
	}
}

func onHeartbeatDisConn(_ string, who string) {
}

func (c *Client) RegisterNodes() {
	hb, err := zkWrap.RegisterHeartbeat(
		"masterLeader",
		util.HEARTBEATDURATION,
		"client"+string(c.clientAddr),
		c.onHeartbeatConn,
		onHeartbeatDisConn,
	)
	if err != nil {
		panic(err)
	}

	c.LeaderHeartbeat = hb

	for _, mate := range hb.GetOriginMates() {
		c.onHeartbeatConn("", mate)
	}
}
