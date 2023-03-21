package main

// UserBindTable uid绑定第三方平台表
type UserBindTable struct {
	UID      int64  `gorm:"column:uid;primary_key;comment:'用户唯一id'" json:"uid"`
	SdkId    int32  `gorm:"column:sdk_id;comment:'sdk配置id'" json:"sdkId"`
	PID      int32  `gorm:"column:pid;comment:'平台id'" json:"pid"`
	OpenId   string `gorm:"column:open_id;comment:'平台帐号open_id'" json:"openId"`
	BindTime int64  `gorm:"column:bind_time;comment:'绑定时间'" json:"bindTime"`
}

func (*UserBindTable) TableName() string {
	return "user_bind"
}

func (p *UserBindTable) PrimaryKey() interface{} {
	return p.UID
}
