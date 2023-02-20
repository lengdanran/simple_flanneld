# ip

## vxlan

ip 命令创建我们的 vxlan interface：
```shell
$ ip link add vxlan0 type vxlan \
id 42 \
dstport 4789 \
remote 192.168.8.101 \
local 192.168.8.100 \
dev enp0s8

# --- example

ip link add vxlan0 type vxlan id 42 dstport 4789 
```
上面这条命令创建一个名字为 vxlan0，类型为 vxlan 的网络 interface，后面是 vxlan interface 需要的参数：

- `id` 42：指定 VNI 的值，这个值可以在 1 到 2^24 之间
- `dstport`：vtep 通信的端口，linux 默认使用 8472（为了保持兼容，默认值一直没有更改），而 IANA 分配的端口是 4789，所以我们这里显式指定了它的值
- `remote` 192.168.8.101：对方 vtep 的地址，类似于点对点协议
- `local` 192.168.8.100：当前节点 vtep 要使用的 IP 地址
- `dev enp0s8`：当节点用于 vtep 通信的网卡设备，用来读取 IP 地址。注意这个参数和 local 参数含义是相同的，在这里写出来是为了告诉大家有两个参数存在

执行完之后，系统就会创建一个名字为 vxlan0 的 interface，可以用 `ip -d link` 查看它的详细信息：



ip link add vxlan100 type vxlan id 100 dstport 8472 local 121.4.195.203 nol

ip link set up vxlan100


## ARP

https://www.w3cschool.cn/linuxc/linuxc-qf1p3l72.html

ip neighbor change 172.17.1.0 lladdr d6:47:2f:96:17:cc dev vxlan100 nud permanent





ip monitor all dev vxlan100

ip link add vxlan100 type vxlan id 100 dstport 8472 dev eth0
ip link set vxlan100 master docker0
ip link set vxlan100 up

fdb
eth0
bridge fdb append 00:16:3e:19:36:c1 dev vxlan100 dst 47.98.166.246
bridge fdb append 52:54:00:a9:c6:dc dev vxlan100 dst 121.4.195.203
vxlan100
bridge fdb append 9e:40:a6:c0:09:3c dev vxlan100 dst 47.98.166.246
bridge fdb append e6:71:40:0b:d6:71 dev vxlan100 dst 121.4.195.203

route
ip route add 172.17.1.0/24 via 172.17.1.1 dev vxlan100 onlink
ip route add 172.17.2.0/24 via 172.17.2.1 dev vxlan100 onlink

arp
ip neighbor add 172.17.1.1 lladdr 9e:40:a6:c0:09:3c dev vxlan100 nud permanent
ip neighbor add 172.17.2.1 lladdr e6:71:40:0b:d6:71 dev vxlan100 nud permanent


腾讯
ip link del vxlan100
ip link add vxlan100 type vxlan id 100 dstport 8472 local 172.17.2.1 nolearning
ip link set vxlan100 master docker0
ip link set vxlan100 up
ip route add 172.17.1.0/24 via 172.17.1.1 dev vxlan100 onlink
ip neighbor add 172.17.1.1 lladdr 00:16:3e:19:36:c1 dev vxlan100 nud permanent
ip neighbor add 172.30.201.121 lladdr 00:16:3e:19:36:c1 dev vxlan100 nud permanent
bridge fdb append 00:16:3e:19:36:c1 dev vxlan100 dst 47.98.166.246


ip neighbor add 172.17.1.1 lladdr <eth0-mac> dev vxlan100 nud permanent
ip neighbor add 172.30.201.121 lladdr <eth0-mac> dev vxlan100 nud permanent
bridge fdb append <eth0-mac> dev vxlan100 dst 47.98.166.246

阿里
ip link del vxlan100
ip link add vxlan100 type vxlan id 100 dstport 8472 local 172.17.0.1 nolearning
ip link set vxlan100 master docker0
ip link set vxlan100 up
ip route add 172.17.1.0/24 via 172.17.1.1 dev vxlan100 onlink
ip neighbor add 172.17.1.1 lladdr 00:16:3e:19:36:c1 dev vxlan100 nud permanent
bridge fdb append 00:16:3e:19:36:c1 dev vxlan100 dst 172.30.201.121


ip link del vxlan100
ip link add vxlan100 type vxlan id 100 dstport 8472 local 172.17.1.1 nolearning
ip link set vxlan100 master docker0
ip link set vxlan100 up
ip route add 172.17.0.0/24 via 172.17.0.1 dev vxlan100 onlink
ip neighbor add 172.17.0.1 lladdr 00:16:3e:08:d1:b1 dev vxlan100 nud permanent
bridge fdb append 00:16:3e:08:d1:b1 dev vxlan100 dst 172.16.10.241



ip neighbor add 172.17.2.1 lladdr <eth0-mac> dev vxlan100 nud permanent
ip neighbor add 172.17.16.13 lladdr <eth0-mac> dev vxlan100 nud permanent
bridge fdb append <eth0-mac> dev vxlan100 dst 121.4.195.203



ip link add vxlan100 type vxlan id 100 dstport 8472 dev eth0
brctl addif docker0 vxlan100






#!/bin/bash
sysctl -w net.ipv4.ip_forward=1
ip link del vxlan100
ip link add vxlan100 type vxlan id 100 dstport 8472 dev eth0 nolearning proxy
ip link set vxlan100 master docker0
ip link set vxlan100 up

bridge fdb append 00:00:00:00:00:00 dev vxlan100 dst 172.16.10.241
bridge fdb append 00:00:00:00:00:00 dev vxlan100 dst 172.16.10.243

bridge fdb append 02:42:ac:11:00:02 dev vxlan100 dst 172.16.10.241
bridge fdb append 02:42:ac:11:01:02 dev vxlan100 dst 172.16.10.243

ip neighbor add 172.17.0.2 lladdr 02:42:ac:11:00:02 dev vxlan100
ip neighbor add 172.17.1.2 lladdr 02:42:ac:11:01:02 dev vxlan100






#!/bin/bash
sysctl -w net.ipv4.ip_forward=1
ip link del vxlan100
ip link add vxlan100 type vxlan id 100 dstport 8472
ip link set vxlan100 master docker0
ip link set vxlan100 up
ip addr add 172.17.0.0/32 dev vxlan100

bridge fdb append 00:00:00:00:00:00 dev vxlan100 dst 172.16.10.241
bridge fdb append 00:00:00:00:00:00 dev vxlan100 dst 172.16.10.243

bridge fdb append 02:42:ac:11:00:02 dev vxlan100 dst 172.16.10.241
bridge fdb append 02:42:ac:11:01:02 dev vxlan100 dst 172.16.10.243

ip neighbor add 172.17.0.2 lladdr 02:42:ac:11:00:02 dev vxlan100
ip neighbor add 172.17.1.2 lladdr 02:42:ac:11:01:02 dev vxlan100









