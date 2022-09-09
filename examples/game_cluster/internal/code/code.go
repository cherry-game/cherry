package code

var (
	OK                       int32 = 0   // is ok
	Error                    int32 = 1   // error
	PIDError                 int32 = 100 // pid错误
	LoginError               int32 = 201 // 登录异常
	AccountAuthFail          int32 = 202 // 帐号授权失败
	AccountBindFail          int32 = 203 // 帐号绑定失败
	AccountTokenValidateFail int32 = 204 // token验证失败
	AccountNameIsExist       int32 = 205 // 帐号已存在
	ActorDenyLogin           int32 = 301 // 角色禁止登录
	ActorDuplicateLogin      int32 = 302 // 角色重复登录
	ActorNameExist           int32 = 303 // 角色名已存在
	ActorCreateFail          int32 = 304 // 角色创建失败
	ActorNotLogin            int32 = 305 // 角色未登录
	ActorIdError             int32 = 306 // 角色id错误
)
