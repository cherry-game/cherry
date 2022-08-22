package cherryDiscovery

//import (
//	"context"
//	"fmt"
//	cfacade "github.com/cherry-game/cherry/facade"
//	clog "github.com/cherry-game/cherry/logger"
//	cproto "github.com/cherry-game/cherry/net/proto"
//	cprofile "github.com/cherry-game/cherry/profile"
//	jsoniter "github.com/json-iterator/go"
//	"go.etcd.io/etcd/api/v3/mvccpb"
//	"go.etcd.io/etcd/client/v3"
//	"go.etcd.io/etcd/client/v3/namespace"
//	"strings"
//	"time"
//)
//
//var (
//	keyPrefix         = "/cherry/node/"
//	registerKeyFormat = keyPrefix + "%s"
//)
//
//// DiscoveryETCD etcd方式发现服务
//type DiscoveryETCD struct {
//	cfacade.IApplication
//	DiscoveryDefault
//	prefix  string
//	config  clientv3.Config
//	ttl     int64
//	cli     *clientv3.Client // etcd client
//	leaseID clientv3.LeaseID // get lease id
//}
//
//func (p *DiscoveryETCD) Name() string {
//	return "etcd"
//}
//
//func (p *DiscoveryETCD) Init(app cfacade.IApplication) {
//	p.IApplication = app
//	p.ttl = 10
//
//	clusterConfig := cprofile.GetConfig("cluster").GetConfig(p.Name())
//	if clusterConfig.LastError() != nil {
//		clog.Fatalf("etcd config not found. err = %v", clusterConfig.LastError())
//		return
//	}
//
//	p.loadConfig(clusterConfig)
//	p.init()
//	p.getLeaseId()
//	p.register()
//	p.watch()
//
//	clog.Infof("[etcd] init complete! [endpoints = %v] [leaseId = %d]", p.config.Endpoints, p.leaseID)
//}
//
//func (p *DiscoveryETCD) OnStop() {
//	key := fmt.Sprintf(registerKeyFormat, p.NodeId())
//	_, err := p.cli.Delete(context.Background(), key)
//	clog.Infof("etcd stopping! err = %v", err)
//
//	err = p.cli.Close()
//	if err != nil {
//		clog.Warnf("etcd stopping error! err = %v", err)
//	}
//}
//
//func getDialTimeout(config jsoniter.Any) time.Duration {
//	t := time.Duration(config.Get("dial_timeout_second").ToInt64()) * time.Second
//	if t < 1*time.Second {
//		t = 3 * time.Second
//	}
//
//	return t
//}
//
//func getEndPoints(config jsoniter.Any) []string {
//	return strings.Split(config.Get("end_points").ToString(), ",")
//}
//
//func (p *DiscoveryETCD) loadConfig(config cfacade.JsonConfig) {
//	p.config = clientv3.Config{
//		Logger: clog.DefaultLogger.Desugar(),
//	}
//
//	p.config.Endpoints = getEndPoints(config)
//	p.config.DialTimeout = getDialTimeout(config)
//	p.config.Username = config.GetString("user")
//	p.config.Password = config.GetString("password")
//	p.ttl = config.GetInt64("ttl", 5)
//	p.prefix = config.GetString("prefix", "cherry")
//}
//
//func (p *DiscoveryETCD) init() {
//	var err error
//	p.cli, err = clientv3.New(p.config)
//	if err != nil {
//		clog.Fatalf("etcd connect fail. err = %v", err)
//		return
//	}
//
//	// set namespace
//	p.cli.KV = namespace.NewKV(p.cli.KV, p.prefix)
//	p.cli.Watcher = namespace.NewWatcher(p.cli.Watcher, p.prefix)
//	p.cli.Lease = namespace.NewLease(p.cli.Lease, p.prefix)
//}
//
//func (p *DiscoveryETCD) getLeaseId() {
//	var err error
//	//设置租约时间
//	resp, err := p.cli.Grant(context.Background(), p.ttl)
//	if err != nil {
//		clog.Fatal(err)
//		return
//	}
//
//	p.leaseID = resp.ID
//
//	//设置续租 定期发送需求请求
//	keepaliveChan, err := p.cli.KeepAlive(context.Background(), resp.ID)
//	if err != nil {
//		clog.Fatal(err)
//		return
//	}
//
//	go func() {
//		for {
//			select {
//			case <-keepaliveChan:
//				{
//				}
//			case die := <-p.DieChan():
//				{
//					if die {
//						return
//					}
//				}
//			}
//		}
//	}()
//}
//
//func (p *DiscoveryETCD) register() {
//	registerMember := &cproto.Member{
//		NodeId:   p.NodeId(),
//		NodeType: p.NodeType(),
//		Address:  p.RpcAddress(),
//		Settings: make(map[string]string),
//	}
//
//	jsonString, err := jsoniter.MarshalToString(registerMember)
//	if err != nil {
//		clog.Fatal(err)
//		return
//	}
//
//	key := fmt.Sprintf(registerKeyFormat, p.NodeId())
//	_, err = p.cli.Put(context.Background(), key, jsonString, clientv3.WithLease(p.leaseID))
//	if err != nil {
//		clog.Fatal(err)
//		return
//	}
//}
//
//func (p *DiscoveryETCD) watch() {
//	resp, err := p.cli.Get(context.Background(), keyPrefix, clientv3.WithPrefix())
//	if err != nil {
//		clog.Fatal(err)
//		return
//	}
//
//	for _, ev := range resp.Kvs {
//		p.addMember(ev.Value)
//	}
//
//	watchChan := p.cli.Watch(context.Background(), keyPrefix, clientv3.WithPrefix())
//	go func() {
//		for rsp := range watchChan {
//			for _, ev := range rsp.Events {
//				switch ev.Type {
//				case mvccpb.PUT:
//					{
//						p.addMember(ev.Kv.Value)
//					}
//				case mvccpb.DELETE:
//					{
//						p.removeMember(ev.Kv)
//					}
//				}
//			}
//		}
//	}()
//}
//
//func (p *DiscoveryETCD) addMember(data []byte) {
//	member := &cproto.Member{}
//	err := jsoniter.Unmarshal(data, member)
//	if err != nil {
//		return
//	}
//
//	if _, ok := p.GetMember(member.NodeId); ok == false {
//		p.AddMember(member)
//	}
//}
//
//func (p *DiscoveryETCD) removeMember(kv *mvccpb.KeyValue) {
//	key := string(kv.Key)
//	nodeId := strings.ReplaceAll(key, keyPrefix, "")
//	if nodeId == "" {
//		clog.Warn("remove member nodeId is empty!")
//	}
//
//	p.RemoveMember(nodeId)
//}
