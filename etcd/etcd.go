package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/client/v3"
	"log"
	"simple_flannel/backend/vxlan"
	"simple_flannel/entity"
	"time"
)

// Cli client of the etcd
var Cli *clientv3.Client

//
//  getClient
//  @Description: 获取etcd的连接客户端
//
func getClientWithEndpoint(endpoint string) *clientv3.Client {
	log.Println(">> Connect to " + endpoint)
	if cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{endpoint},
		//DialTimeout: time.Duration(config.ServerConfig.Datasource.Dialtimeout) * time.Second,
	}); err != nil {
		panic("连接etcd失败，ERROR:" + err.Error())
		return nil
	} else {
		return cli
	}
}

func Connect(endpoint string) *clientv3.Client {
	return getClientWithEndpoint(endpoint)
}

// Put 添加etcd， k,v
func Put(k string, v string, Cli *clientv3.Client) error {
	if k == "" || v == "" {
		return errors.New("key or value is empty, do nothing")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	_, err := Cli.Put(ctx, k, v)
	cancel()
	if err != nil {
		return err
	}
	return nil
}

func Del(k string, Cli *clientv3.Client) {
	kv := clientv3.NewKV(Cli)
	if _, err := kv.Delete(context.TODO(), k); err != nil {
		panic(err)
	}
}

// GetWithPrefix 获取所有包含指定前缀的数据项，返回json字符串数组
func GetWithPrefix(prefix string, Cli *clientv3.Client) (map[string]string, error) {
	kv := clientv3.NewKV(Cli)
	if values, err := kv.Get(context.TODO(), prefix, clientv3.WithPrefix()); err != nil {
		return nil, err
	} else {
		var res = make(map[string]string)
		for _, ev := range values.Kvs {
			res[string(ev.Key)] = string(ev.Value)
		}
		//log.Printf(">> Get %d %s items\n", len(res), prefix)
		return res, nil
	}
}

func WatchPrefix(prefix string, vx *vxlan.VXLANDevice, localIp string, Cli *clientv3.Client) {
	watcher := clientv3.NewWatcher(Cli)
	ctx, _ := context.WithCancel(context.TODO())
	watchChan := watcher.Watch(ctx, prefix, clientv3.WithPrefix(), clientv3.WithPrevKV())
	for w := range watchChan {
		for _, e := range w.Events {
			switch e.Type {
			case mvccpb.PUT:
				node := entity.NodeNetwork{}
				if e.PrevKv == nil { // 新增节点
					_ = json.Unmarshal(e.Kv.Value, &node)
					if node.IpAddr != localIp {
						println("New Node Added to the flannel network......")
						vx.AddNewNodeNetwork(node.Cidr, node.Gateway, node.Docker0MacAddr, node.IpAddr)
					}
				} else { // 节点更新
					_ = json.Unmarshal(e.PrevKv.Value, &node)
					if node.IpAddr != localIp {
						println("New Node Updated to the flannel network......")
						vx.DelNewNodeNetwork(node.Cidr, node.Gateway, node.Docker0MacAddr, node.IpAddr)
						_ = json.Unmarshal(e.Kv.Value, &node)
						vx.AddNewNodeNetwork(node.Cidr, node.Gateway, node.Docker0MacAddr, node.IpAddr)
					}
				}
			case mvccpb.DELETE:
				node := entity.NodeNetwork{}
				println("=====" + string(e.PrevKv.Value))
				_ = json.Unmarshal(e.PrevKv.Value, &node)
				if node.IpAddr != localIp {
					println("Delete node from flannel network.......")
					vx.DelNewNodeNetwork(node.Cidr, node.Gateway, node.Docker0MacAddr, node.IpAddr)
				}
			}
		}
	}
}
