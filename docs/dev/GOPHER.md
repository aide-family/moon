<div style="display: flex; align-items: center;">
  <img 
    src="../images/logo.svg" 
    alt="Logo" 
    style="height: 4em; width: auto; vertical-align: middle; margin-right: 10px;" 
  />
  <h1 style="margin: 0; font-size: 24px; line-height: 1.5;">Moon - Multi-Domain Monitoring and Alerting Platform</h1>
</div>

| [English](GOPHER.md) | [中文简体](GOPHER.zh-CN.md) |

# Must-Read for Developers

## Development Guidelines

### 1. Package Import Standards

When importing packages, whether it's a single external package or multiple, use the following format:

```go
import (
  "fmt"
)
```

When importing multiple packages, follow this order, separated by blank lines:

```go
import (
  // Blank identifier imports
  _ "gorm.io/driver/sqlite"
  
  // Standard library
  "fmt"
  
  // Third-party packages
  "github.com/xxxx/xxxx"
  
  // Internal project packages
  "github.com/aide-family/moon/internal/biz"
)
```

If there are naming conflicts between packages, use the following format:

```go
import (
  // Internal http package
  nethttp "net/http"
  
  // Third-party http package
  transporthttp "github.com/go-kratos/kratos/v2/transport/http"
)
```

### 2. Naming Conventions

* Follow the official Go naming rules (the principle is that there should be no warnings in the Goland editor).

* For variables where it is uncertain whether they should be exported, use lowercase letters. Only explicitly exported
  variables should start with an uppercase letter.

* For interface implementations, the corresponding New method should return the interface to hide implementation
  details. The implementation struct should start with a lowercase letter.

   ```go
   // InterfaceName Interface comment
   type InterfaceName interface {
        // MethodName Method comment
        MethodName(param1, param2 string) (string, error)
   }
   
   type interfaceNameImpl struct {}
   
   // NewInterfaceName Creates an interface implementation
   func NewInterfaceName() InterfaceName {
        return &interfaceNameImpl{}
   }
   ```

* For constants (numbers, strings, etc.) that appear more than twice, define them as constants. Refer to the vobj
  definition style.

### 3. Commenting Standards

* Function comments:

```go
// FuncName Function comment
func FuncName(param1, param2 string) (string, error) {
  return "", nil
}
```

* For multi-line comments, indent subsequent lines by one character, and leave a blank line after the first line:

```go
// FuncName Function comment
// 
//  param1: Parameter 1 comment
//  param2: Parameter 2 comment
//  return1: Return value 1 comment
//  return2: Return value 2 comment
func FuncName(param1, param2 string) (string, error) {
  return "", nil
}
```

* Variable comments:

```go
var (
  // varName Variable comment
  varName string
  // varName2 Variable comment
  varName2 string
)
```

* Struct comments:

```go
// StructName Struct comment
type StructName struct {
  // fieldName Field comment
  fieldName string
  // fieldName2 Field comment
  fieldName2 string
}
```

* Interface comments:

```go
// InterfaceName Interface comment
type InterfaceName interface {
  // MethodName Method comment
  MethodName(param1, param2 string) (string, error)
}
```

* Method comments:

```go
// MethodName Method comment
func (s *StructName) MethodName(param1, param2 string) (string, error) {
    return "", nil
}
```

* Constant comments:

```go
// ConstName Constant comment
const (
  // ConstName Constant comment
  ConstName = "constName"
  // ConstName2 Constant comment
  ConstName2 = "constName2"
  // ConstName3 Constant comment
  ConstName3 = "constName3"
)

```

**Special note: For constants in the vobj package, comments are placed at the end of the line because they are generated
using the stringer tool.**

### 4. Function Definition Standards

In principle, the first parameter of a function or method should be context (except for utility methods). If there is a
return error, the error should be the last return value.

```go
func FuncName(ctx context.Context, param1) (string, error) {
    return "", nil
}
```

### 5. Struct Methods

The receiver name should generally be the first letter of the struct name in lowercase. For example, if the struct is
User and the method is GetName, the receiver should be u. It should be a single character and not an abbreviation of the
struct name or words like this or self.

```go
type User struct {}

func (u *User) GetName() string {
  return ""
}
```

### External Library Import Standards

When developing certain features, if external implementations are required, the team must evaluate and approve the
external dependency before it can be introduced.

For simple functionalities, it is recommended to implement them internally to reduce external dependencies and
maintenance costs.
