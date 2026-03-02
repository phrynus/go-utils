package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/phrynus/go-utils"
)

func main() {
	// TestFeiShu()
	// // TestTa()
	// TestDingtalk()
	// TestSystem()
	// TestCex()
	// TestUyzUser()
	// TestPermission()
	// RunMatchExamples()
	// TestFastMatch()

	fmt.Println(utils.Sign("1234567890", &map[string]interface{}{
		"name":   "test",
		"age":    18,
		"active": true,
		"score":  95.5,
		"tags":   []string{"123", "q", "a", "b", "c"},
		"data":   map[string]interface{}{"key": "value", "akey": "value"},
	}, 123, true, "aasd", "", gin.H{}))
}
