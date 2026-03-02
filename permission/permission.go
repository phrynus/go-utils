package permission

import "strings"

// Checker 权限检查器
type Checker struct{}

// Has 检查用户是否有指定权限
func (c *Checker) Has(userPerms []string, required string) bool {
	requiredLen := len(required)

	// 直接遍历用户权限，使用高效的字符串比较
	for _, userPerm := range userPerms {
		userLen := len(userPerm)

		// 1. 精确匹配 - 直接字符串比较
		if userLen == requiredLen && userPerm == required {
			return true
		}

		// 2. 上级扩展：检查用户权限是否是上级（父权限）
		// 例如：userPerm="eyes:admin:user", required="eyes:admin:user:list"
		if userLen < requiredLen && strings.HasPrefix(required, userPerm) &&
			requiredLen > userLen && required[userLen] == ':' {
			return true
		}

		// 3. 防线机制：检查用户权限是否是下级（子孙权限）
		// 例如：userPerm="eyes:admin:role:list", required="eyes:admin:role"
		if userLen > requiredLen && strings.HasPrefix(userPerm, required) &&
			userLen > requiredLen && userPerm[requiredLen] == ':' {
			return true
		}
	}

	return false
}

// Check 批量检查权限
func (c *Checker) Check(userPerms []string, requiredPerms []string) []bool {
	results := make([]bool, len(requiredPerms))

	for i, perm := range requiredPerms {
		results[i] = c.Has(userPerms, perm)
	}

	return results
}
