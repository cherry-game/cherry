package cherryConnector

//// ConnectStat 连接统计
//type ConnectStat struct {
//	connCount  int64 // 连接总数
//	loginCount int64 // 登陆总数
//}
//
//func (c *ConnectStat) IncreaseLogin() {
//	atomic.AddInt64(&c.loginCount, 1)
//}
//
//func (c *ConnectStat) DecreaseLogin() {
//	atomic.AddInt64(&c.loginCount, -1)
//}
//
//func (c *ConnectStat) IncreaseConn() {
//	atomic.AddInt64(&c.connCount, 1)
//}
//
//func (c *ConnectStat) DecreaseConn() {
//	atomic.AddInt64(&c.connCount, -1)
//}
//
//func (c *ConnectStat) ConnCount() int64 {
//	return c.connCount
//}
//
//func (c *ConnectStat) LoginCount() int64 {
//	return c.loginCount
//}
//
//func (c *ConnectStat) PrintInfo() string {
//	return fmt.Sprintf("connCount = %d, loginCount = %d", c.connCount, c.loginCount)
//}
