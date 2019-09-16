package ylf

import (
	"net"
	"strconv"
)

/*检查端口是否开放*/
func CheckPort(ip string, port string) bool {
	portInt, _ := strconv.Atoi(port)
	tcpAddr := net.TCPAddr{
		IP:   net.ParseIP(ip),
		Port: portInt,
	}
	conn, err := net.DialTCP("tcp", nil, &tcpAddr)
	if err == nil {
		return true
	}
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	return false
}
