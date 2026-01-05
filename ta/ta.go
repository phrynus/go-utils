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

// klineExtractor 定义K线数据提取器函数类型
type klineExtractor func(reflect.Value) (*KlineData, error)

// fieldCache 用于缓存结构体字段的反射信息
// 说明：
//
//	为了提高性能，缓存了结构体中各个字段的索引位置和类型信息
//	支持不同的字段命名方式和类型转换
type fieldCache struct {
	timeFieldIndex   []int          // 时间字段的索引
	openFieldIndex   []int          // 开盘价字段的索引
	highFieldIndex   []int          // 最高价字段的索引
	lowFieldIndex    []int          // 最低价字段的索引
	closeFieldIndex  []int          // 收盘价字段的索引
	volumeFieldIndex []int          // 成交量字段的索引
	isTimeInt64      bool           // 时间字段是否为int64类型
	isStringPrice    bool           // 价格字段是否为字符串类型
	extractor        klineExtractor // 预生成的提取器函数，避免重复反射
}

// arrayFieldIndexes 数组格式字段索引结构体
type arrayFieldIndexes struct {
	timeIndex   int // 时间字段在数组中的索引
	openIndex   int // 开盘价字段在数组中的索引
	highIndex   int // 最高价字段在数组中的索引
	lowIndex    int // 最低价字段在数组中的索引
	closeIndex  int // 收盘价字段在数组中的索引
	volumeIndex int // 成交量字段在数组中的索引
}

// arrayExtractorCache 用于缓存数组格式的提取器
type arrayExtractorCache struct {
	indexes   *arrayFieldIndexes // 字段索引
	extractor klineExtractor     // 预生成的提取器函数
}

// extractArrayFieldIndexes 从FieldNames中提取数组字段索引
func extractArrayFieldIndexes(customFields *FieldNames) *arrayFieldIndexes {
	indexes := &arrayFieldIndexes{
		timeIndex:   -1,
		openIndex:   -1,
		highIndex:   -1,
		lowIndex:    -1,
		closeIndex:  -1,
		volumeIndex: -1,
	}

	// 合并字段列表（自定义优先，默认次之）
	timeFieldList := mergeFieldListsWithDefault(customFields, func(f *FieldNames) []string { return f.TimeFields }, timeFields)
	openFieldList := mergeFieldListsWithDefault(customFields, func(f *FieldNames) []string { return f.OpenFields }, openFields)
	highFieldList := mergeFieldListsWithDefault(customFields, func(f *FieldNames) []string { return f.HighFields }, highFields)
	lowFieldList := mergeFieldListsWithDefault(customFields, func(f *FieldNames) []string { return f.LowFields }, lowFields)
	closeFieldList := mergeFieldListsWithDefault(customFields, func(f *FieldNames) []string { return f.CloseFields }, closeFields)
	volumeFieldList := mergeFieldListsWithDefault(customFields, func(f *FieldNames) []string { return f.VolumeFields }, volumeFields)

	// 从字段列表中提取数字索引
	indexes.timeIndex = findNumericIndex(timeFieldList)
	indexes.openIndex = findNumericIndex(openFieldList)
	indexes.highIndex = findNumericIndex(highFieldList)
	indexes.lowIndex = findNumericIndex(lowFieldList)
	indexes.closeIndex = findNumericIndex(closeFieldList)
	indexes.volumeIndex = findNumericIndex(volumeFieldList)

	// 如果没有找到索引，使用默认顺序
	if indexes.timeIndex == -1 {
		indexes.timeIndex = 0
	}
	if indexes.openIndex == -1 {
		indexes.openIndex = 1
	}
	if indexes.highIndex == -1 {
		indexes.highIndex = 2
	}
	if indexes.lowIndex == -1 {
		indexes.lowIndex = 3
	}
	if indexes.closeIndex == -1 {
		indexes.closeIndex = 4
	}
	if indexes.volumeIndex == -1 {
		indexes.volumeIndex = 5
	}

	return indexes
}

// mergeFieldListsWithDefault 合并自定义字段和默认字段
func mergeFieldListsWithDefault(customFields *FieldNames, getFieldFunc func(*FieldNames) []string, defaults []string) []string {
	if customFields == nil {
		return defaults
	}
	custom := getFieldFunc(customFields)
	return mergeFieldLists(custom, defaults)
}

// findNumericIndex 从字段列表中查找数字索引
func findNumericIndex(fieldList []string) int {
	for _, field := range fieldList {
		if index, err := strconv.Atoi(field); err == nil && index >= 0 {
			return index
		}
	}
	return -1
}

// 全局变量定义
var (
	// 支持的各种字段名称映射
	timeFields             = []string{"0", "StartTime", "OpenTime", "Time", "t", "T", "Timestamp", "OpenAt", "EventTime"} // 支持的时间字段名
	openFields             = []string{"1", "Open", "OpenPrice", "O", "o"}                                                 // 支持的开盘价字段名
	highFields             = []string{"2", "High", "HighPrice", "H", "h"}                                                 // 支持的最高价字段名
	lowFields              = []string{"3", "Low", "LowPrice", "L", "l"}                                                   // 支持的最低价字段名
	closeFields            = []string{"4", "Close", "ClosePrice", "C", "c"}                                               // 支持的收盘价字段名
	volumeFields           = []string{"5", "Volume", "Vol", "V", "v", "Amount", "Quantity"}                               // 支持的成交量字段名
	fieldCacheMap          = make(map[reflect.Type]*fieldCache)                                                           // 字段缓存映射表
	arrayExtractorCacheMap = make(map[string]*arrayExtractorCache)                                                        // 数组提取器缓存映射表
	cacheMutex             sync.RWMutex                                                                                   // 缓存读写锁
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

		// 生成并缓存提取器函数
		cache.extractor = generateStructExtractor(cache)

		fieldCacheMap[t] = cache
		return cache, nil
	}

	// 有自定义字段名称，不使用缓存（因为自定义字段名称的情况较少）
	cache := &fieldCache{}
	if err := findFields(t, cache, customFields); err != nil {
		return nil, err
	}
	// 生成提取器函数
	cache.extractor = generateStructExtractor(cache)
	return cache, nil
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

// generateStructExtractor 生成结构体格式的K线数据提取器
func generateStructExtractor(cache *fieldCache) klineExtractor {
	return func(item reflect.Value) (*KlineData, error) {
		if item.Kind() == reflect.Ptr {
			item = item.Elem()
		}

		var startTime int64
		var open, high, low, close, volume float64

		// 提取时间字段
		timeField := item.FieldByIndex(cache.timeFieldIndex)
		if cache.isTimeInt64 {
			startTime = timeField.Int()
		} else {
			switch timeField.Kind() {
			case reflect.String:
				if t, parseErr := strconv.ParseInt(timeField.String(), 10, 64); parseErr == nil {
					startTime = t
				} else {
					return nil, fmt.Errorf("时间字段转换失败")
				}
			case reflect.Float64:
				startTime = int64(timeField.Float())
			default:
				return nil, fmt.Errorf("不支持的时间字段类型: %v", timeField.Kind())
			}
		}

		// 提取价格字段的辅助函数
		extractPrice := func(fieldIndex []int) (float64, error) {
			if fieldIndex == nil {
				return 0, fmt.Errorf("字段索引为空")
			}
			field := item.FieldByIndex(fieldIndex)
			if cache.isStringPrice {
				return strconv.ParseFloat(field.String(), 64)
			}
			switch field.Kind() {
			case reflect.Float64:
				return field.Float(), nil
			case reflect.Int64:
				return float64(field.Int()), nil
			case reflect.String:
				return strconv.ParseFloat(field.String(), 64)
			default:
				return 0, fmt.Errorf("不支持的价格字段类型: %v", field.Kind())
			}
		}

		var err error
		if open, err = extractPrice(cache.openFieldIndex); err != nil {
			return nil, fmt.Errorf("开盘价字段转换失败: %v", err)
		}
		if high, err = extractPrice(cache.highFieldIndex); err != nil {
			return nil, fmt.Errorf("最高价字段转换失败: %v", err)
		}
		if low, err = extractPrice(cache.lowFieldIndex); err != nil {
			return nil, fmt.Errorf("最低价字段转换失败: %v", err)
		}
		if close, err = extractPrice(cache.closeFieldIndex); err != nil {
			return nil, fmt.Errorf("收盘价字段转换失败: %v", err)
		}
		if volume, err = extractPrice(cache.volumeFieldIndex); err != nil {
			return nil, fmt.Errorf("成交量字段转换失败: %v", err)
		}

		return &KlineData{
			StartTime: startTime,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
		}, nil
	}
}

// getArrayExtractor 获取或创建数组格式的K线数据提取器
func getArrayExtractor(customFields *FieldNames) klineExtractor {
	// 生成缓存键
	key := "default"
	if customFields != nil {
		// 使用字段配置的哈希作为键（简化实现，实际可优化）
		key = fmt.Sprintf("%v-%v-%v-%v-%v-%v",
			customFields.TimeFields,
			customFields.OpenFields,
			customFields.HighFields,
			customFields.LowFields,
			customFields.CloseFields,
			customFields.VolumeFields)
	}

	cacheMutex.RLock()
	if cache, ok := arrayExtractorCacheMap[key]; ok {
		cacheMutex.RUnlock()
		return cache.extractor
	}
	cacheMutex.RUnlock()

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// 双重检查
	if cache, ok := arrayExtractorCacheMap[key]; ok {
		return cache.extractor
	}

	// 创建新的缓存
	indexes := extractArrayFieldIndexes(customFields)
	cache := &arrayExtractorCache{
		indexes:   indexes,
		extractor: generateArrayExtractor(indexes),
	}
	arrayExtractorCacheMap[key] = cache
	return cache.extractor
}

// generateArrayExtractor 生成数组格式的K线数据提取器
func generateArrayExtractor(indexes *arrayFieldIndexes) klineExtractor {
	return func(item reflect.Value) (*KlineData, error) {
		if item.Kind() == reflect.Ptr {
			item = item.Elem()
		}

		arrayLen := item.Len()
		maxIndex := max(indexes.timeIndex, indexes.openIndex, indexes.highIndex, indexes.lowIndex, indexes.closeIndex, indexes.volumeIndex)
		if arrayLen <= maxIndex {
			return nil, fmt.Errorf("数组长度不足，需要至少%d个元素，当前只有%d个", maxIndex+1, arrayLen)
		}

		var startTime int64
		var open, high, low, close, volume float64

		// 提取时间字段
		timeElem := item.Index(indexes.timeIndex)
		// 处理 interface{} 类型
		if timeElem.Kind() == reflect.Interface {
			timeElem = timeElem.Elem()
		}

		switch timeElem.Kind() {
		case reflect.String:
			timeStr := timeElem.String()
			if t, parseErr := strconv.ParseInt(timeStr, 10, 64); parseErr == nil {
				startTime = t
			} else if t, parseErr := strconv.ParseFloat(timeStr, 64); parseErr == nil {
				startTime = int64(t)
			} else {
				return nil, fmt.Errorf("时间字段转换失败")
			}
		case reflect.Float64:
			startTime = int64(timeElem.Float())
		case reflect.Int, reflect.Int64:
			startTime = timeElem.Int()
		default:
			return nil, fmt.Errorf("不支持的时间字段类型: %v", timeElem.Kind())
		}

		// 辅助函数：提取数值元素
		extractNumeric := func(index int) (float64, error) {
			if index < 0 || index >= arrayLen {
				return 0, fmt.Errorf("索引超出范围: %d", index)
			}
			elem := item.Index(index)
			// 处理 interface{} 类型
			if elem.Kind() == reflect.Interface {
				elem = elem.Elem()
			}

			switch elem.Kind() {
			case reflect.Float64:
				return elem.Float(), nil
			case reflect.Float32:
				return float64(elem.Float()), nil
			case reflect.Int, reflect.Int64:
				return float64(elem.Int()), nil
			case reflect.Int32:
				return float64(elem.Int()), nil
			case reflect.Uint, reflect.Uint64:
				return float64(elem.Uint()), nil
			case reflect.Uint32:
				return float64(elem.Uint()), nil
			case reflect.String:
				return strconv.ParseFloat(elem.String(), 64)
			default:
				return 0, fmt.Errorf("不支持的数值类型: %v", elem.Kind())
			}
		}

		var err error
		if open, err = extractNumeric(indexes.openIndex); err != nil {
			return nil, fmt.Errorf("开盘价字段转换失败: %v", err)
		}
		if high, err = extractNumeric(indexes.highIndex); err != nil {
			return nil, fmt.Errorf("最高价字段转换失败: %v", err)
		}
		if low, err = extractNumeric(indexes.lowIndex); err != nil {
			return nil, fmt.Errorf("最低价字段转换失败: %v", err)
		}
		if close, err = extractNumeric(indexes.closeIndex); err != nil {
			return nil, fmt.Errorf("收盘价字段转换失败: %v", err)
		}
		if volume, err = extractNumeric(indexes.volumeIndex); err != nil {
			return nil, fmt.Errorf("成交量字段转换失败: %v", err)
		}

		return &KlineData{
			StartTime: startTime,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
		}, nil
	}
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

	// 结构体格式的处理
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

// extractKlineDataFromArray 从数组格式的K线数据中提取数据
func extractKlineDataFromArray(item reflect.Value, indexes *arrayFieldIndexes) (startTime int64, open, high, low, close, volume string, err error) {
	arrayLen := item.Len()
	maxIndex := max(indexes.timeIndex, indexes.openIndex, indexes.highIndex, indexes.lowIndex, indexes.closeIndex, indexes.volumeIndex)
	if arrayLen <= maxIndex {
		return 0, "", "", "", "", "", fmt.Errorf("数组长度不足，需要至少%d个元素，当前只有%d个", maxIndex+1, arrayLen)
	}

	// 辅助函数：将数组元素转换为字符串
	convertToString := func(index int) string {
		if index < 0 || index >= arrayLen {
			return ""
		}
		elem := item.Index(index)
		switch elem.Kind() {
		case reflect.String:
			return elem.String()
		case reflect.Float64:
			return strconv.FormatFloat(elem.Float(), 'f', -1, 64)
		case reflect.Float32:
			return strconv.FormatFloat(elem.Float(), 'f', -1, 32)
		case reflect.Int, reflect.Int64:
			return strconv.FormatInt(elem.Int(), 10)
		case reflect.Int32:
			return strconv.FormatInt(elem.Int(), 10)
		case reflect.Uint, reflect.Uint64:
			return strconv.FormatUint(elem.Uint(), 10)
		case reflect.Uint32:
			return strconv.FormatUint(elem.Uint(), 10)
		default:
			return fmt.Sprintf("%v", elem.Interface())
		}
	}

	// 根据配置的索引提取字段
	timeStr := convertToString(indexes.timeIndex)
	if t, parseErr := strconv.ParseInt(timeStr, 10, 64); parseErr == nil {
		startTime = t
	} else if t, parseErr := strconv.ParseFloat(timeStr, 64); parseErr == nil {
		startTime = int64(t)
	}

	open = convertToString(indexes.openIndex)
	high = convertToString(indexes.highIndex)
	low = convertToString(indexes.lowIndex)
	close = convertToString(indexes.closeIndex)
	volume = convertToString(indexes.volumeIndex)

	return
}

// max 返回多个整数中的最大值
func max(values ...int) int {
	if len(values) == 0 {
		return 0
	}
	maxVal := values[0]
	for _, v := range values[1:] {
		if v > maxVal {
			maxVal = v
		}
	}
	return maxVal
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

	// 检查是否为数组格式的K线数据
	isArrayFormat := firstItem.Kind() == reflect.Slice || firstItem.Kind() == reflect.Array

	var extractor klineExtractor
	var err error

	// 处理自定义字段配置
	var customFieldsPtr *FieldNames
	if len(customFields) > 0 && customFields[0] != nil {
		customFieldsPtr = customFields[0]
	}

	if isArrayFormat {
		// 数组格式：获取或创建提取器
		extractor = getArrayExtractor(customFieldsPtr)
	} else {
		// 结构体格式：获取字段缓存和提取器
		var cache *fieldCache
		cache, err = findAndCacheFields(firstItem.Type(), customFieldsPtr)
		if err != nil {
			return nil, err
		}
		extractor = cache.extractor
	}

	// 提取单条K线转换函数，直接使用预生成的提取器
	convertKlineItem := func(index int) error {
		item := v.Index(index)
		klineData, extractErr := extractor(item)
		if extractErr != nil {
			return fmt.Errorf("处理第%d条数据时出错: %v", index+1, extractErr)
		}
		klineDataList[index] = klineData
		return nil
	}

	// 并发处理大量数据
	// 限制 worker 数量，避免创建过多 goroutine
	workers := runtime.NumCPU()
	if workers > length {
		workers = length
	}

	if workers <= 1 {
		// 数据量小，直接顺序处理
		for i := 0; i < length; i++ {
			if err := convertKlineItem(i); err != nil {
				return nil, err
			}
		}
	} else {
		// 并发处理
		var wg sync.WaitGroup
		errChan := make(chan error, 1) // 只接收第一个错误

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
							// 只发送第一个错误
						default:
							// 通道已满，说明已有其他错误被处理
						}
						return
					}
				}
			}(start, end)
		}

		// 等待所有 goroutine 完成
		go func() {
			wg.Wait()
			close(errChan)
		}()

		// 检查是否有错误
		if err := <-errChan; err != nil {
			return nil, err
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
	if len(*k) == 0 {
		return nil, nil
	}

	// 预分配切片避免动态扩容
	prices := make([]float64, len(*k))
	for i, kline := range *k {
		switch priceType {
		case "open":
			prices[i] = kline.Open
		case "high":
			prices[i] = kline.High
		case "low":
			prices[i] = kline.Low
		case "close":
			prices[i] = kline.Close
		case "volume":
			prices[i] = kline.Volume
		default:
			return nil, fmt.Errorf("不支持的价格类型: %s", priceType)
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

	// 检查是否为数组格式的K线数据
	isArrayFormat := v.Kind() == reflect.Slice || v.Kind() == reflect.Array

	var extractor klineExtractor
	var err error

	// 处理自定义字段配置
	var customFieldsPtr *FieldNames
	if len(customFields) > 0 && customFields[0] != nil {
		customFieldsPtr = customFields[0]
	}

	if isArrayFormat {
		// 数组格式：获取或创建提取器
		extractor = getArrayExtractor(customFieldsPtr)
	} else {
		// 结构体格式：获取字段缓存和提取器
		if v.Kind() != reflect.Struct {
			return fmt.Errorf("数据必须是结构体或数组类型")
		}
		var cache *fieldCache
		cache, err = findAndCacheFields(v.Type(), customFieldsPtr)
		if err != nil {
			return err
		}
		extractor = cache.extractor
	}

	// 使用提取器获取K线数据
	klineData, err := extractor(v)
	if err != nil {
		return err
	}

	*k = append(*k, klineData)
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
