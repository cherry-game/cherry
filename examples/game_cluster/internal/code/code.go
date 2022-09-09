package code

var (
	OK                  int32 = 0   // is ok
	Error               int32 = 1   // error
	PIDError            int32 = 100 // pid错误
	LoginError          int32 = 201 // 登录异常
	AccountAuthFail     int32 = 202 // 帐号授权失败
	AccountBindFail     int32 = 203 // 帐号绑定失败
	TokenValidateFail   int32 = 204 // token验证失败
	ActorDenyLogin      int32 = 205 // 角色禁止登录
	ActorDuplicateLogin int32 = 206 // 角色重复登录
)
