package vxlan

import (
	"fmt"
	"log"
	"os/exec"
)

// VXLANDevice vxlan设备
type VXLANDevice struct {
	Name         string
	Vni          int
	DstPort      int
	Mtu          int
	MasterBridge string
}

// NewVXLANDevice : 构建Vxlan结构体
func NewVXLANDevice(Name string, Vni int, DstPort int, MasterBridge string) *VXLANDevice {
	return &VXLANDevice{Name: Name, Vni: Vni, DstPort: DstPort, Mtu: 1450, MasterBridge: MasterBridge}
}

func printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func runCommand(cmd string) {
	printf("[RunCmd]:%s\n", cmd)
	command := exec.Command("/bin/bash", "-c", cmd)
	if _, err := command.CombinedOutput(); err != nil {
		printf("[ERROR]:%s\n", err)
		panic(err)
	}
}

func (vx *VXLANDevice) Create() {
	cmdStr := fmt.Sprintf("ip link add %s type vxlan id %d dstport %d", vx.Name, vx.Vni, vx.DstPort)
	runCommand(cmdStr)
	cmdStr = fmt.Sprintf("ip link set dev %s mtu %d", vx.Name, vx.Mtu)
	runCommand(cmdStr)
	cmdStr = fmt.Sprintf("ip link set dev %s up", vx.Name)
	runCommand(cmdStr)
	cmdStr = fmt.Sprintf("ip link set %s master %s", vx.Name, vx.MasterBridge)
	runCommand(cmdStr)
	printf("The vxlan device has successfully created.\n")
}

func (vx *VXLANDevice) AddNewNodeNetwork(cidr string, gateway string, docker0MacAddr string, ipAddr string) {
	printf("Add the new Node to the flannel network：cidr: %s, gateway: %s, docker0MacAddr: %s, ipAddr: %s\n", cidr, gateway, docker0MacAddr, ipAddr)
	// 配置路由
	runCommand(fmt.Sprintf("ip route add %s via %s dev %s onlink", cidr, gateway, vx.Name))
	// 配置arp表
	runCommand(fmt.Sprintf("ip nei add %s dev %s lladdr %s", gateway, vx.Name, docker0MacAddr))
	// 配置FDB表
	runCommand(fmt.Sprintf("bridge fdb add %s dev %s dst %s", docker0MacAddr, vx.Name, ipAddr))
}

func (vx *VXLANDevice) DelNewNodeNetwork(cidr string, gateway string, docker0MacAddr string, ipAddr string) {
	printf("Delete the new Node to the flannel network：cidr: %s, gateway: %s, docker0MacAddr: %s, ipAddr: %s\n", cidr, gateway, docker0MacAddr, ipAddr)
	// 删除路由
	runCommand(fmt.Sprintf("ip route del %s via %s dev %s onlink", cidr, gateway, vx.Name))
	// 删除arp表配置
	runCommand(fmt.Sprintf("ip nei del %s dev %s lladdr %s", gateway, vx.Name, docker0MacAddr))
	// 删除FDB表配置
	runCommand(fmt.Sprintf("bridge fdb del %s dev %s dst %s", docker0MacAddr, vx.Name, ipAddr))
}
