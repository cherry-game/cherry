package cherryDiscovery

import (
	"context"
	"fmt"
	facade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryProto "github.com/cherry-game/cherry/net/proto"
	cherryProfile "github.com/cherry-game/cherry/profile"
	jsoniter "github.com/json-iterator/go"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/namespace"
	"strings"
	"time"
)

var (
	keyPrefix         = "/cherry/node/"
	registerKeyFormat = keyPrefix + "%s"
)

// DiscoveryETCD etcd方式发现服务
type DiscoveryETCD struct {
	facade.IApplication
	DiscoveryDefault
	prefix  string
	config  clientv3.Config
	ttl     int64
	cli     *clientv3.Client // etcd client
	leaseID clientv3.LeaseID // get lease id
}

func (p *DiscoveryETCD) Name() string {
	return "etcd"
}

func (p *DiscoveryETCD) Init(app facade.IApplication) {
	p.IApplication = app
	p.ttl = 10

	cfg := cherryProfile.Config().Get("cluster").Get(p.Name())
	if cfg.LastError() != nil {
		cherryLogger.Fatalf("etcd config not found. err = %v", cfg.LastError())
		return
	}
	p.loadConfig(cfg)
	p.register()
	p.getLeaseId()
	p.watch()

	cherryLogger.Infof("[etcd] init complete! [endpoints = %v] [leaseId = %d]", p.config.Endpoints, p.leaseID)
}

func (p *DiscoveryETCD) OnStop() {
	key := fmt.Sprintf(registerKeyFormat, p.NodeId())
	_, err := p.cli.Delete(context.Background(), key)
	cherryLogger.Infof("etcd stopping! err = %v", err)

	err = p.cli.Close()
	if err != nil {
		cherryLogger.Warnf("etcd stopping error! err = %v", err)
	}
}

func (p *DiscoveryETCD) loadConfig(config jsoniter.Any) {
	p.config = clientv3.Config{}

	p.config.Endpoints = strings.Split(config.Get("end_points").ToString(), ",")
	p.config.DialTimeout = time.Duration(config.Get("dial_timeout_second").ToInt64()) * time.Second
	if p.config.DialTimeout < 1*time.Second {
		p.config.DialTimeout = 3 * time.Second
	}
	p.config.Username = config.Get("user").ToString()
	p.config.Password = config.Get("password").ToString()

	p.ttl = config.Get("ttl").ToInt64()
	if p.ttl < 1 {
		p.ttl = 5
	}

	p.prefix = config.Get("prefix").ToString()
	if p.prefix == "" {
		p.prefix = "cherry"
	}
}

func (p *DiscoveryETCD) getLeaseId() {
	var err error
	//设置租约时间
	resp, err := p.cli.Grant(context.Background(), p.ttl)
	if err != nil {
		cherryLogger.Fatal(err)
		return
	}

	p.leaseID = resp.ID

	//设置续租 定期发送需求请求
	keepaliveChan, err := p.cli.KeepAlive(context.Background(), resp.ID)
	if err != nil {
		cherryLogger.Fatal(err)
		return
	}

	go func() {
		for {
			select {
			case <-keepaliveChan:
				{
				}
			case die := <-p.DieChan():
				{
					if die {
						return
					}
				}
			}
		}
	}()
}

func (p *DiscoveryETCD) register() {
	var err error
	p.cli, err = clientv3.New(p.config)
	if err != nil {
		cherryLogger.Fatalf("etcd connect fail. err = %v", err)
		return
	}

	// set namespace
	p.cli.KV = namespace.NewKV(p.cli.KV, p.prefix)
	p.cli.Watcher = namespace.NewWatcher(p.cli.Watcher, p.prefix)
	p.cli.Lease = namespace.NewLease(p.cli.Lease, p.prefix)

	registerMember := &cherryProto.Member{
		NodeId:   p.NodeId(),
		NodeType: p.NodeType(),
		Address:  p.RpcAddress(),
		Settings: make(map[string]string),
	}

	jsonString, err := jsoniter.MarshalToString(registerMember)
	if err != nil {
		cherryLogger.Fatal(err)
		return
	}

	key := fmt.Sprintf(registerKeyFormat, p.NodeId())

	_, err = p.cli.Put(context.Background(), key, jsonString, clientv3.WithLease(p.leaseID))
	if err != nil {
		cherryLogger.Fatal(err)
		return
	}
}

func (p *DiscoveryETCD) watch() {
	resp, err := p.cli.Get(context.Background(), keyPrefix, clientv3.WithPrefix())
	if err != nil {
		cherryLogger.Fatal(err)
		return
	}

	for _, ev := range resp.Kvs {
		p.addMember(ev.Value)
	}

	watchChan := p.cli.Watch(context.Background(), keyPrefix, clientv3.WithPrefix())
	go func() {
		for rsp := range watchChan {
			for _, ev := range rsp.Events {
				switch ev.Type {
				case mvccpb.PUT:
					{
						p.addMember(ev.Kv.Value)
					}
				case mvccpb.DELETE:
					{
						p.removeMember(ev.Kv)
					}
				}
			}
		}
	}()
}

func (p *DiscoveryETCD) addMember(data []byte) {
	member := &cherryProto.Member{}
	err := jsoniter.Unmarshal(data, member)
	if err != nil {
		return
	}

	if _, ok := p.GetMember(member.NodeId); ok == false {
		p.AddMember(member)
	}
}

func (p *DiscoveryETCD) removeMember(kv *mvccpb.KeyValue) {
	key := string(kv.Key)
	nodeId := strings.ReplaceAll(key, keyPrefix, "")
	if nodeId == "" {
		cherryLogger.Warn("remove member nodeId is empty!")
	}

	p.RemoveMember(nodeId)
}
