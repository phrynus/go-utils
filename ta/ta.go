package ta

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"sync"
)

type KlineData struct {
	StartTime int64   `json:"startTime"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
}

type KlineDatas []*KlineData

type fieldCache struct {
	timeFieldIndex   []int
	openFieldIndex   []int
	highFieldIndex   []int
	lowFieldIndex    []int
	closeFieldIndex  []int
	volumeFieldIndex []int
	isTimeInt64      bool
	isStringPrice    bool
}

var (
	timeFields    = []string{"StartTime", "OpenTime", "Time", "t", "T", "Timestamp", "OpenAt", "EventTime"}
	openFields    = []string{"Open", "OpenPrice", "O", "o"}
	highFields    = []string{"High", "HighPrice", "H", "h"}
	lowFields     = []string{"Low", "LowPrice", "L", "l"}
	closeFields   = []string{"Close", "ClosePrice", "C", "c"}
	volumeFields  = []string{"Volume", "Vol", "V", "v", "Amount", "Quantity"}
	fieldCacheMap = make(map[reflect.Type]*fieldCache)
	cacheMutex    sync.RWMutex
)

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

func preallocateSlices(length int, count int) [][]float64 {
	slices := make([][]float64, count)
	for i := range slices {
		slices[i] = make([]float64, length)
	}
	return slices
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
