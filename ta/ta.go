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

// FieldNames 自定义字段名称配置
// 用于扩展支持的字段名称，如果某个字段为 nil 或空，则使用默认字段名称
type FieldNames struct {
	TimeFields   []string // 自定义时间字段名称
	OpenFields   []string // 自定义开盘价字段名称
	HighFields   []string // 自定义最高价字段名称
	LowFields    []string // 自定义最低价字段名称
	CloseFields  []string // 自定义收盘价字段名称
	VolumeFields []string // 自定义成交量字段名称
}

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
//	如果提供了自定义字段名称，则优先使用自定义字段名称
//
// 参数：
//   - t: 要分析的结构体类型
//   - customFields: 自定义字段名称，如果为 nil 则使用默认字段名称
//
// 返回值：
//   - *fieldCache: 字段缓存信息
//   - error: 查找过程中的错误，如字段不存在等
func findAndCacheFields(t reflect.Type, customFields *FieldNames) (*fieldCache, error) {
	// 如果没有自定义字段名称，使用缓存
	if customFields == nil {
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
		if err := findFields(t, cache, nil); err != nil {
			return nil, err
		}

		fieldCacheMap[t] = cache
		return cache, nil
	}

	// 有自定义字段名称，不使用缓存（因为自定义字段名称的情况较少）
	cache := &fieldCache{}
	return cache, findFields(t, cache, customFields)
}

// findFields 查找字段的通用逻辑
func findFields(t reflect.Type, cache *fieldCache, customFields *FieldNames) error {
	// 如果没有自定义字段，直接使用默认字段列表（最快路径）
	if customFields == nil {
		return findFieldsWithList(t, cache, timeFields, openFields, highFields, lowFields, closeFields, volumeFields)
	}

	// 有自定义字段，合并自定义字段和默认字段（自定义字段优先，去重）
	timeFieldList := mergeFieldLists(customFields.TimeFields, timeFields)
	openFieldList := mergeFieldLists(customFields.OpenFields, openFields)
	highFieldList := mergeFieldLists(customFields.HighFields, highFields)
	lowFieldList := mergeFieldLists(customFields.LowFields, lowFields)
	closeFieldList := mergeFieldLists(customFields.CloseFields, closeFields)
	volumeFieldList := mergeFieldLists(customFields.VolumeFields, volumeFields)

	return findFieldsWithList(t, cache, timeFieldList, openFieldList, highFieldList, lowFieldList, closeFieldList, volumeFieldList)
}

// mergeFieldLists 合并两个字段列表，自定义字段在前，默认字段在后，去重
func mergeFieldLists(custom, defaults []string) []string {
	if len(custom) == 0 {
		return defaults
	}
	// 构建自定义字段的集合用于快速查找
	customSet := make(map[string]bool, len(custom))
	for _, f := range custom {
		customSet[f] = true
	}
	// 合并列表，自定义字段在前
	result := make([]string, len(custom), len(custom)+len(defaults))
	copy(result, custom)
	// 添加不在自定义字段中的默认字段
	for _, f := range defaults {
		if !customSet[f] {
			result = append(result, f)
		}
	}
	return result
}

// findFieldsWithList 使用指定的字段列表查找字段
func findFieldsWithList(t reflect.Type, cache *fieldCache, timeFields, openFields, highFields, lowFields, closeFields, volumeFields []string) error {
	// 查找时间字段
	for _, field := range timeFields {
		if f, ok := t.FieldByName(field); ok {
			cache.timeFieldIndex = f.Index
			cache.isTimeInt64 = f.Type.Kind() == reflect.Int64
			break
		}
	}
	if cache.timeFieldIndex == nil {
		return fmt.Errorf("未找到时间字段，支持的字段名：%v", timeFields)
	}

	// 查找开盘价字段
	for _, field := range openFields {
		if f, ok := t.FieldByName(field); ok {
			cache.openFieldIndex = f.Index
			cache.isStringPrice = f.Type.Kind() == reflect.String
			break
		}
	}

	// 查找最高价字段
	for _, field := range highFields {
		if f, ok := t.FieldByName(field); ok {
			cache.highFieldIndex = f.Index
			break
		}
	}

	// 查找最低价字段
	for _, field := range lowFields {
		if f, ok := t.FieldByName(field); ok {
			cache.lowFieldIndex = f.Index
			break
		}
	}

	// 查找收盘价字段
	for _, field := range closeFields {
		if f, ok := t.FieldByName(field); ok {
			cache.closeFieldIndex = f.Index
			break
		}
	}

	// 查找成交量字段
	for _, field := range volumeFields {
		if f, ok := t.FieldByName(field); ok {
			cache.volumeFieldIndex = f.Index
			break
		}
	}

	return nil
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

	// 提取时间字段
	timeField := item.FieldByIndex(cache.timeFieldIndex)
	if cache.isTimeInt64 {
		startTime = timeField.Int()
	} else {
		switch timeField.Kind() {
		case reflect.String:
			if t, parseErr := strconv.ParseInt(timeField.String(), 10, 64); parseErr == nil {
				startTime = t
			}
		case reflect.Float64:
			startTime = int64(timeField.Float())
		}
	}

	// 内联字段值提取函数，减少函数调用开销
	getFieldValue := func(fieldIndex []int) string {
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
		case reflect.String:
			return field.String()
		}
		return ""
	}

	open = getFieldValue(cache.openFieldIndex)
	high = getFieldValue(cache.highFieldIndex)
	low = getFieldValue(cache.lowFieldIndex)
	close = getFieldValue(cache.closeFieldIndex)
	volume = getFieldValue(cache.volumeFieldIndex)

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
//   - customFields: 可选的自定义字段名称，用于扩展支持的字段名称
//
// 返回值：
//   - KlineDatas: 标准格式的K线数据集合
//   - error: 转换过程中的错误
func NewKlineDatas(klines interface{}, l bool, customFields ...*FieldNames) (KlineDatas, error) {
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

	var customFieldsPtr *FieldNames
	if len(customFields) > 0 && customFields[0] != nil {
		customFieldsPtr = customFields[0]
	}

	cache, err := findAndCacheFields(firstItem.Type(), customFieldsPtr)
	if err != nil {
		return nil, err
	}

	// 提取单条K线转换函数，减少代码重复
	convertKlineItem := func(index int) error {
		item := v.Index(index)
		startTime, open, high, low, close, volume, extractErr := extractKlineData(item, cache)
		if extractErr != nil {
			return fmt.Errorf("处理第%d条数据时出错: %v", index+1, extractErr)
		}

		if open == "" || high == "" || low == "" || close == "" || volume == "" {
			return fmt.Errorf("第%d条数据缺少必要字段", index+1)
		}

		o, err1 := strconv.ParseFloat(open, 64)
		h, err2 := strconv.ParseFloat(high, 64)
		l, err3 := strconv.ParseFloat(low, 64)
		c, err4 := strconv.ParseFloat(close, 64)
		vol, err5 := strconv.ParseFloat(volume, 64)

		if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil {
			return fmt.Errorf("第%d条数据转换失败", index+1)
		}

		klineDataList[index] = &KlineData{
			StartTime: startTime,
			Open:      o,
			High:      h,
			Low:       l,
			Close:     c,
			Volume:    vol,
		}
		return nil
	}

	// 并发处理大量数据
	var wg sync.WaitGroup
	errChan := make(chan error, 1) // 只需要一个错误通道
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
			for idx := start; idx < end; idx++ {
				if err := convertKlineItem(idx); err != nil {
					select {
					case errChan <- err:
					default:
					}
					return
				}
			}
		}(start, end)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	if err := <-errChan; err != nil {
		return nil, err
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
//   - customFields: 可选的自定义字段名称，用于扩展支持的字段名称
//
// 返回值：
//   - error: 添加过程中的错误
func (k *KlineDatas) Add(wsKline interface{}, customFields ...*FieldNames) error {
	v := reflect.ValueOf(wsKline)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("数据必须是结构体类型")
	}

	var customFieldsPtr *FieldNames
	if len(customFields) > 0 && customFields[0] != nil {
		customFieldsPtr = customFields[0]
	}

	cache, err := findAndCacheFields(v.Type(), customFieldsPtr)
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
