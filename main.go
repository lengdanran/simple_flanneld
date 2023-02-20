package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"simple_flannel/backend/vxlan"
	"simple_flannel/entity"
	"simple_flannel/etcd"
	"strings"
	"syscall"
)

type CmdLineOpts struct {
	etcdEndpoint         string
	etcdPrefix           string
	etcdUsername         string
	etcdPassword         string
	dockerNet            string
	flannelNetwork       string // flannel 主网络
	flannelSubnet        string // flannel 子网络
	flannelSubnetGateway string
	publicIP             string // ip地址
}

var opts CmdLineOpts

func printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func init() {
	//flag.StringVar(&opts.etcdEndpoint, "etcd-endpoint", "http://localhost:2379", "etcd数据库的连接地址")
	//flag.StringVar(&opts.etcdPrefix, "etcd-prefix", "/flannel.com/network", "etcd数据读取前缀")
	//flag.StringVar(&opts.etcdUsername, "etcd-username", "", "etcd用户名")
	//flag.StringVar(&opts.etcdPassword, "etcd-password", "", "etcd密码")
	//flag.StringVar(&opts.flannelNetwork, "flannel-network", "162.16.0.0/16", "flannel主网络")
	//flag.StringVar(&opts.flannelSubnet, "flannel-subnet", "162.16.1.0/24", "flannel子网络")
	//flag.StringVar(&opts.flannelSubnetGateway, "flannel-subnet-gateway", "162.16.1.1", "flannel子网络网关地址")
	//flag.StringVar(&opts.publicIP, "public-ip", "127.0.0.1", "ip地址")

	configuration := readConfiguration("./flannel.properties")

	opts.etcdEndpoint = configuration["etcd-endpoint"]
	opts.etcdPrefix = configuration["etcd-prefix"]
	opts.etcdUsername = configuration["etcd-username"]
	opts.etcdPassword = configuration["etcd-password"]
	opts.dockerNet = configuration["docker-net"]
	opts.flannelNetwork = configuration["flannel-network"]
	opts.flannelSubnet = configuration["flannel-subnet"]
	opts.flannelSubnetGateway = configuration["flannel-subnet-gateway"]
	opts.publicIP = configuration["public-ip"]

	parseArgs()
}

func readConfiguration(configurationFile string) map[string]string {
	var properties = make(map[string]string)
	confFile, err := os.OpenFile(configurationFile, os.O_RDONLY, 0666)
	defer func(confFile *os.File) {
		if err := confFile.Close(); err != nil {
			panic(err)
		}
	}(confFile)
	if err != nil {
		printf("The config file %s is not exits.", configurationFile)
	} else {
		reader := bufio.NewReader(confFile)
		for {
			if confString, err := reader.ReadString('\n'); err != nil {
				if err == io.EOF {
					break
				}
			} else {
				if len(confString) == 0 || confString == "\n" || confString[0] == '#' {
					continue
				}
				properties[strings.Split(confString, "=")[0]] = strings.Replace(strings.Split(confString, "=")[1], "\n", "", -1)
			}
		}
	}
	return properties
}

func parseArgs() {
	//flag.Parse()
	printf("==================================== Starting Flanneld ===================================\n")
	printf("        etcd-endpoints:%s\n", opts.etcdEndpoint)
	printf("           etcd-prefix:%s\n", opts.etcdPrefix)
	printf("         etcd-username:%s\n", opts.etcdUsername)
	printf("         etcd-password:%s\n", opts.etcdPassword)
	printf("       flannel-network:%s\n", opts.flannelNetwork)
	printf("        flannel-subnet:%s\n", opts.flannelSubnet)
	printf("flannel-subnet-gateway:%s\n", opts.flannelSubnetGateway)
	printf("             public-ip:%s\n", opts.publicIP)
	printf("=========================================================================================\n")
}

func runCommand(cmd string) {
	printf("[RunCmd]:%s\n", cmd)
	command := exec.Command("/bin/bash", "-c", cmd)
	if _, err := command.CombinedOutput(); err != nil {
		printf("[ERROR]:%s\n", err)
		panic(err)
	}
}

func runCommandWithRsp(cmd string) string {
	printf("[RunCmd]:%s\n", cmd)
	command := exec.Command("/bin/bash", "-c", cmd)
	if rsp, err := command.CombinedOutput(); err != nil {
		printf("[ERROR]:%s\n", err)
		panic(err)
	} else {
		return string(rsp)
	}
}

func rewriteDockerDaemonJson() {
	daemonJsonStr := `"{
    \"bip\":\"` + opts.dockerNet + `\"
}"`
	runCommand("echo " + daemonJsonStr + " > /etc/docker/daemon.json")
}

func restartDocker() {
	runCommand("systemctl restart docker")
}

func setDockerBipNet() {
	rewriteDockerDaemonJson()
	restartDocker()
}

func initIptablesConfig() {
	runCommand("iptables -t filter -P FORWARD ACCEPT")
	runCommand(fmt.Sprintf("iptables -t nat -I POSTROUTING -d %s -j ACCEPT", opts.flannelNetwork))
}

func getLocalMacAddr(devName string) string {
	//rsp := runCommandWithRsp("ip addr show " + devName)
	//lines := strings.Split(rsp, "\n")
	//MacAddr := strings.Split(strings.TrimSpace(lines[1]), " ")[1]
	interfaces, _ := net.InterfaceByName(devName)
	MacAddr := fmt.Sprintf("%v", interfaces.HardwareAddr)
	printf("GetMacAddr: %s\n", MacAddr)
	return MacAddr
	//for i, line := range lines {
	//	if strings.HasPrefix(line, devName) {
	//		t := lines[i+2]
	//		t = strings.TrimSpace(t)
	//		printf("Get Docker0 MAC: %s\n", strings.Split(t, " ")[1])
	//		return strings.Split(t, " ")[1]
	//	}
	//}
	//panic("MAC is not found!!")
}

func getLocalNodeNetwork() entity.NodeNetwork {
	return entity.NodeNetwork{
		IpAddr:         opts.publicIP,
		Docker0MacAddr: getLocalMacAddr("docker0"),
		Cidr:           opts.flannelSubnet,
		Gateway:        opts.flannelSubnetGateway,
	}
}

func addExitsNodes(Cli *clientv3.Client, vx *vxlan.VXLANDevice) {
	if exitsNodes, err := etcd.GetWithPrefix(opts.etcdPrefix, Cli); err != nil {
		printf("addExitsNodes ERROR occurred.....")
		panic(err)
	} else {
		for _, v := range exitsNodes {
			node := entity.NodeNetwork{}
			printf("Add exits node %s\n", v)
			_ = json.Unmarshal([]byte(v), &node)
			vx.AddNewNodeNetwork(node.Cidr, node.Gateway, node.Docker0MacAddr, node.IpAddr)
		}
	}
}

func SetupCloseHandler(Cli *clientv3.Client, vx *vxlan.VXLANDevice) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func(Cli *clientv3.Client, vx *vxlan.VXLANDevice) {
		<-c
		printf("\r- Ctrl+C pressed in Terminal")
		runCommand(fmt.Sprintf("ip link del %s", vx.Name))
		etcd.Del(opts.etcdPrefix+"/"+opts.publicIP, Cli)
		printf("Del the vxlan device....")
		os.Exit(0)
	}(Cli, vx)
}

func main() {
	setDockerBipNet()
	initIptablesConfig()
	vx := vxlan.NewVXLANDevice("vxlan0", 1, 8472, "docker0")
	val, _ := json.Marshal(getLocalNodeNetwork())
	cli := etcd.Connect(opts.etcdEndpoint)
	SetupCloseHandler(cli, vx)
	vx.Create()
	addExitsNodes(cli, vx)
	etcd.Put(opts.etcdPrefix+"/"+opts.publicIP, string(val), cli)
	etcd.WatchPrefix(opts.etcdPrefix, vx, opts.publicIP, cli)
}
