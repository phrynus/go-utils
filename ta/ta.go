package ta

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"sync"
)

// KlineData 表示一根K线的基本数据结构
// 说明：
//
//	包含了一根K线的所有基本信息，包括：
//	- 开盘时间
//	- OHLCV (开盘价、最高价、最低价、收盘价、成交量)
type KlineData struct {
	StartTime int64   `json:"startTime"` // K线的开始时间戳（毫秒）
	Open      float64 `json:"open"`      // 开盘价
	High      float64 `json:"high"`      // 最高价
	Low       float64 `json:"low"`       // 最低价
	Close     float64 `json:"close"`     // 收盘价
	Volume    float64 `json:"volume"`    // 成交量
}

// KlineDatas 是KlineData的切片类型，代表一组K线数据
type KlineDatas []*KlineData

// fieldCache 用于缓存结构体字段的反射信息
// 说明：
//
//	为了提高性能，缓存了结构体中各个字段的索引位置和类型信息
//	支持不同的字段命名方式和类型转换
type fieldCache struct {
	timeFieldIndex   []int // 时间字段的索引
	openFieldIndex   []int // 开盘价字段的索引
	highFieldIndex   []int // 最高价字段的索引
	lowFieldIndex    []int // 最低价字段的索引
	closeFieldIndex  []int // 收盘价字段的索引
	volumeFieldIndex []int // 成交量字段的索引
	isTimeInt64      bool  // 时间字段是否为int64类型
	isStringPrice    bool  // 价格字段是否为字符串类型
}

// 全局变量定义
var (
	// 支持的各种字段名称映射
	timeFields    = []string{"StartTime", "OpenTime", "Time", "t", "T", "Timestamp", "OpenAt", "EventTime"} // 支持的时间字段名
	openFields    = []string{"Open", "OpenPrice", "O", "o"}                                                 // 支持的开盘价字段名
	highFields    = []string{"High", "HighPrice", "H", "h"}                                                 // 支持的最高价字段名
	lowFields     = []string{"Low", "LowPrice", "L", "l"}                                                   // 支持的最低价字段名
	closeFields   = []string{"Close", "ClosePrice", "C", "c"}                                               // 支持的收盘价字段名
	volumeFields  = []string{"Volume", "Vol", "V", "v", "Amount", "Quantity"}                               // 支持的成交量字段名
	fieldCacheMap = make(map[reflect.Type]*fieldCache)                                                      // 字段缓存映射表
	cacheMutex    sync.RWMutex                                                                              // 缓存读写锁
)

// findAndCacheFields 查找并缓存结构体的字段信息
// 说明：
//
//	使用反射查找结构体中的K线相关字段，并将结果缓存
//	支持多种常见的字段命名方式，提高后续处理效率
//
// 参数：
//   - t: 要分析的结构体类型
//
// 返回值：
//   - *fieldCache: 字段缓存信息
//   - error: 查找过程中的错误，如字段不存在等
func findAndCacheFields(t reflect.Type) (*fieldCache, error) {
	cacheMutex.RLock()
	if cache, ok := fieldCacheMap[t]; ok {
		cacheMutex.RUnlock()
		return cache, nil
	}
	cacheMutex.RUnlock()

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if cache, ok := fieldCacheMap[t]; ok {
		return cache, nil
	}

	cache := &fieldCache{}

	for _, field := range timeFields {
		if f, ok := t.FieldByName(field); ok {
			cache.timeFieldIndex = f.Index
			cache.isTimeInt64 = f.Type.Kind() == reflect.Int64
			break
		}
	}
	if cache.timeFieldIndex == nil {
		return nil, fmt.Errorf("未找到时间字段，支持的字段名：%v", timeFields)
	}

	for _, field := range openFields {
		if f, ok := t.FieldByName(field); ok {
			cache.openFieldIndex = f.Index
			cache.isStringPrice = f.Type.Kind() == reflect.String
			break
		}
	}
	for _, field := range highFields {
		if f, ok := t.FieldByName(field); ok {
			cache.highFieldIndex = f.Index
			break
		}
	}
	for _, field := range lowFields {
		if f, ok := t.FieldByName(field); ok {
			cache.lowFieldIndex = f.Index
			break
		}
	}
	for _, field := range closeFields {
		if f, ok := t.FieldByName(field); ok {
			cache.closeFieldIndex = f.Index
			break
		}
	}
	for _, field := range volumeFields {
		if f, ok := t.FieldByName(field); ok {
			cache.volumeFieldIndex = f.Index
			break
		}
	}

	fieldCacheMap[t] = cache
	return cache, nil
}

// extractKlineData 从反射值中提取K线数据
// 说明：
//
//	从给定的结构体值中提取K线数据的各个字段
//	支持多种数据类型的转换（如string到float64）
//
// 参数：
//   - item: 要处理的结构体值
//   - cache: 字段缓存信息
//
// 返回值：
//   - startTime: K线开始时间
//   - open, high, low, close, volume: OHLCV值（字符串格式）
//   - err: 提取过程中的错误
func extractKlineData(item reflect.Value, cache *fieldCache) (startTime int64, open, high, low, close, volume string, err error) {
	if item.Kind() == reflect.Ptr {
		item = item.Elem()
	}

	timeField := item.FieldByIndex(cache.timeFieldIndex)
	if cache.isTimeInt64 {
		startTime = timeField.Int()
	} else {

		switch timeField.Kind() {
		case reflect.String:
			if t, err := strconv.ParseInt(timeField.String(), 10, 64); err == nil {
				startTime = t
			}
		case reflect.Float64:
			startTime = int64(timeField.Float())
		}
	}

	getValue := func(fieldIndex []int) string {
		if fieldIndex == nil {
			return ""
		}
		field := item.FieldByIndex(fieldIndex)
		if cache.isStringPrice {
			return field.String()
		}

		switch field.Kind() {
		case reflect.Float64:
			return strconv.FormatFloat(field.Float(), 'f', -1, 64)
		case reflect.Int64:
			return strconv.FormatInt(field.Int(), 10)
		}
		return field.String()
	}

	open = getValue(cache.openFieldIndex)
	high = getValue(cache.highFieldIndex)
	low = getValue(cache.lowFieldIndex)
	close = getValue(cache.closeFieldIndex)
	volume = getValue(cache.volumeFieldIndex)

	return
}

// NewKlineDatas 创建新的K线数据集合
// 说明：
//
//	将任意格式的K线数据转换为标准的KlineDatas格式
//	支持并发处理大量数据，自动根据CPU核心数分配工作
//
// 参数：
//   - klines: 输入的K线数据（支持多种格式）
//   - l: 是否排除最后一根K线（通常用于处理未完成的K线）
//
// 返回值：
//   - KlineDatas: 标准格式的K线数据集合
//   - error: 转换过程中的错误
func NewKlineDatas(klines interface{}, l bool) (KlineDatas, error) {
	v := reflect.ValueOf(klines)
	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("输入必须是切片类型")
	}

	length := v.Len()
	if l && length > 0 {
		length--
	}
	if length == 0 {
		return nil, errors.New("没有K线数据")
	}

	klineDataList := make(KlineDatas, length)

	firstItem := v.Index(0)
	if firstItem.Kind() == reflect.Ptr {
		firstItem = firstItem.Elem()
	}
	cache, err := findAndCacheFields(firstItem.Type())
	if err != nil {
		return nil, err
	}

	if length > 1000 {
		var wg sync.WaitGroup
		errChan := make(chan error, length)
		workers := runtime.NumCPU()
		batchSize := length / workers
		if batchSize == 0 {
			batchSize = 1
		}

		for i := 0; i < workers; i++ {
			start := i * batchSize
			end := start + batchSize
			if i == workers-1 {
				end = length
			}

			wg.Add(1)
			go func(start, end int) {
				defer wg.Done()
				for i := start; i < end; i++ {
					startTime, open, high, low, close, volume, err := extractKlineData(v.Index(i), cache)
					if err != nil {
						errChan <- fmt.Errorf("处理第%d条数据时出错: %v", i+1, err)
						return
					}

					if open == "" || high == "" || low == "" || close == "" || volume == "" {
						errChan <- fmt.Errorf("第%d条数据缺少必要字段", i+1)
						return
					}

					o, err1 := strconv.ParseFloat(open, 64)
					h, err2 := strconv.ParseFloat(high, 64)
					l, err3 := strconv.ParseFloat(low, 64)
					c, err4 := strconv.ParseFloat(close, 64)
					v, err5 := strconv.ParseFloat(volume, 64)

					if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil {
						errChan <- fmt.Errorf("第%d条数据转换失败", i+1)
						return
					}

					klineDataList[i] = &KlineData{
						StartTime: startTime,
						Open:      o,
						High:      h,
						Low:       l,
						Close:     c,
						Volume:    v,
					}
				}
			}(start, end)
		}

		go func() {
			wg.Wait()
			close(errChan)
		}()

		for err := range errChan {
			if err != nil {
				return nil, err
			}
		}
	} else {

		for i := 0; i < length; i++ {
			startTime, open, high, low, close, volume, err := extractKlineData(v.Index(i), cache)
			if err != nil {
				return nil, fmt.Errorf("处理第%d条数据时出错: %v", i+1, err)
			}

			if open == "" || high == "" || low == "" || close == "" || volume == "" {
				return nil, fmt.Errorf("第%d条数据缺少必要字段", i+1)
			}

			o, err1 := strconv.ParseFloat(open, 64)
			h, err2 := strconv.ParseFloat(high, 64)
			l, err3 := strconv.ParseFloat(low, 64)
			c, err4 := strconv.ParseFloat(close, 64)
			v, err5 := strconv.ParseFloat(volume, 64)

			if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil {
				return nil, fmt.Errorf("第%d条数据转换失败", i+1)
			}

			klineDataList[i] = &KlineData{
				StartTime: startTime,
				Open:      o,
				High:      h,
				Low:       l,
				Close:     c,
				Volume:    v,
			}
		}
	}

	return klineDataList, nil
}

// ExtractSlice 从K线数据中提取指定类型的价格序列
// 说明：
//
//	从K线数据中提取特定类型的价格数据（如收盘价序列）
//
// 参数：
//   - priceType: 价格类型，支持"open"、"high"、"low"、"close"、"volume"
//
// 返回值：
//   - []float64: 提取的价格序列
//   - error: 提取过程中的错误
func (k *KlineDatas) ExtractSlice(priceType string) ([]float64, error) {
	var prices []float64
	for _, kline := range *k {
		switch priceType {
		case "open":
			prices = append(prices, kline.Open)
		case "high":
			prices = append(prices, kline.High)
		case "low":
			prices = append(prices, kline.Low)
		case "close":
			prices = append(prices, kline.Close)
		case "volume":
			prices = append(prices, kline.Volume)
		}
	}
	return prices, nil
}

// Add 添加一根新的K线数据
// 说明：
//
//	向K线数据集合中添加一根新的K线
//	支持多种输入格式的自动转换
//
// 参数：
//   - wsKline: 要添加的K线数据（支持多种格式）
//
// 返回值：
//   - error: 添加过程中的错误
func (k *KlineDatas) Add(wsKline interface{}) error {
	v := reflect.ValueOf(wsKline)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("数据必须是结构体类型")
	}

	cache, err := findAndCacheFields(v.Type())
	if err != nil {
		return err
	}

	startTime, open, high, low, close, volume, err := extractKlineData(v, cache)
	if err != nil {
		return err
	}

	if open == "" || high == "" || low == "" || close == "" || volume == "" {
		return fmt.Errorf("缺少必要字段")
	}

	o, err1 := strconv.ParseFloat(open, 64)
	h, err2 := strconv.ParseFloat(high, 64)
	l, err3 := strconv.ParseFloat(low, 64)
	c, err4 := strconv.ParseFloat(close, 64)
	v5, err5 := strconv.ParseFloat(volume, 64)

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil {
		return fmt.Errorf("数据转换失败")
	}

	*k = append(*k, &KlineData{
		StartTime: startTime,
		Open:      o,
		High:      h,
		Low:       l,
		Close:     c,
		Volume:    v5,
	})
	return nil
}

// Remove 从开始位置移除指定数量的K线
// 说明：
//
//	从K线数据集合的开始位置删除指定数量的K线
//
// 参数：
//   - n: 要删除的K线数量
//
// 返回值：
//   - error: 删除过程中的错误
func (k *KlineDatas) Remove(n int) error {
	if n <= 0 {
		return fmt.Errorf("删除数量必须大于0")
	}

	if len(*k) < n {
		return fmt.Errorf("要删除的数量(%d)大于现有数据量(%d)", n, len(*k))
	}

	*k = (*k)[n:]
	return nil
}

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
	if len(*k) == 0 {
		return -1
	}
	lastKline := (*k)[len(*k)-1]
	switch source {
	case "open":
		return lastKline.Open
	case "high":
		return lastKline.High
	case "low":
		return lastKline.Low
	case "close":
		return lastKline.Close
	case "volume":
		return lastKline.Volume
	default:
		return -1
	}
}

// preallocateSlices 预分配多个float64切片
// 说明：
//
//	为了提高性能，预先分配指定数量的float64切片
//
// 参数：
//   - length: 每个切片的长度
//   - count: 需要分配的切片数量
//
// 返回值：
//   - [][]float64: 预分配的切片数组
func preallocateSlices(length int, count int) [][]float64 {
	slices := make([][]float64, count)
	for i := range slices {
		slices[i] = make([]float64, length)
	}
	return slices
}

// max 返回两个float64中的较大值
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// min 返回两个float64中的较小值
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
