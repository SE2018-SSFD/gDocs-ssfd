package test
//
//import (
//	"backend/lib/zkWrap"
//	"stathat.com/c/consistent"
//	"strconv"
//	"strings"
//	"time"
//)
//
//var (
//	myAddr					string
//	myNodeNum				int
//	consistentHash			*consistent.Consistent
//	clusterHeartbeat		*zkWrap.Heartbeat
//)
//
//func marshalHashName(addr string, num int) string {
//	return addr + "#" + strconv.Itoa(num)
//}
//
//func unMarshalHashName(hashName string) (addr string, num int) {
//	split := strings.SplitN(hashName, "#", 2)
//	addr = split[0]
//	num, err := strconv.Atoi(split[1])
//	if err != nil {
//		panic(err)
//	}
//
//	return addr, num
//}
//
//func onHeartbeatConn(_ string, who string) {
//	addr, num := unMarshalHashName(who)
//
//	for i := 0; i < num; i += 1 {
//		consistentHash.Add(marshalHashName(addr, i))
//	}
//}
//
//func onHeartbeatDisConn(_ string, who string) {
//	addr, num := unMarshalHashName(who)
//
//	for i := 0; i < num; i += 1 {
//		consistentHash.Remove(marshalHashName(addr, i))
//	}
//}
//
//func RegisterNodes(addr string, num int) {
//	myAddr, myNodeNum = addr, num
//	consistentHash = consistent.New()
//	for i := 0; i < num; i += 1 {
//		consistentHash.Add(addr + "#" + strconv.Itoa(i))
//	}
//
//	hb, err := zkWrap.RegisterHeartbeat(
//		"cache",
//		15 * time.Second,
//		addr + "#" + strconv.Itoa(num),
//		onHeartbeatConn,
//		onHeartbeatDisConn,
//	)
//	if err != nil {
//		panic(err)
//	}
//
//	clusterHeartbeat = hb
//
//	for _, mate := range hb.GetOriginMates() {
//		onHeartbeatConn("", mate)
//	}
//}
//
//func FileBelongsTo(filename string, fid uint) (addr string, isMine bool) {
//	addrRaw, err := consistentHash.Get(filename + strconv.FormatUint(uint64(fid), 10))
//	if err != nil {
//		panic(err)
//	}
//
//	addr = strings.SplitN(addrRaw, "#", 2)[0]
//	isMine = addr == myAddr
//
//	return addr, isMine
//}