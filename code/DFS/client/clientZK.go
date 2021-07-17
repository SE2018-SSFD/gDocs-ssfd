package client

import (
	"DFS/util"
	"DFS/util/zkWrap"
	"github.com/sirupsen/logrus"
	"strings"
)

func (c *Client) onHeartbeatConn(me string, who string) {
	if strings.Compare("master", who[:6]) == 0 {
		logrus.Infof("%v leader heartbeart conn : master leader %v join",me,who)
		c.masterAddr = util.Address(who[6:])
		//TODO: maybe we should clean fdTable
	}else{
		logrus.Infof("%v leader heartbeart conn : another client %v join",me,who)
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
