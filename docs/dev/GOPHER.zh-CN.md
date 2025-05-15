<div style="display: flex; align-items: center;">
  <img 
    src="../images/logo.svg" 
    alt="Logo" 
    style="height: 4em; width: auto; vertical-align: middle; margin-right: 10px;" 
  />
  <h1 style="margin: 0; font-size: 24px; line-height: 1.5;">Moon - 多领域监控告警平台</h1>
</div>

| [English](GOPHER.md) | [中文简体](GOPHER.zh-CN.md) |

# 开发者必看

## 开发规范

### 1. 包导入规范

引入包的时候， 不管是否只有一个外部包还是多个外部包，统一使用如下写法

```go
import (
	"fmt"
)
```

在同时包含多方包导入时候， 导入顺序按照以下顺序导入， 中间用空行分隔

```go
import (
  // 空白标识符导入
  _ "gorm.io/driver/sqlite"
  
  // 标准库
  "fmt"
  
  // 第三方包
  "github.com/xxxx/xxxx"
  
  // 项目内部包
  "github.com/moon-monitor/moon/internal/biz"
)
```

如果有多方包命名冲突， 可以使用如下写法

```go
import (
    // 内部http包
    nethttp "net/http"
	
    // 第三方http包
    transporthttp "github.com/go-kratos/kratos/v2/transport/http"
)
```

### 2. 命名规范

* 统一按照Go官方命名规则（原则是Goland编辑器不能有警告⚠️）

* 对于不确定是否导出的变量， 统一使用小写字母开头，只有明确导出的，才使用大写字母开头

* 对于实现接口，对应的New方法应返回该接口，从而屏蔽实现细节，对应实现采用结构体采用小写字母开头方式

    ```go
    // InterfaceName 接口注释
    type InterfaceName interface {
        // MethodName 方法注释
        MethodName(param1, param2 string) (string, error)
    }
  
  
    type interfaceNameImpl struct {}
  
    // NewInterfaceName 创建接口实现
    func NewInterfaceName() InterfaceName {
        return &interfaceNameImpl{}
    }
    ```

* 对于出现频率超过`2次`以上的常量（`数字`、`字符串`）等，应统一定义为常量，可借鉴`vobj`定义方式


### 3. 注释规范

* 函数注释

```go
// FuncName 函数注释
func FuncName(param1, param2 string) (string, error) {
    return "", nil
}
```

多行的场景, 后面行缩进一字符， 第一行与后面空行间隔

```go
// FuncName 函数注释
// 
//  param1: 参数1注释
//  param2: 参数2注释
//  return1: 返回值1注释
//  return2: 返回值2注释
func FuncName(param1, param2 string) (string, error) {
    return "", nil
}
```

* 变量注释

```go
var (
	// varName 变量注释
	varName string
	// varName2 变量注释
	varName2 string
)
```

* 结构体注释

```go
// StructName 结构体注释
type StructName struct {
	// fieldName 字段注释
	fieldName string
	// fieldName2 字段注释
	fieldName2 string
}
```

* 接口注释

```go
// InterfaceName 接口注释
type InterfaceName interface {
	// MethodName 方法注释
	MethodName(param1, param2 string) (string, error)
}
```

* 方法注释

```go
// MethodName 方法注释
func (s *StructName) MethodName(param1, param2 string) (string, error) {
    return "", nil
}
```

* 常量注释

```go
// ConstName 常量注释
const (
  // ConstName 常量注释
  ConstName = "constName"
  // ConstName2 常量注释
  ConstName2 = "constName2"
  // ConstName3 常量注释
  ConstName3 = "constName3"
)

```

**特殊说明， 项目中vobj里面的常量， 注释是放在行尾，因为会通过stringer工具生成，对应的方法**

### 4. 函数定义规范

原则上函数或者方法，第一参数都是context（不涉及外部调用或者调用链中涉及上下文管理的工具类方法除外），如果有return error， error放在最后

```go
func FuncName(ctx context.Context, param1) (string, error) {
	return "", nil
}
```

### 5. 结构体方法

接收器名一般为结构体名的首字母小写，比如结构体为User，方法为GetName，接收器为u，需为单字符，不可以结构体名称各个单词缩写命名, 也不可使用this、self等

```go
type User struct {}

func (u *User) GetName() string {
	return ""
}
```

## 外部库引入规范

在开发某些功能时， 需要引入对应的外部实现来完成， 需要团队评估该外部依赖，同意后方可引入

对于简单功能的外部包，建议是内部自己实现, 减少外部依赖，减少维护成本
