package ylf

import (
	"fmt"
	"net"
	"strconv"
	"time"
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

/*获取一个 指定范围内的可用端口*/
func GetUsefulPort(ip string, startPort int, endPort int) int {
	if startPort > endPort {
		return 0
	}
	/*默认开50个并发*/
	useFulPort := make(chan int, 1000)
	for i := startPort; i <= endPort; i++ {
		go BeachCheckPort(ip, i, &useFulPort)
	}

	for {
		select {
		case uPort := <-useFulPort:
			if uPort != 0 {
				return uPort
			}
		case <-time.After(3 * time.Second):
			return 0
		}
	}
	return 0
}

/*检测端口并写入数据到channel*/
func BeachCheckPort(ip string, port int, useFulPort *chan int) {
	fmt.Println("检测了端口:", port)
	if !CheckPort(ip, fmt.Sprintf("%v", port)) {
		*useFulPort <- port
	} else {
		fmt.Println("端口存在:", port)
	}
}
