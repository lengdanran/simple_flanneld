node1:172.16.10.245
#!/bin/bash
docker0_mac_addr="02:42:c4:1a:6e:67"
flannel_subnet="162.20.0.0/16"

# create vxlan device
ip link add vxlan0 type vxlan id 1 dstport 8472
# up the vxlan device
ip link set dev vxlan0 up
# connect vxlan0 to docker0 bridge
ip link set vxlan0 master docker0

# add other node sub net to vxlan0
#ip route add 162.20.4.0/24 dev vxlan0
ip route add 162.20.4.0/24 via 162.20.4.1 dev vxlan0 onlink
ip nei add 162.20.4.1 dev vxlan0 lladdr ${docker0_mac_addr}

bridge fdb add ${docker0_mac_addr} dev vxlan0 dst 172.16.10.244

iptables -t filter -P FORWARD ACCEPT
iptables -t nat -I POSTROUTING -d ${flannel_subnet} -j ACCEPT


node2:172.16.10.244
#!/bin/bash
docker0_mac_addr="02:42:d4:11:83:01"
flannel_subnet="162.20.0.0/16"

# create vxlan device
ip link add vxlan0 type vxlan id 1 dstport 8472
# up the vxlan device
ip link set dev vxlan0 up
# connect vxlan0 to docker0 bridge
ip link set vxlan0 master docker0

# add other node sub net to vxlan0
#ip route add 162.20.4.0/24 dev vxlan0
ip route add 162.20.5.0/24 via 162.20.5.1 dev vxlan0 onlink
ip nei add 162.20.5.1 dev vxlan0 lladdr ${docker0_mac_addr}

bridge fdb add ${docker0_mac_addr} dev vxlan0 dst 172.16.10.245

iptables -t filter -P FORWARD ACCEPT
iptables -t nat -I POSTROUTING -d ${flannel_subnet} -j ACCEPT






#!/bin/bash
docker0_mac_addr="02:42:c4:1a:6e:67"
flannel_subnet="162.12.0.0/16"

# create vxlan device
ip link add vxlan0 type vxlan id 1 dstport 8472
# up the vxlan device
ip link set dev vxlan0 up
# connect vxlan0 to docker0 bridge
ip link set vxlan0 master docker0

# add other node sub net to vxlan0
#ip route add 162.20.4.0/24 dev vxlan0
ip route add 162.16.2.0/24 via 162.16.2.1 dev vxlan0 onlink
ip nei add 162.16.2.1 dev vxlan0 lladdr ${docker0_mac_addr}

bridge fdb add ${docker0_mac_addr} dev vxlan0 dst 172.16.10.240

iptables -t filter -P FORWARD ACCEPT
iptables -t nat -I POSTROUTING -d ${flannel_subnet} -j ACCEPT