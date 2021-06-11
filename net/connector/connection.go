package cherryConnector

import (
	facade "github.com/cherry-game/cherry/facade"
	"sync"
	"sync/atomic"
)

// LoginRecord 登录记录器
type LoginRecord struct {
	loginTime int        // login time
	uid       facade.UID // uid
	ip        string     // ip address
}

// ConnectStat 连接统计
type ConnectStat struct {
	sync.RWMutex
	connCount    int32                       // 连接总数
	loginCount   int32                       // 登陆总数
	loginRecords map[facade.UID]*LoginRecord // 用户登陆记录
}

func (c *ConnectStat) Add(info *LoginRecord) {
	c.Lock()
	defer c.Unlock()

	if _, found := c.loginRecords[info.uid]; !found {
		c.loginCount += 1
	}
	c.loginRecords[info.uid] = info
}

func (c *ConnectStat) Remove(uid facade.UID) {
	c.Lock()
	defer c.Unlock()

	if _, found := c.loginRecords[uid]; found {
		c.loginCount--
	}
	delete(c.loginRecords, uid)
}

func (c *ConnectStat) IncreaseConn() {
	atomic.AddInt32(&c.connCount, 1)
}

func (c *ConnectStat) DecreaseConn() {
	atomic.AddInt32(&c.connCount, -1)
}

func (c *ConnectStat) ConnCount() int32 {
	return c.connCount
}

func (c *ConnectStat) LoginCount() int32 {
	return c.loginCount
}

func (c *ConnectStat) LoginRecords() map[facade.UID]*LoginRecord {
	return c.loginRecords
}
