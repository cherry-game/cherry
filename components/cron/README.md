# cron组件
- 支持cron表达式
- 根据设定的时间规则定时执行函数

## Install

### Prerequisites
- GO >= 1.17

### Using go get
```
go get github.com/cherry-game/cherry/components/cron@latest
```


## Quick Start
```
import cherryCron "github.com/cherry-game/cherry/components/cron"
```


```go
// 以组件方式注入到cherry引擎
func Run(path, env, node string) {
    // 加载profile配置
    cherry.Configure(path, env, node)
    // cron以组件方式注册到cherry引擎
    cherryCron.RegisterComponent()
    // 启动cherry引擎
    cherry.Run(false, cherry.Cluster)
}

// 手工方式启动cron
func main() {
    cherryCron.Init()

    for i := 0; i <= 23; i++ {
        cherryCron.AddEveryDayFunc(func() {
            now := cherryTime.Now()
            cherryLogger.Infof("每天第%d点%d分%d秒运行", now.Hour(), now.Minute(), now.Second())
        }, i, 12, 34)
        cherryLogger.Infof("添加 每天第%d点执行的定时器", i)
    }

    for i := 0; i <= 59; i++ {
        cherryCron.AddEveryHourFunc(func() {
            cherryLogger.Infof("每小时第%d分执行一次", cherryTime.Now().Minute())
        }, i, 0)
        cherryLogger.Infof("添加 每小时第%d分的定时器", i)
    }

    cherryCron.Run()
}

```

## example
- [示例代码跳转](https://github.com/cherry-game/cherry/blob/master/components/cron/cron_test.go)