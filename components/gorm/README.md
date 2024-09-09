# gorm组件
- 基于gorm实现多数据实例管理

## Install

### Prerequisites
- GO >= 1.18

### Using go get
```
go get github.com/cherry-game/cherry/components/gorm@latest
```


## Quick Start
```
import cherryGORM "github.com/cherry-game/cherry/components/gorm"
```

## example
- 请查看 `examples/demo_gorm`


## Update
- 数据库配置信息中添加`dsn`属性，未填写该属性时使用默认的`dsn`连接字符串 