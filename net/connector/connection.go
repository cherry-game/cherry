package cherryConnector

import (
	"github.com/cherry-game/cherry/interfaces"
	"sync"
	"sync/atomic"
)

// LoginRecord 登陆记录器
type LoginRecord struct {
	loginTime int                  // login time
	uid       cherryInterfaces.UID // uid
	ip        string               // ip address
}

// Connection 连接统计
type Connection struct {
	sync.RWMutex
	connCount    int32                                 // connection count
	loginCount   int32                                 // user login count
	loginRecords map[cherryInterfaces.UID]*LoginRecord // user login record info
}

func (c *Connection) Add(info *LoginRecord) {
	c.Lock()
	defer c.Unlock()

	if _, found := c.loginRecords[info.uid]; !found {
		c.loginCount += 1
	}
	c.loginRecords[info.uid] = info
}

func (c *Connection) Remove(uid cherryInterfaces.UID) {
	c.Lock()
	defer c.Unlock()

	if _, found := c.loginRecords[uid]; found {
		c.loginCount--
	}
	delete(c.loginRecords, uid)
}

func (c *Connection) IncreaseConn() {
	atomic.AddInt32(&c.connCount, 1)
}

func (c *Connection) DecreaseConn() {
	atomic.AddInt32(&c.connCount, -1)
}

func (c *Connection) List() (connCount int32, loginCount int32, loginRecords map[cherryInterfaces.UID]*LoginRecord) {
	return c.connCount, c.loginCount, c.loginRecords
}
