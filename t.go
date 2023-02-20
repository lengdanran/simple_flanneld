package main

import (
	"fmt"
	"net"
)

func main() {
	interfaces, _ := net.InterfaceByName("docker0")
	fmt.Printf("%v", interfaces.HardwareAddr)
}
