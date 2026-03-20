package main

import (
	"encoding/json"
	"fmt"
	"time"

	lzstring "github.com/daku10/go-lz-string"
	"github.com/phrynus/go-utils/unknown"
)

// TestUnknown 测试 SmartUnmarshal 对所有 Go 类型的支持
func TestUnknown() {
	// 计时
	start := time.Now()
	// ==================== 基础类型转换测试 ====================
	fmt.Println("\n========== 基础类型转换测试 ==========")

	// Bool 转换
	fmt.Println("\n--- Bool 转换 ---")
	fmt.Printf("true -> Bool: %v\n", unknown.NewUnknown(true).Bool())
	fmt.Printf("\"true\" -> Bool: %v\n", unknown.NewUnknown("true").Bool())
	fmt.Printf("\"yes\" -> Bool: %v\n", unknown.NewUnknown("yes").Bool())
	fmt.Printf("\"1\" -> Bool: %v\n", unknown.NewUnknown("1").Bool())
	fmt.Printf("123 -> Bool: %v\n", unknown.NewUnknown(123).Bool())
	fmt.Printf("0 -> Bool: %v\n", unknown.NewUnknown(0).Bool())
	fmt.Printf("nil -> Bool: %v\n", unknown.NewUnknown(nil).Bool())

	// Int 转换
	fmt.Println("\n--- Int 转换 ---")
	fmt.Printf("\"123\" -> Int: %v\n", unknown.NewUnknown("123").Int())
	fmt.Printf("\"123.45\" -> Int: %v\n", unknown.NewUnknown("123.45").Int())
	fmt.Printf("123 -> Int: %v\n", unknown.NewUnknown(123).Int())
	fmt.Printf("int64(999) -> Int: %v\n", unknown.NewUnknown(int64(999)).Int())
	fmt.Printf("uint64(888) -> Int: %v\n", unknown.NewUnknown(uint64(888)).Int())
	fmt.Printf("float64(77.5) -> Int: %v\n", unknown.NewUnknown(float64(77.5)).Int())
	fmt.Printf("true -> Int: %v\n", unknown.NewUnknown(true).Int())
	fmt.Printf("\"hello\" -> Int: %v\n", unknown.NewUnknown("hello").Int())

	// Float 转换
	fmt.Println("\n--- Float 转换 ---")
	fmt.Printf("\"123.45\" -> Float: %v\n", unknown.NewUnknown("123.45").Float())
	fmt.Printf("123 -> Float: %v\n", unknown.NewUnknown(123).Float())
	fmt.Printf("json.Number(\"99.5\") -> Float: %v\n", unknown.NewUnknown(json.Number("99.5")).Float())

	// Uint 转换
	fmt.Println("\n--- Uint 转换 ---")
	fmt.Printf("\"100\" -> Uint: %v\n", unknown.NewUnknown("100").Uint())
	fmt.Printf("-10 -> Uint: %v (负数返回0)\n", unknown.NewUnknown(-10).Uint())
	fmt.Printf("uint(200) -> Uint: %v\n", unknown.NewUnknown(uint(200)).Uint())

	// ==================== Time 转换测试 ====================
	fmt.Println("\n========== Time 转换测试 ==========")

	now := time.Now()

	// time.Time 直接转换
	fmt.Printf("time.Now() -> Unix: %v\n", unknown.NewUnknown(now).Unix())
	fmt.Printf("time.Now() -> UnixMilli: %v\n", unknown.NewUnknown(now).UnixMilli())

	// 字符串解析为时间
	fmt.Printf("\"2024-01-15 10:30:00\" -> Time: %v\n", unknown.NewUnknown("2024-01-15 10:30:00").Time())
	fmt.Printf("\"2024-01-15T10:30:00Z\" -> Time: %v\n", unknown.NewUnknown("2024-01-15T10:30:00Z").Time())
	fmt.Printf("\"2024/01/15 10:30:00\" -> Time: %v\n", unknown.NewUnknown("2024/01/15 10:30:00").Time())
	fmt.Printf("\"20240115\" -> Time: %v\n", unknown.NewUnknown("20240115").Time())

	// 时间戳解析
	fmt.Printf("1705312200 (秒) -> Time: %v\n", unknown.NewUnknown(1705312200).Time())
	fmt.Printf("1705312200000 (毫秒) -> Time: %v\n", unknown.NewUnknown(1705312200000).Time())
	fmt.Printf("\"1705312200\" (字符串秒) -> Time: %v\n", unknown.NewUnknown("1705312200").Time())
	fmt.Printf("\"1705312200000\" (字符串毫秒) -> Time: %v\n", unknown.NewUnknown("1705312200000").Time())

	// JSON Date 格式
	fmt.Printf("\"/Date(1705312200000)/\" -> Time: %v\n", unknown.NewUnknown("/Date(1705312200000)/").Time())

	// ==================== Duration 转换测试 ====================
	fmt.Println("\n========== Duration 转换测试 ==========")
	fmt.Printf("\"1h30m\" -> Duration: %v\n", unknown.NewUnknown("1h30m").Duration())
	fmt.Printf("\"5000\" (毫秒) -> Duration: %v\n", unknown.NewUnknown("5000").Duration())
	fmt.Printf("5000 (毫秒) -> Duration: %v\n", unknown.NewUnknown(5000).Duration())
	fmt.Printf("time.Hour -> Duration: %v\n", unknown.NewUnknown(time.Hour).Duration())

	// ==================== Map 转换测试 ====================
	fmt.Println("\n========== Map 转换测试 ==========")

	mapData := map[string]interface{}{
		"name":   "张三",
		"age":    25,
		"score":  95.5,
		"active": true,
	}

	ut := unknown.NewUnknown(mapData)

	fmt.Printf("Map Keys: %v\n", ut.MapKeys())
	fmt.Printf("Map Values: %v\n", ut.MapValues())
	fmt.Printf("StringMap: %v\n", ut.StringMap())
	fmt.Printf("IntMap: %v\n", ut.IntMap())
	fmt.Printf("FloatMap: %v\n", ut.FloatMap())
	fmt.Printf("BoolMap: %v\n", ut.BoolMap())
	fmt.Printf("GetMapValue(\"name\"): %v\n", ut.GetMapValue("name"))
	fmt.Printf("GetMapValue(\"age\"): %v\n", ut.GetMapValue("age"))

	// 嵌套 Map 测试
	nestedMap := map[string]interface{}{
		"user": map[string]interface{}{
			"profile": map[string]interface{}{
				"name": "李四",
			},
		},
	}
	nestedUt := unknown.NewUnknown(nestedMap)
	fmt.Printf("GetMapValue(\"user.profile.name\"): %v\n", nestedUt.GetMapValue("user.profile.name"))

	// ==================== Array/Slice 转换测试 ====================
	fmt.Println("\n========== Array/Slice 转换测试 ==========")

	// 字符串转字符数组
	strArr := unknown.NewUnknown("hello")
	fmt.Printf("\"hello\" -> Array: %v (runes)\n", strArr.Array())
	fmt.Printf("\"hello\" -> Len: %v\n", strArr.Len())

	// 切片
	sliceData := []interface{}{1, "two", 3.0, true}
	sliceUt := unknown.NewUnknown(sliceData)
	fmt.Printf("[]interface{}{1, \"two\", 3.0, true} -> Array: %v\n", sliceUt.Array())
	fmt.Printf("[]interface{}{...} -> Len: %v\n", sliceUt.Len())

	// Map 的 Keys 作为数组
	mapArr := unknown.NewUnknown(map[string]int{"a": 1, "b": 2, "c": 3})
	fmt.Printf("map[string]int -> Array (keys): %v\n", mapArr.Array())

	// ==================== 类型检查测试 ====================
	fmt.Println("\n========== 类型检查测试 ==========")

	checkUt := unknown.NewUnknown(123)
	fmt.Printf("123 -> IsInt: %v, IsFloat: %v, IsString: %v, IsBool: %v, IsNumber: %v\n",
		checkUt.IsInt(), checkUt.IsFloat(), checkUt.IsString(), checkUt.IsBool(), checkUt.IsNumber())

	strUt := unknown.NewUnknown("hello")
	fmt.Printf("\"hello\" -> IsInt: %v, IsFloat: %v, IsString: %v\n",
		strUt.IsInt(), strUt.IsFloat(), strUt.IsString())

	boolUt := unknown.NewUnknown(true)
	fmt.Printf("true -> IsBool: %v, KindName: %v\n", boolUt.IsBool(), boolUt.KindName())

	timeUt := unknown.NewUnknown(now)
	fmt.Printf("time.Now() -> IsTime: %v\n", timeUt.IsTime())

	durationUt := unknown.NewUnknown(time.Hour)
	fmt.Printf("time.Hour -> IsDuration: %v\n", durationUt.IsDuration())

	nilUt := unknown.NewUnknown(nil)
	fmt.Printf("nil -> IsNil: %v, IsZero: %v, Kind: %v\n", nilUt.IsNil(), nilUt.IsZero(), nilUt.Kind())

	// ==================== Struct 相关测试 ====================
	fmt.Println("\n========== Struct 相关测试 ==========")

	type User struct {
		Name  string `json:"user_name"`
		Age   int    `json:"user_age"`
		Email string `json:"email"`
	}

	userData := map[string]interface{}{
		"user_name": "王五",
		"user_age":  30,
		"email":     "wang@example.com",
	}

	userUt := unknown.NewUnknown(userData)
	fmt.Printf("User Fields: %v\n", userUt.Fields())

	// StructField 测试
	type Person struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Age       int    `json:"age"`
	}

	person := Person{FirstName: "张", LastName: "三", Age: 28}
	personUt := unknown.NewUnknown(person)
	fmt.Printf("StructField(\"first_name\"): %v\n", personUt.StructField("first_name"))
	fmt.Printf("StructField(\"age\"): %v\n", personUt.StructField("age"))

	// ==================== SmartUnmarshal 测试 ====================
	fmt.Println("\n========== SmartUnmarshal 测试 ==========")

	type NestedStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	type AllTypesStruct struct {
		String2   *string        `json:"string2"`
		String3   *string        `json:"string3"`
		Bool      *bool          `json:"bool"`
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
		"string2": "",
		"string3": "string3_value",
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
	err := unknown.NewUnknown(jsonStr).SmartUnmarshal(&result)
	if err != nil {
		fmt.Printf("SmartUnmarshal error: %v\n", err)
		return
	}

	fmt.Printf("String3: %v\n", *result.String3)
	fmt.Printf("Bool: %v\n", *result.Bool)
	fmt.Printf("Int: %v\n", result.Int)
	fmt.Printf("Int8: %v, Int16: %v, Int32: %v, Int64: %v\n", result.Int8, result.Int16, result.Int32, result.Int64)
	fmt.Printf("Uint: %v, Uint8: %v, Uint16: %v, Uint32: %v, Uint64: %v\n",
		result.Uint, result.Uint8, result.Uint16, result.Uint32, result.Uint64)
	fmt.Printf("Float32: %v, Float64: %v\n", result.Float32, result.Float64)
	fmt.Printf("Array: %v\n", result.Array)
	fmt.Printf("Slice: %v\n", result.Slice)
	fmt.Printf("Map: %v\n", result.Map)
	fmt.Printf("Struct: %+v\n", result.Struct)
	fmt.Printf("Interface: %v\n", result.Interface)

	// ==================== JSON 序列化测试 ====================
	fmt.Println("\n========== JSON 序列化测试 ==========")

	fmt.Printf("result.JsonString():\n%v\n", unknown.NewUnknown(result).JsonString())

	fmt.Printf("\nresult.JsonIndentString():\n%v\n", unknown.NewUnknown(result).JsonIndentString())

	// ==================== 指针解引用测试 ====================
	fmt.Println("\n========== 指针解引用测试 ==========")

	ptrValue := 42
	ptr := &ptrValue
	fmt.Printf("*ptr (42) -> Deref: %v\n", unknown.NewUnknown(ptr).Deref())

	doublePtr := &ptr
	fmt.Printf("**ptr -> Deref: %v\n", unknown.NewUnknown(doublePtr).Deref())

	// ==================== String 转换测试 ====================
	fmt.Println("\n========== String 转换测试 ==========")

	fmt.Printf("true -> String: %v\n", unknown.NewUnknown(true).String())
	fmt.Printf("123 -> String: %v\n", unknown.NewUnknown(123).String())
	fmt.Printf("3.14 -> String: %v\n", unknown.NewUnknown(3.14).String())
	fmt.Printf("(1+2i) -> String: %v\n", unknown.NewUnknown(complex(1, 2)).String())
	fmt.Printf("time.Now() -> String: %v\n", unknown.NewUnknown(now).String())
	fmt.Printf("time.Hour -> String: %v\n", unknown.NewUnknown(time.Hour).String())

	// ==================== LZString 压缩测试 ====================
	fmt.Println("\n========== LZString 压缩测试 ==========")

	jsonBytes := unknown.NewUnknown(result).JsonString()
	compressed, _ := lzstring.CompressToUint8Array(jsonBytes)

	fmt.Printf("原始大小: %d 字节\n", len(jsonBytes))
	fmt.Printf("压缩后: %d 字节\n", len(compressed))
	fmt.Printf("压缩比: %.2f%%\n", float64(len(compressed))/float64(len(jsonBytes))*100)

	fmt.Println("\n========== 所有测试完成 ==========")

	elapsed := time.Since(start)
	fmt.Printf("测试完成，耗时: %s\n", elapsed)
}
