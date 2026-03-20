# Unknown 工具库

`unknown.go` 提供了一个强大的 `UnknownType` 类型封装，支持任意类型到目标类型的智能转换，是处理动态数据和 JSON 解析的利器。

## 功能特性

- ✅ 20+ 种类型转换（Bool、Int/Uint/Float 系列、Complex、Duration、Time、String、Bytes、Array、Map、Struct、指针、通道、函数）
- ✅ 智能类型推断（自动识别 string/int/float/bool/json.Number 等多种输入类型）
- ✅ 零值安全（所有转换方法均提供 Must 系列，支持默认值参数，nil 值不 panic）
- ✅ 丰富的时间解析（RFC3339、RFC1123、紧凑格式、时间戳、JSON Date 等 30+ 种格式）
- ✅ 嵌套 Map 访问（支持 `user.profile.name` 形式的点号路径）
- ✅ SmartUnmarshal（JSON 字符串或 Map 自动填充到任意目标结构体，支持多种 key 匹配策略）

## 安装

```bash
go get github.com/phrynus/go-utils
```

## 快速开始

```go
import "github.com/phrynus/go-utils"

ut := utils.NewUnknown("123")
fmt.Println(ut.Int()) // 输出: 123

ut2 := utils.NewUnknown(true)
fmt.Println(ut2.Int()) // 输出: 1
```

## 核心类型

### UnknownType

通用类型封装结构体：

```go
type UnknownType struct {
    Value any       // 实际存储的值
    Type  reflect.Type  // 值的反射类型
}
```

## 类型转换

### Bool 转换

```go
ut := utils.NewUnknown("true")
fmt.Println(ut.Bool()) // true

ut = utils.NewUnknown(123)
fmt.Println(ut.Bool()) // true (非0为true)

ut = utils.NewUnknown("yes")
fmt.Println(ut.Bool()) // true

ut = utils.NewUnknown(nil)
fmt.Println(ut.Bool()) // false
```

**支持的输入类型**：
- `bool`: 直接返回
- `string`: "true", "1", "yes", "on" 返回 true
- `int/uint/float`: 非0返回true
- `json.Number`: 解析数字
- 指针/接口: 非nil返回true

### Int 系列

```go
ut := utils.NewUnknown("123")
fmt.Println(ut.Int())     // 123 (int64)
fmt.Println(ut.Int8())    // 123 (int8)
fmt.Println(ut.Int16())   // 123 (int16)
fmt.Println(ut.Int32())   // 123 (int32)

ut = utils.NewUnknown("123.45")
fmt.Println(ut.Int())     // 123 (截断小数)

ut = utils.NewUnknown(true)
fmt.Println(ut.Int())     // 1

ut = utils.NewUnknown(time.Now())
fmt.Println(ut.Int())    // Unix时间戳(秒)

ut = utils.NewUnknown(time.Hour)
fmt.Println(ut.Int())     // 3600000 (毫秒)
```

### Uint 系列

```go
ut := utils.NewUnknown(-10)
fmt.Println(ut.Uint())    // 0 (负数返回0)

ut = utils.NewUnknown("100")
fmt.Println(ut.Uint())    // 100
fmt.Println(ut.Uint8())   // 100
fmt.Println(ut.Uint16())  // 100
fmt.Println(ut.Uint32())  // 100
```

### Float 系列

```go
ut := utils.NewUnknown("123.45")
fmt.Println(ut.Float())   // 123.45 (float64)
fmt.Println(ut.Float32()) // 123.45 (float32)

ut = utils.NewUnknown(123)
fmt.Println(ut.Float())   // 123.0

ut = utils.NewUnknown(time.Now())
fmt.Println(ut.Float())    // Unix时间戳(纳秒)
```

### Complex 系列

```go
ut := utils.NewUnknown("1+2i")
fmt.Println(ut.Complex()) // (1+2i)

ut = utils.NewUnknown(5)
fmt.Println(ut.Complex()) // (5+0i)
```

### Duration 转换

```go
ut := utils.NewUnknown("1h30m")
fmt.Println(ut.Duration()) // 1h30m0s

ut = utils.NewUnknown("5000") // 毫秒
fmt.Println(ut.Duration())        // 5s

ut = utils.NewUnknown(time.Hour)
fmt.Println(ut.Duration())        // 1h0m0s
```

### Time 转换

支持丰富的时间格式解析：

```go
ut := utils.NewUnknown("2024-01-15 10:30:00")
fmt.Println(ut.Time()) // 2024-01-15 10:30:00 +0000 UTC

// RFC3339 格式
ut = utils.NewUnknown("2024-01-15T10:30:00Z")
fmt.Println(ut.Time())

// 斜杠分隔
ut = utils.NewUnknown("2024/01/15 10:30:00")
fmt.Println(ut.Time())

// 紧凑格式
ut = utils.NewUnknown("20240115")
fmt.Println(ut.Time())

ut = utils.NewUnknown("20240115103000")
fmt.Println(ut.Time())

// 时间戳 (秒级)
ut = utils.NewUnknown(1705312200)
fmt.Println(ut.Time())

// 时间戳 (毫秒级)
ut = utils.NewUnknown(1705312200000)
fmt.Println(ut.Time())

// JSON Date 格式
ut = utils.NewUnknown("/Date(1705312200000)/")
fmt.Println(ut.Time())

// Unix 时间戳便捷方法
fmt.Println(ut.Unix())      // 秒
fmt.Println(ut.UnixMilli()) // 毫秒
fmt.Println(ut.UnixNano())  // 纳秒
```

**支持的日期格式**：
- RFC3339 / RFC3339Nano
- RFC1123 / RFC1123Z / RFC822Z / RFC822
- ANSIC / UnixDate / RubyDate
- `2006-01-02T15:04:05[.999][Z07:00]`
- `2006-01-02 15:04:05[.999]`
- `2006/01/02 [15:04:05]`
- `2006-01-02`
- `01/02/2006 [15:04:05]` (美式/英式)
- `20060102150405` / `20060102` (紧凑格式)
- `/Date(毫秒)/` (JSON Date)
- 秒级/毫秒级时间戳

## Map 操作

### 基础 Map 转换

```go
data := map[string]interface{}{
    "name":   "张三",
    "age":    25,
    "score":  95.5,
    "active": true,
}

ut := utils.NewUnknown(data)

// 转换为 map[string]interface{}
fmt.Println(ut.Map())

// 转换为 map[string]string
fmt.Println(ut.StringMap())

// 转换为 map[string]int64
fmt.Println(ut.IntMap())

// 转换为 map[string]float64
fmt.Println(ut.FloatMap())

// 转换为 map[string]bool
fmt.Println(ut.BoolMap())
```

### Map 键值操作

```go
ut := utils.NewUnknown(data)

// 获取所有键
fmt.Println(ut.MapKeys()) // [name age score active]

// 获取所有值
fmt.Println(ut.MapValues()) // [张三 25 95.5 true]

// 获取指定键的值
fmt.Println(ut.GetMapValue("name")) // 张三
```

### 嵌套 Map 访问

```go
nested := map[string]interface{}{
    "user": map[string]interface{}{
        "profile": map[string]interface{}{
            "name": "李四",
        },
    },
}

ut := utils.NewUnknown(nested)
fmt.Println(ut.GetMapValue("user.profile.name")) // 李四
```

## Array/Slice 操作

### 转换为数组

```go
// 字符串转字符数组 (runes)
ut := utils.NewUnknown("hello")
fmt.Println(ut.Array()) // [104 101 108 108 111]

// 获取长度
fmt.Println(ut.Len()) // 5

// 切片
slice := []interface{}{1, "two", 3.0, true}
ut = utils.NewUnknown(slice)
fmt.Println(ut.Array()) // [1 two 3 true]

// Map 的键作为数组
m := map[string]int{"a": 1, "b": 2}
ut = utils.NewUnknown(m)
fmt.Println(ut.Array()) // [a b]
```

## 指针操作

### 指针解引用

```go
val := 42
ptr := &val
doublePtr := &ptr

ut := utils.NewUnknown(doublePtr)
fmt.Println(ut.Deref()) // 42 (递归解引用)

ut = utils.NewUnknown(ptr)
fmt.Println(ut.Deref()) // 42

// Pointer 方法 (非递归)
fmt.Println(ut.Pointer()) // *ptr 的值
```

## String 操作

### 转换为字符串

```go
ut := utils.NewUnknown(true)
fmt.Println(ut.String()) // true

ut = utils.NewUnknown(123)
fmt.Println(ut.String()) // 123

ut = utils.NewUnknown(3.14)
fmt.Println(ut.String()) // 3.14

ut = utils.NewUnknown(complex(1, 2))
fmt.Println(ut.String()) // (1+2i)

ut = utils.NewUnknown(time.Now())
fmt.Println(ut.String()) // 2024-01-15T10:30:00+08:00

ut = utils.NewUnknown(time.Hour)
fmt.Println(ut.String()) // 1h0m0s

// 带引号
fmt.Println(ut.Quote()) // "value"
```

## JSON 操作

### 序列化

```go
data := map[string]interface{}{
    "name": "张三",
    "age":  25,
}

ut := utils.NewUnknown(data)

// JSON 字节
fmt.Println(ut.Json()) // {"age":25,"name":"张三"}

// JSON 字符串
fmt.Println(ut.JsonString())

// 格式化输出
fmt.Println(ut.JsonIndentString())
// 输出:
// {
//   "age": 25,
//   "name": "张三"
// }
```

### 反序列化 (SmartUnmarshal)

智能地将数据填充到目标结构体：

```go
type User struct {
    Name    string `json:"user_name"`
    Age     int    `json:"user_age"`
    Email   string `json:"email"`
}

jsonStr := `{"user_name":"王五","user_age":30,"email":"wang@example.com"}`

var user User
err := utils.NewUnknwonType(jsonStr).SmartUnmarshal(&user)
if err != nil {
    fmt.Println("解析错误:", err)
}
fmt.Printf("%+v\n", user)
// {Name:王五 Age:30 Email:wang@example.com}
```

**SmartUnmarshal 特性**：
- 自动解析 JSON 字符串
- 支持 map 转结构体
- 支持多种 key 匹配：原始名、小写名、snake_case
- 支持嵌套结构体
- 支持所有 Go 基础类型

### 完整类型转换示例

```go
type NestedStruct struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

type AllTypesStruct struct {
    String    *string        `json:"string"`
    Bool      bool           `json:"bool"`
    Int       int            `json:"int"`
    Int8      int8           `json:"int8"`
    Int16     int16          `json:"int16"`
    Int32     int32          `json:"int32"`
    Int64     int64          `json:"int64"`
    Uint      uint           `json:"uint"`
    Uint8     uint8          `json:"uint8"`
    Uint16    uint16         `json:"uint16"`
    Uint32    uint32         `json:"uint32"`
    Uint64    uint64         `json:"uint64"`
    Float32   float32        `json:"float32"`
    Float64   float64        `json:"float64"`
    Array     [3]int         `json:"array"`
    Slice     []int          `json:"slice"`
    Map       map[string]int `json:"map"`
    Struct    NestedStruct   `json:"struct"`
    Interface interface{}    `json:"interface"`
}

jsonStr := `{
    "string": "hello",
    "bool": true,
    "int": 123,
    "int8": 8,
    "int16": 16,
    "int32": 32,
    "int64": 64,
    "uint": 100,
    "uint8": 8,
    "uint16": 16,
    "uint32": 32,
    "uint64": 64,
    "float32": 3.54,
    "float64": 6.28,
    "array": [1, 2, 3],
    "slice": [4, 5, 6],
    "map": {"a": 1, "b": 2},
    "struct": {"name": "test", "age": 25},
    "interface": "interface_value"
}`

var result AllTypesStruct
err := utils.NewUnknown(jsonStr).SmartUnmarshal(&result)
```

## 类型检查

```go
ut := utils.NewUnknown(123)

fmt.Println(ut.IsInt())      // true
fmt.Println(ut.IsFloat())    // false
fmt.Println(ut.IsString())   // false
fmt.Println(ut.IsBool())     // false
fmt.Println(ut.IsNumber())   // true
fmt.Println(ut.IsSlice())    // false
fmt.Println(ut.IsMap())       // false
fmt.Println(ut.IsStruct())   // false
fmt.Println(ut.IsPtr())      // false
fmt.Println(ut.IsTime())      // false
fmt.Println(ut.IsDuration())  // false

// 类型信息
fmt.Println(ut.Kind())      // int
fmt.Println(ut.TypeName())  // int
fmt.Println(ut.KindName())  // int

// nil 和零值检查
var nilUt utils.UnknownType
fmt.Println(nilUt.IsNil())  // true
fmt.Println(nilUt.IsZero())  // true
```

## 辅助方法

### Bytes 转换

```go
ut := utils.NewUnknown("hello")
fmt.Println(ut.Bytes()) // [104 101 108 108 111]

ut = utils.NewUnknown([]byte("world"))
fmt.Println(ut.Bytes()) // [119 111 114 108 100]
```

### Struct 操作

```go
type Person struct {
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
    Age       int    `json:"age"`
}

person := Person{FirstName: "张", LastName: "三", Age: 28}
ut := utils.NewUnknown(person)

// 获取所有字段
fmt.Println(ut.Fields())
// map[age:28 first_name:张 last_name:三]

// 获取指定字段
fmt.Println(ut.StructField("first_name")) // 张
```

### 零值安全方法

```go
ut := utils.NewUnknown(nil)

// 带默认值的方法
fmt.Println(ut.MustBool(true))          // true
fmt.Println(ut.MustInt(100))           // 100
fmt.Println(ut.MustFloat(3.14))        // 3.14
fmt.Println(ut.MustString("default"))  // default
fmt.Println(ut.MustTime(time.Now()))   // 当前时间
```

## 错误处理

```go
var (
    ErrNotConvertible = errors.New("值无法转换为目标类型")
    ErrNilValue       = errors.New("值为 nil")
)

// SmartUnmarshal 返回错误
err := utils.NewUnknown("invalid").SmartUnmarshal(&target)
if err != nil {
    fmt.Println(err) // 目标必须是非空指针: 值无法转换为目标类型
}
```

## 方法列表

| 方法 | 返回类型 | 说明 |
|------|---------|------|
| `Bool()` | `bool` | 转换为布尔值 |
| `Int()` | `int64` | 转换为整数 |
| `Int8()` | `int8` | 转换为 int8 |
| `Int16()` | `int16` | 转换为 int16 |
| `Int32()` | `int32` | 转换为 int32 |
| `Uint()` | `uint64` | 转换为无符号整数 |
| `Uint8()` | `uint8` | 转换为 uint8 |
| `Uint16()` | `uint16` | 转换为 uint16 |
| `Uint32()` | `uint32` | 转换为 uint32 |
| `Float()` | `float64` | 转换为浮点数 |
| `Float32()` | `float32` | 转换为 float32 |
| `Complex()` | `complex128` | 转换为复数 |
| `Duration()` | `time.Duration` | 转换为 Duration |
| `Time()` | `time.Time` | 转换为时间 |
| `Unix()` | `int64` | 获取 Unix 时间戳(秒) |
| `UnixMilli()` | `int64` | 获取 Unix 时间戳(毫秒) |
| `UnixNano()` | `int64` | 获取 Unix 时间戳(纳秒) |
| `Bytes()` | `[]byte` | 转换为字节切片 |
| `Array()` / `Slice()` | `[]interface{}` | 转换为切片 |
| `Len()` | `int` | 获取长度 |
| `Map()` | `map[string]interface{}` | 转换为 Map |
| `StringMap()` | `map[string]string` | 转换为字符串 Map |
| `IntMap()` | `map[string]int64` | 转换为整数 Map |
| `FloatMap()` | `map[string]float64` | 转换为浮点 Map |
| `BoolMap()` | `map[string]bool` | 转换为布尔 Map |
| `MapKeys()` | `[]interface{}` | 获取所有键 |
| `MapValues()` | `[]interface{}` | 获取所有值 |
| `GetMapValue(key)` | `interface{}` | 获取指定键的值 |
| `Pointer()` | `interface{}` | 提取指针值 |
| `Deref()` | `interface{}` | 递归解引用 |
| `Chan()` | `interface{}` | 转换为通道 |
| `Func()` | `interface{}` | 转换为函数 |
| `Struct()` | `interface{}` | 转换为结构体 |
| `StructField(name)` | `interface{}` | 获取结构体字段 |
| `Fields()` | `map[string]interface{}` | 获取所有字段 |
| `String()` | `string` | 转换为字符串 |
| `Quote()` | `string` | 带引号的字符串 |
| `Json()` | `[]byte` | JSON 序列化 |
| `JsonString()` | `string` | JSON 字符串 |
| `JsonIndent()` | `[]byte` | 格式化 JSON |
| `JsonIndentString()` | `string` | 格式化 JSON 字符串 |
| `Kind()` | `reflect.Kind` | 获取值的 Kind |
| `IsNil()` | `bool` | 是否为 nil |
| `IsZero()` | `bool` | 是否为零值 |
| `IsBool()` | `bool` | 是否为布尔 |
| `IsInt()` | `bool` | 是否为整数 |
| `IsFloat()` | `bool` | 是否为浮点 |
| `IsNumber()` | `bool` | 是否为数字 |
| `IsString()` | `bool` | 是否为字符串 |
| `IsSlice()` | `bool` | 是否为切片 |
| `IsMap()` | `bool` | 是否为 Map |
| `IsStruct()` | `bool` | 是否为结构体 |
| `IsPtr()` | `bool` | 是否为指针 |
| `IsArray()` | `bool` | 是否为数组 |
| `IsTime()` | `bool` | 是否为时间 |
| `IsDuration()` | `bool` | 是否为 Duration |
| `TypeName()` | `string` | 获取类型名 |
| `KindName()` | `string` | 获取 Kind 名 |
| `SmartUnmarshal(v)` | `error` | 智能反序列化 |

## Must 系列方法

以下方法支持可选默认值参数：

```go
ut := utils.NewUnknown(nil)

ut.MustBool(true)
ut.MustInt(0)
ut.MustFloat(0.0)
ut.MustString("")
ut.MustTime(time.Time{})
ut.MustBytes(nil)
ut.MustArray(nil)
ut.MustMap(nil)
ut.MustChan(nil)
ut.MustFunc(nil)
ut.MustComplex(0)
ut.MustDuration(0)
ut.MustJson([]byte("null"))
```

## 全局变量

### TimeFormats

可导出的时间格式切片，用于自定义支持的时间格式：

```go
// 添加自定义格式
utils.TimeFormats = append(utils.TimeFormats, "2006年01月02日")
```
