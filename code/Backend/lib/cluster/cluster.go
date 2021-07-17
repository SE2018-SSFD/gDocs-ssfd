package cluster

import (
	"backend/lib/zkWrap"
	"backend/utils/logger"
	"stathat.com/c/consistent"
	"strconv"
	"strings"
	"time"
)

var (
	myAddr					string
	myNodeNum				int
	consistentHash			*consistent.Consistent
	clusterHeartbeat		*zkWrap.Heartbeat
)

func marshalHashName(addr string, num int) string {
	return addr + "#" + strconv.Itoa(num)
}

func unMarshalHashName(hashName string) (addr string, num int) {
	split := strings.SplitN(hashName, "#", 2)
	addr = split[0]
	num, err := strconv.Atoi(split[1])
	if err != nil {
		panic(err)
	}

	return addr, num
}

func onHeartbeatConn(_ string, who string) {
	addr, num := unMarshalHashName(who)

	for i := 0; i < num; i += 1 {
		consistentHash.Add(marshalHashName(addr, i))
	}

	logger.Infof("[addr(%s)\tnum(%d)] Backend enter!", addr, num)
	logger.Infof("Current members: %+v", consistentHash.Members())
}

func onHeartbeatDisConn(_ string, who string) {
	addr, num := unMarshalHashName(who)

	for i := 0; i < num; i += 1 {
		consistentHash.Remove(marshalHashName(addr, i))
	}

	logger.Infof("[addr(%s)\tnum(%d)] Backend leave!", addr, num)
	logger.Infof("Current members: %+v", consistentHash.Members())
}

func RegisterNodes(addr string, num int) {
	myAddr, myNodeNum = addr, num
	consistentHash = consistent.New()
	for i := 0; i < num; i += 1 {
		consistentHash.Add(addr + "#" + strconv.Itoa(i))
	}

	hb, err := zkWrap.RegisterHeartbeat(
		"cache",
		7 * time.Second,
		addr + "#" + strconv.Itoa(num),
		onHeartbeatConn,
		onHeartbeatDisConn,
	)
	if err != nil {
		panic(err)
	}

	clusterHeartbeat = hb

	mates := hb.GetOriginMates()
	for _, mate := range mates {
		onHeartbeatConn("", mate)
	}
}

func FileBelongsTo(filename string, fid uint) (addr string, isMine bool) {
	addrRaw, err := consistentHash.Get(filename + strconv.FormatUint(uint64(fid), 10))
	if err != nil {
		panic(err)
	}

	addr, _ = unMarshalHashName(addrRaw)
	isMine = addr == myAddr

	return addr, isMine
}