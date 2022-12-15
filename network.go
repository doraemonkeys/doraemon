package doraemon

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/cookiejar"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/axgle/mahonia"
	"github.com/prestonTao/upnp"
	"golang.org/x/net/publicsuffix"
)

// 获取本机已保存的所有wifi
func GetSavedWifi() (string, error) {
	cmd := exec.Command("netsh", "wlan", "show", "profiles")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	dec := mahonia.NewDecoder("gbk")
	out = []byte(dec.ConvertString(string(out)))
	return string(out), nil
}

// 获取当前wifi名称
func GetWifiName() (string, error) {
	cmd := exec.Command("netsh", "wlan", "show", "interfaces")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	dec := mahonia.NewDecoder("gbk")
	out = []byte(dec.ConvertString(string(out)))
	re := regexp.MustCompile(`配置文件[ ]*:[ ]*([\S]+)`)
	match := re.FindSubmatch(out)
	if match == nil {
		return "", nil
	}
	return string(match[1]), nil
}

// 获取当前网络密码
func GetWifiPassword(wifiname string) (string, error) {
	cmd := exec.Command("netsh", "wlan", "show", "profile", wifiname, "key=clear")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	dec := mahonia.NewDecoder("gbk")
	out = []byte(dec.ConvertString(string(out)))
	re := regexp.MustCompile(`关键内容[ ]*:[ ]*([\S]+)`)
	match := re.FindSubmatch(out)
	if match == nil {
		return "", nil
	}
	return string(match[1]), nil
}

// ping局域网内所有ip
func PingAll(srcIP string) error {
	//检查ip是否合法
	ip := net.ParseIP(srcIP)
	if ip == nil {
		return fmt.Errorf("ip地址不合法")
	}
	//检查ip是否为ipv4
	if ip.To4() == nil {
		return fmt.Errorf("ip地址不是ipv4")
	}
	IpPrefix := srcIP[:strings.LastIndex(srcIP, ".")+1]
	ch := make(chan string)
	workers := 40
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(ch chan string, wg *sync.WaitGroup) {
			for v := range ch {
				PingToUpdateARP(v)
			}
			wg.Done()
		}(ch, &wg)
	}
	//ping -l 1 -n 1 -w 100 ip
	for i := 1; i < 255; i++ {
		ip := IpPrefix + fmt.Sprint(i)
		ch <- ip
	}
	close(ch)
	wg.Wait()
	return nil
}

func PingToUpdateARP(ip string) {
	CmdNoOutput("", []string{"ping", "-l", "1", "-n", "1", "-w", "500", ip, "&", "exit"})
}

// 获取局域网内所有主机IP与MAC地址(通过ping命令更新arp表,不包含自己),
// 通过LanIP获取局域网IP段,通过PingAll更新arp表。
// 返回map[ip]mac。
// 更好的通用方法https://studygolang.com/articles/1202
// https://github.com/mnhkahn/go_code/blob/master/fing.go
func GetAllHosts(lanIP string) (map[string]string, error) {
	err := PingAll(lanIP)
	if err != nil {
		return nil, err
	}
	time.Sleep(time.Second * 6)
	output, err := Cmd("", []string{"arp", "-a", "&", "exit"})
	if err != nil {
		return nil, err
	}
	//使用正则表达式匹配ip与mac地址
	ipPrefix := lanIP[:strings.LastIndex(lanIP, ".")+1]
	//转义前缀
	ipPrefixEscape := strings.Replace(ipPrefix, ".", `\.`, -1)
	re := regexp.MustCompile("(" + ipPrefixEscape + `[0-9]{1,3})[ ]*([0-9a-zA-Z-:]{10,})`)
	//re := regexp.MustCompile("(" + ipPrefix + `(?!255)(?!1[ ])[0-9]{1,3})[ ]*([0-9a-zA-Z-:]+)`)
	match := re.FindAllStringSubmatch(output, -1)
	if match == nil {
		return nil, fmt.Errorf("未找到主机")
	}
	hosts := make(map[string]string)
	for _, v := range match {
		hosts[v[1]] = v[2]
	}
	//删除子网广播地址
	_, ok := hosts[ipPrefix+"255"]
	if ok {
		delete(hosts, ipPrefix+"255")
	}
	return hosts, nil
}

// 获取网关地址
// 还可以通过构建ICMP报文实现路由追踪(traceroute)，记录访问某个网站经过的路径，那么第一条路径就是访问路由器
func GetGateway() (string, error) {
	upnpMan := new(upnp.Upnp)
	err := upnpMan.SearchGateway()
	if err != nil {
		return "", err
	} else {
		return upnpMan.Gateway.Host[0:strings.LastIndex(upnpMan.Gateway.Host, ":")], nil
	}
}

// 获取路由器的外部IP，需要路由器支持UPNP协议。
func GetGatewayOutsideIP() (string, error) {
	upnpMan := new(upnp.Upnp)
	err := upnpMan.ExternalIPAddr()
	if err != nil {
		return "", err
	} else {
		return upnpMan.GatewayOutsideIP, nil
	}
}

func GetLocalIP() (string, error) {
	upnpMan := new(upnp.Upnp)
	err := upnpMan.SearchGateway()
	if err != nil {
		return "", err
	} else {
		return upnpMan.LocalHost, nil
	}
}

// 获取指定网卡的ipv4地址,如WLAN
func GetIPv4ByInterfaceName(name string) (string, error) {
	inter, err := net.InterfaceByName(name)
	if err != nil {
		return "", err
	}
	addrs, err := inter.Addrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ip, ok := addr.(*net.IPNet); ok && !ip.IP.IsLoopback() {
			if ip.IP.To4() != nil {
				return ip.IP.String(), nil
			}
		}
	}
	return "", errors.New(name + " interface not found")
}

// 获取指定网卡的ipv6地址，如WLAN
func GetIPv6ByInterfaceName(name string) (string, error) {
	inter, err := net.InterfaceByName(name)
	if err != nil {
		return "", err
	}
	addrs, err := inter.Addrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ip, ok := addr.(*net.IPNet); ok && !ip.IP.IsLoopback() {
			if ip.IP.To16() != nil {
				return ip.IP.String(), nil
			}
		}
	}
	return "", errors.New(name + " interface not found")
}

func GetPublicIPV4() (string, error) {
	resp, err := http.Get("https://ipv4.netarm.com")
	//resp, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	content, _ := io.ReadAll(resp.Body)
	return strings.TrimSpace(string(content)), nil
}

func GetPublicIPV6() (string, error) {
	resp, err := http.Get("https://ipv6.netarm.com")
	//resp, err := http.Get("http://v6.ipv6-test.com/api/myip.php")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	content, _ := io.ReadAll(resp.Body)
	return strings.TrimSpace(string(content)), nil
}

// 初始化client
func Get_client() (http.Client, error) {
	// PublicSuffixList 是一种用来指定哪些后缀是公共后缀的列表。
	// 在 cookie 技术中，公共后缀是指可以被许多不同域名所拥有的后缀。
	// 例如，如果 .com 是公共后缀，那么 example.com 和 mywebsite.com 都可以设置和读取 .com 后缀的 cookie。
	// PublicSuffixList 是用来确保浏览器在设置和读取 cookie 时，使用的是正确的公共后缀。
	jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	return http.Client{Jar: jar}, nil
}

// 获取指定网卡的ipv6子网掩码
func GetIpv6MaskByInterfaceName(name string) (string, error) {
	inter, err := net.InterfaceByName(name)
	if err != nil {
		return "", err
	}
	addrs, err := inter.Addrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ip, ok := addr.(*net.IPNet); ok && !ip.IP.IsLoopback() {
			if ip.IP.To16() != nil {
				return ip.Mask.String(), nil
			}
		}
	}
	return "", errors.New("not found")
}

// 获取指定网卡的ipv4子网掩码
func GetIpv4MaskByInterfaceName(name string) (string, error) {
	inter, err := net.InterfaceByName(name)
	if err != nil {
		return "", err
	}
	addrs, err := inter.Addrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ip, ok := addr.(*net.IPNet); ok && !ip.IP.IsLoopback() {
			if ip.IP.To4() != nil {
				return ip.Mask.String(), nil
			}
		}
	}
	return "", errors.New("not found")
}

// 转换十六进制的子网掩码为点分十进制(请确保传入的是十六进制的子网掩码)
func HexMaskToDotMask(hexMask string) string {
	var dotMask string
	for i := 0; i < len(hexMask); i += 2 {
		num, _ := HexToInt(hexMask[i : i+2])
		dotMask += strconv.Itoa(num) + "."
	}
	return dotMask[:len(dotMask)-1]
}

// 转换十六进制的子网掩码为冒号分隔的十六进制(请确保传入的是十六进制的子网掩码)
func HexMaskToColonMask(hexMask string) string {
	var colonMask string
	for i := 0; i < len(hexMask); i += 4 {
		colonMask += hexMask[i:i+4] + ":"
	}
	return colonMask[:len(colonMask)-1]
}

// 获取本机真实的无线局域网的mac地址
func GetMyWLANMAC() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, inter := range interfaces {
		if strings.Contains(inter.Name, "WLAN") {
			return inter.HardwareAddr.String(), nil
		}
	}
	return "", errors.New("not found")
}

// 通过ipconfig命令获取WLAN的默认网关
func GetWLANDefaultGateway() (string, error) {
	cmd := exec.Command("ipconfig")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	out = GbkToUtf8(out)
	n := bytes.Index(out, []byte("WLAN"))
	//只要WLAN的信息
	if n+304 > len(out) {
		out = out[n:]
	} else {
		out = out[n : n+304]
	}
	//匹配out中最后一个ipv4地址作为子网掩码
	reg := regexp.MustCompile(`^[\s\S]+([0-9]{3}\.[0-9]+\.[0-9]+\.[0-9]+)`)
	macth := reg.FindSubmatch(out)
	if macth == nil {
		return "", errors.New("not found")
	}
	return string(macth[1]), nil
}

func NetWorkStatus() bool {
	timeout := time.Duration(time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get("https://www.baidu.com")
	if err != nil {
		log.Println("测试网络连接出现问题！", err)
		return false
	}
	defer resp.Body.Close()
	log.Println("Net Status , OK", resp.Status)
	return true
}
