# data-config组件
- 自定义数据源
- 读取数据
- 热更新数据

## Install

### Prerequisites
- GO >= 1.17

### Using go get
```
go get github.com/cherry-game/cherry/components/data-config@latest
```


## Quick Start
```
import cherryDataConfig "github.com/cherry-game/cherry/components/data-config"
```


```go
// 注册struct到data-config
func RegisterComponent() {
    dataConfig := cherryDataConfig.NewComponent()
    dataConfig.Register(
        PlatformConfig,
        StatusCodeConfig,
        MaskWordConfig,
    )
    
    //data-config组件注册到cherry引擎
    cherry.RegisterComponent(dataConfig)
}

```

## example
- [示例代码跳转](https://github.com/cherry-game/cherry/tree/master/examples/test_data_config)