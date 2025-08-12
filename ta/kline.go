package ta

import "fmt"

// Keep 保留最后N根K线并返回新的数据集
// 说明：
//
//	创建新的数据集，只包含原数据集最后N根K线
//
// 参数：
//   - n: 要保留的K线数量
//
// 返回值：
//   - KlineDatas: 新的K线数据集合
//   - error: 处理过程中的错误
func (k *KlineDatas) Keep(n int) (KlineDatas, error) {
	if n <= 0 {
		return nil, fmt.Errorf("保留数量必须大于0")
	}

	if len(*k) < n {
		return nil, fmt.Errorf("要保留的数量(%d)大于现有数据量(%d)", n, len(*k))
	}

	newK := make(KlineDatas, n)
	copy(newK, (*k)[len(*k)-n:])
	return newK, nil
}

// Keep_ 保留最后N根K线（直接修改原数据）
// 说明：
//
//	直接修改原数据集，只保留最后N根K线
//
// 参数：
//   - n: 要保留的K线数量
//
// 返回值：
//   - error: 处理过程中的错误
func (k *KlineDatas) Keep_(n int) error {
	if n <= 0 {
		return fmt.Errorf("保留数量必须大于0")
	}

	if len(*k) < n {
		return fmt.Errorf("要保留的数量(%d)大于现有数据量(%d)", n, len(*k))
	}

	*k = (*k)[len(*k)-n:]
	return nil
}

// GetLast 获取最后一根K线的指定数据
// 说明：
//
//	获取最后一根K线的特定价格数据
//
// 参数：
//   - source: 数据类型，支持"open"、"high"、"low"、"close"、"volume"
//
// 返回值：
//   - float64: 请求的价格数据，如果数据不存在返回-1
func (k *KlineDatas) GetLast(source string) float64 {
	total := len(*k)
	if total == 0 {
		return -1
	}
	return (*k)[total-1].Value(source)
}

// GetLastN 获取最后指定K线的指定数据
//
// 说明：
//
//	获取最后指定K线的特定价格数据
//
// 参数：
//   - n: 要获取的K位置，倒序取值 0为最后一根 1为倒数第二根 2为倒数第三根
//   - source: 数据类型，支持"open"、"high"、"low"、"close"、"volume"
//
// 返回值：
//   - float64: 请求的价格数据，如果数据不存在返回-1
func (k *KlineDatas) GetLastN(n int, source string) float64 {
	total := len(*k)
	if total == 0 {
		return -1
	}

	// 检查索引是否有效
	if n < 0 || n >= total {
		return -1
	}

	// 计算实际的索引位置（从后往前数）
	index := total - 1 - n
	return (*k)[index].Value(source)
}

// GetKlineValue 从单个K线数据中获取指定类型的值
// 说明：
//
//	从给定的K线数据中获取特定类型的价格或成交量数据
//	支持多种格式的参数输入
//
// 参数：
//   - source: 数据类型，支持多种格式：
//   - 完整名称: "open"、"high"、"low"、"close"、"volume"
//   - 数字格式: "1"(open)、"2"(high)、"3"(low)、"4"(close)、"5"(volume)
//   - 简写格式: "o"(open)、"h"(high)、"l"(low)、"c"(close)、"v"(volume)
//
// 返回值：
//   - float64: 请求的价格数据，如果数据类型不支持返回-1
func (kline *KlineData) Value(source string) float64 {
	if kline == nil {
		return -1
	}

	switch source {
	case "open", "1", "o":
		return kline.Open
	case "high", "2", "h":
		return kline.High
	case "low", "3", "l":
		return kline.Low
	case "close", "4", "c":
		return kline.Close
	case "volume", "5", "v":
		return kline.Volume
	default:
		return -1
	}
}
