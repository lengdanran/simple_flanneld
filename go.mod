module simple_flannel

go 1.16
// go env -w GOPROXY=https://goproxy.io,direct
require (
	go.etcd.io/etcd/api/v3 v3.5.5 // indirect
	go.etcd.io/etcd/client/v3 v3.5.5
	golang.org/x/sys v0.0.0-20220928140112-f11e5e49a4ec // indirect
)
