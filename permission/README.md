# Permission 权限检查包

这是一个高效的权限检查器，支持层级权限管理和批量权限验证。

## 特性

- **内存高效**: 零额外内存分配
- **层级权限**: 支持父子权限关系检查
- **批量处理**: 支持一次性检查多个权限
- **防线机制**: 子权限可以验证父权限
- **性能监控**: 内置时间比对测试

## 使用方法

### 基本使用

```go
import "github.com/phrynus/go-utils/permission"

// 创建权限检查器
checker := &permission.Checker{}

// 用户拥有的权限列表
userPerms := []string{
    "eyes:admin:user",
    "eyes:admin:user:list",
    "eyes:admin:role",
}

// 检查单个权限
hasPermission := checker.Has(userPerms, "eyes:admin:user:list")
fmt.Println(hasPermission) // true
```

### 权限层级规则

1. **精确匹配**: 权限字符串完全相同
2. **上级扩展**: 父权限可以访问子权限
   - 用户权限: `"eyes:admin:user"`
   - 所需权限: `"eyes:admin:user:list"` → ✅ 允许
3. **防线机制**: 子权限可以验证父权限
   - 用户权限: `"eyes:admin:user:list"`
   - 所需权限: `"eyes:admin:user"` → ✅ 允许

### 批量检查

```go
// 需要检查的权限列表
requiredPerms := []string{
    "eyes:admin:user",
    "eyes:admin:user:list",
    "eyes:admin:role:edit",
}

// 批量检查所有权限
results := checker.Check(userPerms, requiredPerms)
// results: [true, true, false]
```

## 性能特点

`Checker` 在小到中等规模的权限检查中表现出色：

- **精确匹配**: O(n) 时间复杂度，n为用户权限数量
- **层级检查**: 自动处理父子权限关系
- **内存友好**: 零额外内存分配
- **适用场景**: 权限集不大（<1000个），检查频率中等

对于大型权限系统，可以考虑使用专门的权限管理中间件。

## 运行测试

```bash
cd example
go run .
```

测试会输出详细的权限检查结果，包括单权限检查、批量检查和性能测试。
