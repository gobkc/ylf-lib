package ylf

import (
	"errors"
	"fmt"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	"log"
	"math/rand"
	"net"
	"time"
)

/*给eth添加IP/子网掩码*/
func AddIpToEth(eth string, ip string) error {
	ethObj, err := netlink.LinkByName(eth)
	if err != nil {
		return err
	}
	addr, err := netlink.ParseAddr(ip)
	if err != nil {
		return err
	}
	if err = netlink.AddrAdd(ethObj, addr); err != nil {
		return err
	}
	return nil
}

/*获取随机MAC地址*/
func GetRandMac() string {
	//generator mac addr
	var m [6]byte

	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 6; i++ {
		mac_byte := rand.Intn(256)
		m[i] = byte(mac_byte)

		rand.Seed(int64(mac_byte))
	}
	return fmt.Sprintf("00:35:5d:%02x:%02x:%02x", m[3], m[4], m[5])
}

/*添加网桥*/
func AddBridge(bridgeName string) error {
	la := netlink.NewLinkAttrs()
	la.Name = bridgeName
	myBridge := &netlink.Bridge{LinkAttrs: la}
	err := netlink.LinkAdd(myBridge)
	if err != nil {
		return err
	}
	return nil
}

/*设置网桥地址并启用*/
func SetBridgeAddr(name string, ip string, tryTimes int) error {
	/*尝试次数至少是1*/
	if tryTimes < 1 {
		tryTimes = 1
	}
	var iFace netlink.Link
	var err error
	for i := 0; i < tryTimes; i++ {
		// 根据名字找到设备
		iFace, err = netlink.LinkByName(name)
		if err == nil {
			break
		}
		log.Printf("获取网桥链接[ %s ]时出错... 重试第%s次", name, i)
		/*第一次尝试失败,则沉睡2秒后继续尝试*/
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return err
	}
	/*将原始ip如192.168.0.1/24转换成*net.IPNet类型*/
	ipNet, err := netlink.ParseIPNet(ip)
	if err != nil {
		return err
	}

	/*获取netLink的一个IP地址结构体引用,参数就是ipNet*/
	addr := &netlink.Addr{ipNet, "", 0, 0, nil, nil, 0, 0}

	/*等于 ip addr add 192.168.0.1/24 dev testbridge*/
	err = netlink.AddrAdd(iFace, addr)

	if err != nil {
		return err
	}

	/*启用网桥 等于 ip link set testbridge up*/
	if err := netlink.LinkSetUp(iFace); err != nil {
		return err
	}

	return nil
}

/*删除虚拟设备 网桥 vEth等*/
func DeleteDev(name string) error {
	/*根据设备名找到该设备*/
	l, err := netlink.LinkByName(name)
	if err != nil {
		return err
	}

	/*删除网桥就等于 ifconfig testbridge down && ip link delete testbridge type bridge*/
	/*删除vEth就等于  ip link delete 12345 type veth*/
	if err := netlink.LinkDel(l); err != nil {
		return err
	}
	return nil
}

/*启用设备*/
func SetDeviceUP(devName string) error {
	iFace, err := netlink.LinkByName(devName)
	if err != nil {
		return err
	}

	if err := netlink.LinkSetUp(iFace); err != nil {
		return err
	}
	return nil
}

/*给设备或虚拟设备设置IP*/
func SetDeviceIP(name string, ip string) error {
	retries := 2
	var iFace netlink.Link
	var err error
	for i := 0; i < retries; i++ {
		iFace, err = netlink.LinkByName(name)
		if err == nil {
			break
		}
		log.Println("获取网桥链接[ %s ]时出错... 重试第%s次", name, i)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return err
	}
	ipNet, err := netlink.ParseIPNet(ip)
	if err != nil {
		return err
	}
	addr := &netlink.Addr{ipNet, "", 0, 0, nil, nil, 0, 0}
	return netlink.AddrAdd(iFace, addr)
}

/*添加macVLan*/
func AddMacVLan(devName string, parentDevName string, macAddr string) error {
	var err error

	/*根据设备名找到设备*/
	link, err := netlink.LinkByName(parentDevName)
	if err != nil {
		return err
	}

	mac, err := net.ParseMAC(macAddr)
	if err != nil {
		return errors.New("在添加macvlan时无法解析mac地址：" + err.Error())
	}

	mv := netlink.NewLinkAttrs()
	mv.Name = devName
	mv.ParentIndex = link.Attrs().Index
	mv.HardwareAddr = mac

	macVLan := netlink.Macvlan{
		LinkAttrs: mv,
		Mode:      netlink.MACVLAN_MODE_BRIDGE,
	}

	/*先删除同名的macVLan，防止出错*/
	netlink.LinkDel(&macVLan)

	/*添加macVLan*/
	if err = netlink.LinkAdd(&macVLan); err != nil {
		return errors.New("在添加macvlan时报错:" + err.Error())
	}

	err = netlink.LinkSetDown(&macVLan)
	if err != nil {
		return err
	}

	/*启用macVLan*/
	if err = netlink.LinkSetUp(&macVLan); err != nil {
		return errors.New("在启用macvlan时报错:" + devName + " " + err.Error())
	}

	return nil
}

/*添加默认路由*/
func AddDefaultRoute(ppp string, gatewayIp string, tableId int) error {
	_, inet, _ := net.ParseCIDR(fmt.Sprintf("%s/32", gatewayIp))
	_, cidr, _ := net.ParseCIDR("0.0.0.0/0")

	peerLink, err := netlink.LinkByName(ppp)
	if err != nil {
		return errors.New("在获取网卡设备时发生错误:" + err.Error())
	}
	defaultRoute := &netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		Gw:        inet.IP,
		Dst:       cidr,
		Table:     tableId,
	}
	if err := netlink.RouteAdd(defaultRoute); err != nil {
		return errors.New("在添加默认路由时发生错误:" + err.Error())
	}
	return nil
}

/*添加策略*/
func AddRule(src string, tableID int, fref int) error {
	var err error

	/*分割字符串IP地址.使之符合要求*/
	_, srcIpNet, err := net.ParseCIDR(src)
	if err != nil {
		return errors.New("添加策略时，因分割字符串转ipNet时发生错误")
	}

	rule := netlink.NewRule()
	rule.Table = tableID
	rule.Src = srcIpNet
	rule.Priority = fref
	rule.Invert = false

	/*添加rule*/
	if err = netlink.RuleAdd(rule); err != nil {
		return errors.New("在添加策略时报错:" + err.Error())
	}

	return nil
}

/*使用eth获取Mac地址*/
func GetMacByEth(eth string) (string, error) {
	net, err := net.InterfaceByName(eth)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	mac := net.HardwareAddr.String()
	return mac, nil
}

/*添加DNS*/
func AddDns(src string, tableID int, fref int) error {
	var err error

	/*分割字符串IP地址.使之符合要求*/
	_, srcIpNet, err := net.ParseCIDR(src)
	if err != nil {
		return errors.New("添加DNS时，因分割字符串转ipNet时发生错误")
	}

	rule := netlink.NewRule()
	rule.Table = tableID
	rule.Dst = srcIpNet
	rule.Priority = fref
	rule.Invert = false

	/*添加dns*/
	if err = netlink.RuleAdd(rule); err != nil {
		return errors.New("在添加DNS时报错:" + err.Error())
	}

	return nil
}

/*批量删除策略 根据策略表ID*/
func DelRuleByTableId(tableID int) {
	rule := netlink.NewRule()
	rule.Table = tableID
	for {
		if err := netlink.RuleDel(rule); err != nil {
			break
		}
	}
}

/*给PPPOE指定策略表*/
func SetPPPOE(eth string, ruleTableID int) error {
	//未考虑重复规则情况
	link, err := netlink.LinkByName(eth)
	if err != nil {
		return err
	}
	addrList, err := netlink.AddrList(link, unix.AF_INET)
	for _, addr := range addrList {
		/*addr.Peer获取IP+网关*/
		log.Println(addr.Peer)
		/*addr.IP 只获取IP*/
		log.Println(addr.IP)
		route := netlink.Route{
			LinkIndex: link.Attrs().Index,
			Dst:       addr.Peer,
			Src:       addr.IP,
			Table:     ruleTableID,
			Scope:     unix.RT_SCOPE_LINK,
			Protocol:  unix.RTPROT_KERNEL,
		}
		// 添加路由
		if err := netlink.RouteAdd(&route); err != nil {
			log.Println(err)
			return err
		}
		// 添加策略
		rule := netlink.NewRule()
		_, srcnet, _ := net.ParseCIDR(addr.IP.String())
		rule.Src = srcnet
		rule.Dst = addr.Peer
		rule.Table = ruleTableID
		netlink.RuleAdd(rule)
	}
	return nil
}