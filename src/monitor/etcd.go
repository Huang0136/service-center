package monitor

import (
	"constants"
	"log"

	"time"

	"github.com/coreos/etcd/clientv3"
	"golang.org/x/net/context"
)

func InitMonitorEtcd() {
	go monitorServerNode()
}

// 监听Etcd
func monitorServerNode() {
	log.Println("正在监听Etcd服务配置信息...")

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{constants.Configs["etcd.url"]},
		DialTimeout: 1 * time.Minute})
	if err != nil {
		log.Fatalln("建立etcd连接失败,", err)
	}

	constants.MonitorEtcdChan <- true

	// 一直监听
	for {
		watchChan := cli.Watch(context.Background(), "servers", clientv3.WithPrefix())
		for wc := range watchChan {
			for _, e := range wc.Events {
				saveMonitorServerNodeInfo(e.Type.String(), string(e.Kv.Key), string(e.Kv.Value))
			}
		}
	}

	log.Println("监听Etcd服务配置信息完成")
}

// 保存监听到的服务节点信息
func saveMonitorServerNodeInfo(opType, key, value string) {
	log.Printf("监听到服务节点变化,信息详情,type:%s,key:%s,value:%s", opType, key, value)
	constants.TranEtcdServerInfo(opType, key, value)
}
