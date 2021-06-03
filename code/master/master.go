package master

import "net"

type Master struct {
	addr    string
	 string
	metaPath string
	l          net.Listener
}

func InitMaster(addr string, metaPath string) *Master{
	m := &Master{
		addr: addr,
		metaPath: metaPath,
	}
	return m
}
func (m *Master)GetStatusString()string{
	return "Master address :"+m.addr+ ",metaPath :"+m.metaPath
}