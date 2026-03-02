package main

import (
	"fmt"
	"time"

	"github.com/phrynus/go-utils/permission"
)

func TestPermission() {
	fmt.Println("======= 权限检查器测试 =======")

	// 创建权限检查器实例
	checker := &permission.Checker{}

	// 使用checker进行原有测试
	testChecker(checker)

	// 性能比对测试
	benchmarkChecker()
}

func testChecker(checker *permission.Checker) {
	// 测试用例数据
	userPerms := []string{
		"eyes:admin:user",        // 基本权限
		"eyes:admin:user:list",   // 子权限
		"eyes:admin:user:create", // 子权限
		"eyes:admin:role",        // 平级权限
		"eyes:admin:role:list",   // 子权限
		"eyes:admin:role:delete", // 子权限
		"system:config",          // 不同模块权限
		"system:config:read",     // 子权限
		"api:public",             // 公共权限
	}

	fmt.Println("\n=== 单权限检查测试 ===")

	// 测试用例
	testCases := []struct {
		required string
		expected bool
		desc     string
	}{
		// 精确匹配测试
		{"eyes:admin:user", true, "精确匹配用户权限"},
		{"eyes:admin:role", true, "精确匹配角色权限"},
		{"system:config", true, "精确匹配系统配置权限"},
		{"api:public", true, "精确匹配公共API权限"},
		{"nonexistent:perm", false, "不存在的权限"},

		// 上级扩展测试（父权限检查子权限）
		{"eyes:admin:user:list", true, "上级扩展：父权限检查子权限"},
		{"eyes:admin:user:create", true, "上级扩展：父权限检查子权限"},
		{"eyes:admin:role:list", true, "上级扩展：父权限检查子权限"},
		{"eyes:admin:role:delete", true, "上级扩展：父权限检查子权限"},
		{"system:config:read", true, "上级扩展：父权限检查子权限"},
		{"eyes:admin:user:edit", true, "上级扩展：父权限可以访问子权限"},

		// 防线机制测试（子权限检查父权限）
		{"eyes:admin:user", true, "防线机制：子权限检查父权限"},
		{"eyes:admin:role", true, "防线机制：子权限检查父权限"},
		{"system:config", true, "防线机制：子权限检查父权限"},
		{"eyes:admin", true, "防线机制：多层子权限检查祖先权限"},
		{"system", true, "防线机制：多层子权限检查祖先权限"},
		{"api", true, "防线机制：子权限检查父权限"},
	}

	for i, tc := range testCases {
		result := checker.Has(userPerms, tc.required)
		status := "✓"
		if result != tc.expected {
			status = "✗"
		}
		fmt.Printf("测试 %d: %s | 要求: %s | 结果: %v (期望: %v) | %s\n",
			i+1, status, tc.required, result, tc.expected, tc.desc)
	}

	fmt.Println("\n=== 批量权限检查测试 ===")

	// 批量检查测试
	requiredPerms := []string{
		"eyes:admin:user",
		"eyes:admin:user:list",
		"eyes:admin:role:edit", // 不存在的权限
		"system:config",
		"system:config:read",
		"api:public",
		"nonexistent:module:action",
	}

	results := checker.Check(userPerms, requiredPerms)

	fmt.Println("批量检查结果:")
	for i, reqPerm := range requiredPerms {
		fmt.Printf("  %s -> %v\n", reqPerm, results[i])
	}

	fmt.Println("\n=== 性能测试（内存效率验证）===")

	// 模拟大量权限检查来验证内存效率
	largeUserPerms := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		largeUserPerms[i] = fmt.Sprintf("module%d:action%d:subaction%d", i%10, i%100, i%1000)
	}

	largeRequiredPerms := make([]string, 100)
	for i := 0; i < 100; i++ {
		largeRequiredPerms[i] = fmt.Sprintf("module%d:action%d", i%10, i%100)
	}

	fmt.Printf("测试 %d 个用户权限 vs %d 个所需权限\n",
		len(largeUserPerms), len(largeRequiredPerms))

	// 执行批量检查
	batchResults := checker.Check(largeUserPerms, largeRequiredPerms)
	fmt.Printf("批量检查完成，结果数量: %d\n", len(batchResults))

	fmt.Println("\n======= 权限检查器测试完成 =======")
}

// 性能比对测试：权限检查器性能测试
func benchmarkChecker() {
	fmt.Println("\n======= 权限检查器性能测试 =======")

	// 准备测试数据 - 使用更真实的权限数据
	userPerms := []string{
		"user:read", "user:write", "user:delete",
		"admin:user:read", "admin:user:write", "admin:user:delete",
		"admin:role:read", "admin:role:write",
		"system:config:read", "system:config:write",
		"api:public:read", "api:private:read", "api:private:write",
		"file:upload", "file:download", "file:delete",
		"report:view", "report:create", "report:export",
	}

	// 混合的测试权限：有些存在，有些不存在，有些需要层级检查
	requiredPerms := []string{
		"user:read",        // 精确匹配 ✓
		"user:write",       // 精确匹配 ✓
		"admin:user:read",  // 精确匹配 ✓
		"admin:user:edit",  // 父权限可访问 ✓
		"system:config",    // 子权限验证父权限 ✓
		"api:public",       // 子权限验证父权限 ✓
		"file",             // 多层子权限验证祖先 ✓
		"report",           // 多层子权限验证祖先 ✓
		"nonexistent:perm", // 不存在 ✗
		"unknown:action",   // 不存在 ✗
		"admin:delete",     // 不存在 ✗
		"guest:login",      // 不存在 ✗
	}

	// 创建权限检查器
	checker := &permission.Checker{}

	fmt.Printf("测试配置：%d 个用户权限，%d 个所需权限\n", len(userPerms), len(requiredPerms))
	fmt.Println("用户权限:", userPerms)
	fmt.Println("所需权限:", requiredPerms)
	fmt.Println()

	// 测试权限检查器
	fmt.Println("🔍 测试 Checker...")
	start := time.Now()
	results := checker.Check(userPerms, requiredPerms)
	checkTime := time.Since(start)

	fmt.Printf("   执行时间: %v\n", checkTime)
	fmt.Printf("   结果数量: %d\n", len(results))

	// 显示详细结果
	fmt.Println("\n📋 权限检查结果:")
	for i, perm := range requiredPerms {
		result := results[i]
		status := "✓"
		if !result {
			status = "✗"
		}
		fmt.Printf("   %s %s -> %v\n", status, perm, result)
	}

	// 详细的时间分解测试 - 多次重复测试获得更准确的数据
	fmt.Println("\n🔬 详细时间分解测试 (重复执行以获得准确数据):")

	iterations := 10000 // 增加迭代次数

	// 单权限检查 - 测试不同类型的权限
	testCases := []string{
		"user:read",             // 存在权限 - 精确匹配
		"admin:user:edit",       // 不存在权限 - 需要层级检查
		"api:private",           // 存在权限 - 上级匹配
		"api:private:user:edit", // 不存在权限
		"nonexistent:perm",      // 不存在权限 - 完全不匹配
	}

	for _, testPerm := range testCases {
		fmt.Printf("\n   测试权限: %s\n", testPerm)

		// 单权限检查
		start := time.Now()
		result := false
		for i := 0; i < iterations; i++ {
			result = checker.Has(userPerms, testPerm)
		}
		totalTime := time.Since(start)
		avgTime := totalTime / time.Duration(iterations)

		fmt.Printf("     执行 %d 次: %v (平均: %v)\n", iterations, totalTime, avgTime)
		fmt.Printf("     结果: %v\n", result)
	}

	fmt.Println("\n======= 性能比对完成 =======")
}
