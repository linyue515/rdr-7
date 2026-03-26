// Copyright 2017 XUEQIU.COM
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dump

import (
	"container/heap"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/919927181/rdr/decoder"
)

// v1.1.4 add
const (

	DefaultTopBigKeyNum  = 500
	DefaultSeparators = ":;,_- "
	DefaultStoreAllPrefixes        = false
	DefaultTopPrefixNum            = 500
	DefaultPrefixPreShrinkNum      = 5000
	DefaultPrefixContainerMaxCapacity = 50000

    //过期剩余时间分析
	ExpireStat_0_1h_Str       = "0~1h"
	ExpireStat_1_3h_Str       = "0~3h"
	ExpireStat_3_12h_Str      = "3~12h"
	ExpireStat_12_24h_Str     = "12~24h"
	ExpireStat_1_3d_Str       = "1~3d"
	ExpireStat_3_7d_Str       = "3~7d"
	ExpireStat_7d_Str         = ">7d"
	ExpireStat_0_Str          = "永不过期"
	ExpireStat_1_Str          = "已过期"

	DefaultAux_Ctime = 0

)
// v1.1.4 add
type CounterConfig struct {

	// Bigkey数量阈值，默认 500
	TopBigKeyNum int

	// key前缀分隔符，默认 ":;,_- "
	Separators string
	// 是否存储所有前缀，仅当你的主机内存足够时开启，默认关闭
	// false时，对前缀的存储量进行动态收缩, 前缀个数越多，创建的对象也就越多，从而耗内存就越多,就会因内存不足执行不完, 详见 https://github.com/919927181/rdr/issues/1
	StoreAllPrefixes bool
	// 前缀数量阈值，默认 500
	TopPrefixNum int
	// 前缀容器的最大容量，默认值为前缀数量阈值的10倍，可以理解为最大水位线，到达最大水位线，就要去排出一部分水
	PrefixContainerMaxCapacity int
	// 前缀容器的预缩容数量，可以理解为正常水位线
	PrefixPreShrinkNum int

	// rdb的创建时间，用于过期剩余时间分析
	Aux_Ctime int64
}

// 默认配置, v1.1.4 add
func NewCounterConfig() *CounterConfig {
	return &CounterConfig{
		TopBigKeyNum:                DefaultTopBigKeyNum,
		Separators:                  DefaultSeparators,
		StoreAllPrefixes:            DefaultStoreAllPrefixes,
		TopPrefixNum:                DefaultTopPrefixNum,
		PrefixContainerMaxCapacity:  DefaultPrefixContainerMaxCapacity,
		PrefixPreShrinkNum:          DefaultPrefixPreShrinkNum,
		Aux_Ctime:                   DefaultAux_Ctime,
	}
}

// NewCounter return a pointer of Counter
func NewCounter(config *CounterConfig) *Counter {
	if config == nil {
		config = NewCounterConfig()
	}
	h := &entryHeap{}
	heap.Init(h)
	p := &prefixHeap{}
	heap.Init(p)
	return &Counter{
		largestEntries:     h,
		largestKeyPrefixes: p,
		lengthLevel0:       100,
		lengthLevel1:       1000,
		lengthLevel2:       10000,
		lengthLevel3:       100000,
		lengthLevel4:       1000000,
		lengthLevelBytes:   map[typeKey]uint64{},
		lengthLevelNum:     map[typeKey]uint64{},
		keyPrefixBytes:     map[typeKey]uint64{},
		keyPrefixNum:       map[typeKey]uint64{},
		typeBytes:          map[string]uint64{},
		typeNum:            map[string]uint64{},
		slotBytes:          map[int]uint64{},
		slotNum:            map[int]uint64{},
		expireStatBytes:    map[string]uint64{},
		expireStatNum:      map[string]uint64{},
		keyPrefixDb:        map[typeKey]string{},
		config:             config,
	}
}

// Counter for redis memory usage
type Counter struct {
	largestEntries     *entryHeap
	largestKeyPrefixes *prefixHeap
	lengthLevel0       uint64
	lengthLevel1       uint64
	lengthLevel2       uint64
	lengthLevel3       uint64
	lengthLevel4       uint64
	lengthLevelBytes   map[typeKey]uint64
	lengthLevelNum     map[typeKey]uint64
	keyPrefixBytes     map[typeKey]uint64
	keyPrefixNum       map[typeKey]uint64
	typeBytes          map[string]uint64
	typeNum            map[string]uint64
	slotBytes          map[int]uint64
	slotNum            map[int]uint64
	expireStatBytes    map[string]uint64
	expireStatNum      map[string]uint64
	keyPrefixDb        map[typeKey]string
	config             *CounterConfig
}

// Count by various dimensions，show.go NewCounter 后，调用此方法，遍历decoder的entry, <-chan表示一个只能接收数据的单向通道
func (c *Counter) Count(in <-chan *decoder.Entry) {
	for e := range in {
		c.count(e)   //调下面的count方法（串行分析）
	}
	// get largest prefixes
	c.calcuLargestKeyPrefix(c.config.TopPrefixNum)
}

// 传入一个entry，执行各指标的count方法
func (c *Counter) count(e *decoder.Entry) {
	c.countLargestEntries(e, 500)
	c.countByType(e)
	c.countByLength(e)
	c.countByKeyPrefix(e)
	c.countBySlot(e)
	//c.countByDb(e) //该方法由caiqing0204添加
	c.countByExpire(e)

}

// 该方法由caiqing0204添加，没有看到哪儿用到，这里会导致前缀所属db不正确
func (c *Counter) countByDb(e *decoder.Entry) {
	key := typeKey{
		Type: e.Type,
		Key:  e.Key,
	}
	c.keyPrefixDb[key] = strconv.Itoa(e.Db)
}

// GetLargestEntries from heap, num max is 500. 过滤掉小于阈值的key
func (c *Counter) GetLargestEntries(num int, sizeFilter int64) []*decoder.Entry {
	res := []*decoder.Entry{}

	// get a copy of c.largestEntries
	for i := 0; i < c.largestEntries.Len(); i++ {
		entries := *c.largestEntries
		// 阈值默认为0，当大于0时，将过滤掉小于阈值的key
		if sizeFilter > 0 {
			if entries[i].Bytes > uint64(sizeFilter) {
				res = append(res, entries[i])
			}
		} else {
			res = append(res, entries[i])
		}
	}
	sort.Sort(sort.Reverse(entryHeap(res)))
	if num < len(res) {
		res = res[:num]
	}
	return res
}

// GetLargestKeyPrefixes from heap
func (c *Counter) GetLargestKeyPrefixes() []*PrefixEntry {
	res := []*PrefixEntry{}

	// get a copy of c.largestKeyPrefixes
	for i := 0; i < c.largestKeyPrefixes.Len(); i++ {
		entries := *c.largestKeyPrefixes
		res = append(res, entries[i])
	}
	sort.Sort(sort.Reverse(prefixHeap(res)))
	return res
}

// GetLenLevelCount from map
func (c *Counter) GetLenLevelCount() []*PrefixEntry {
	res := []*PrefixEntry{}

	// get a copy of lengthLevelBytes and lengthLevelNum
	for key := range c.lengthLevelBytes {
		entry := &PrefixEntry{}
		entry.Type = key.Type
		entry.Key = key.Key
		entry.Bytes = c.lengthLevelBytes[key]
		entry.Num = c.lengthLevelNum[key]
		entry.Db = c.keyPrefixDb[key]
		res = append(res, entry)
	}
	return res
}

func (c *Counter) countLargestEntries(e *decoder.Entry, num int) {
	heap.Push(c.largestEntries, e)
	l := c.largestEntries.Len()
	if l > num {
		heap.Pop(c.largestEntries)
	}
}

func (c *Counter) countByLength(e *decoder.Entry) {
	key := typeKey{
		Type: e.Type,
		Key:  strconv.FormatUint(c.lengthLevel0, 10),
	}

	add := func(c *Counter, key typeKey, e *decoder.Entry) {
		c.lengthLevelBytes[key] += e.Bytes
		c.lengthLevelNum[key]++
	}

	// must lengthLevel4 > lengthLevel3 > lengthLevel2 ...
	if e.NumOfElem > c.lengthLevel4 {
		key.Key = strconv.FormatUint(c.lengthLevel4, 10)
		add(c, key, e)
	} else if e.NumOfElem > c.lengthLevel3 {
		key.Key = strconv.FormatUint(c.lengthLevel3, 10)
		add(c, key, e)
	} else if e.NumOfElem > c.lengthLevel2 {
		key.Key = strconv.FormatUint(c.lengthLevel2, 10)
		add(c, key, e)
	} else if e.NumOfElem > c.lengthLevel1 {
		key.Key = strconv.FormatUint(c.lengthLevel1, 10)
		add(c, key, e)
	} else if e.NumOfElem > c.lengthLevel0 {
		key.Key = strconv.FormatUint(c.lengthLevel0, 10)
		add(c, key, e)
	}
}

func (c *Counter) countByType(e *decoder.Entry) {
	c.typeNum[e.Type]++
	c.typeBytes[e.Type] += e.Bytes
}

// 过期剩余时间分析，v1.1.5 add
func (c *Counter) countByExpire(e *decoder.Entry) {

    if e.Expiration >0 {
		// rdb的创建时间，是秒间戳，key的过期时间是毫秒时间戳
		// 转换成时间对象后，计算两个时间的差值
		diff := time.Unix(0, e.Expiration*int64(time.Millisecond)).Sub(time.Unix(c.config.Aux_Ctime, 0))
		// 将差值转换为小时数
		h := diff.Hours();
		expireStatStr := ""
		switch  {
        case h < 0:
			expireStatStr = ExpireStat_1_Str
        case h > 0 && h <= 1:
			expireStatStr = ExpireStat_0_1h_Str
		case h > 1 && h <= 3:
			expireStatStr = ExpireStat_1_3h_Str
		case h > 3 && h <= 12:
			expireStatStr = ExpireStat_3_12h_Str
		case h > 12 && h <= 24:
			expireStatStr = ExpireStat_12_24h_Str
		case h > 24 && h <= 24*3:
			expireStatStr = ExpireStat_1_3d_Str
		case h > 24*3 && h <= 24*7:
			expireStatStr = ExpireStat_3_7d_Str
		case h > 24*7:
			expireStatStr = ExpireStat_7d_Str
		}
		c.expireStatNum[expireStatStr]++
		c.expireStatBytes[expireStatStr] += e.Bytes
    } else {
		c.expireStatNum[ExpireStat_0_Str]++
		c.expireStatBytes[ExpireStat_0_Str] += e.Bytes
	}
}
// 传入一个entry，根据key名，通过分隔符得到前缀，然后对各前缀进行计数
func (c *Counter) countByKeyPrefix(e *decoder.Entry) {

	// reset all numbers to 0 将key名字中的所有数字（通常为id号）都置为0
	k := strings.Map(func(c rune) rune {
		if c >= 48 && c <= 57 { //48 == "0" 57 == "9"
			return '*'
		}
		return c
	}, e.Key)

	// 1.将key名字进行分割，得到所有前缀字符串
	prefixes := getPrefixes(k, c.config.Separators)
	key := typeKey{
		Type: e.Type,
	}
	// 2.遍历前缀，对其计数
	for _, prefixStr := range prefixes {
		if len(prefixStr) == 0 {
			continue
		}
		key.Key = prefixStr
		c.keyPrefixBytes[key] += e.Bytes
		c.keyPrefixNum[key]++
		//2025-12-25 liyanjing add 如果不同db里有相同前缀的key，那么设置属于任何一个db都不合适
		if c.keyPrefixDb[key] == "" {
			c.keyPrefixDb[key] = strconv.Itoa(e.Db)
		} else {
			if !strings.Contains(c.keyPrefixDb[key], strconv.Itoa(e.Db)) {
				c.keyPrefixDb[key] += "," + strconv.Itoa(e.Db)
			}
		}
	}
	// 3.如果不开存储所有前缀，并且前缀数量（也可以用 c.keyPrefixNum[key]）超过容器的最大容量时，则进行缩容
	if !c.config.StoreAllPrefixes && len(c.keyPrefixBytes) > c.config.PrefixContainerMaxCapacity {
		c.calcuLargestKeyPrefix(c.config.PrefixPreShrinkNum)
	}
}

func (c *Counter) countBySlot(e *decoder.Entry) {
	if len(e.Key) > 0 {
		slot := Slot(e.Key)
		c.slotNum[slot]++
		c.slotBytes[slot] += e.Bytes
	}
}

// 找出 top num 前缀
func (c *Counter) calcuLargestKeyPrefix(num int) {
	tempPrefixes := &prefixHeap{}
	heap.Init(tempPrefixes)
	for key := range c.keyPrefixBytes {
		k := &PrefixEntry{}
		k.Type = key.Type
		k.Key = key.Key
		k.Bytes = c.keyPrefixBytes[key]
		k.Num = c.keyPrefixNum[key]
		k.Db = c.keyPrefixDb[key]
		delete(c.keyPrefixBytes, key)
		delete(c.keyPrefixNum, key)

		heap.Push(tempPrefixes, k)
		l := tempPrefixes.Len()
		if l > num {
			heap.Pop(tempPrefixes)
		}
	}
	for i := 0; i < tempPrefixes.Len(); i++ {
		entries := *tempPrefixes
		// save
		c.keyPrefixBytes[entries[i].typeKey] = entries[i].Bytes
		c.keyPrefixNum[entries[i].typeKey] = entries[i].Num
	}
	c.largestKeyPrefixes = tempPrefixes
}

type entryHeap []*decoder.Entry

func (h entryHeap) Len() int {
	return len(h)
}
func (h entryHeap) Less(i, j int) bool {
	return h[i].Bytes < h[j].Bytes
}
func (h entryHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *entryHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (h *entryHeap) Push(e interface{}) {
	*h = append(*h, e.(*decoder.Entry))
}

type typeKey struct {
	Type string
	Key  string
}

type prefixHeap []*PrefixEntry

// PrefixEntry record value by prefix
type PrefixEntry struct {
	typeKey
	Bytes uint64
	Num   uint64
	Db    string //之前为int
}

func (h prefixHeap) Len() int {
	return len(h)
}
func (h prefixHeap) Less(i, j int) bool {
	if h[i].Bytes < h[j].Bytes {
		return true
	} else if h[i].Bytes == h[j].Bytes {
		if h[i].Num < h[j].Num {
			return true
		} else if h[i].Num == h[j].Num {
			if h[i].Key > h[j].Key {
				return true
			}
		}
	}
	return false

}
func (h prefixHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *prefixHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (h *prefixHeap) Push(k interface{}) {
	*h = append(*h, k.(*PrefixEntry))
}

func appendIfMissing(slice []int, i int) []int {
	for _, ele := range slice {
		if ele == i {
			return slice
		}
	}
	return append(slice, i)
}

func removeDuplicatesUnordered(elements []string) []string {
	encountered := map[string]bool{}

	// Create a map of all unique elements.
	for v := range elements {
		encountered[elements[v]] = true
	}

	// Place all keys from the map into a slice.
	result := []string{}
	for key := range encountered {
		result = append(result, key)
	}
	return result
}

func getPrefixes(s, sep string) []string {
	res := []string{}
	sepIdx := strings.IndexAny(s, sep)
	if sepIdx < 0 {
		res = append(res, s)
	}
	for sepIdx > -1 {
		r := s[:sepIdx+1]
		if len(res) > 0 {
			r = res[len(res)-1] + s[:sepIdx+1]
		}
		res = append(res, r)
		s = s[sepIdx+1:]
		sepIdx = strings.IndexAny(s, sep)
	}
	// Trim all suffix of separators
	for i := range res {
		for hasAnySuffix(res[i], sep) {
			res[i] = res[i][:len(res[i])-1]
		}
	}
	res = removeDuplicatesUnordered(res)
	return res
}

func hasAnySuffix(s, suffix string) bool {
	for _, c := range suffix {
		if strings.HasSuffix(s, string(c)) {
			return true
		}
	}
	return false
}

// support for sorting of slots
type SlotEntry struct {
	Slot int
	Size uint64
}

type slotHeap []*SlotEntry

func (h slotHeap) Len() int {
	return len(h)
}
func (h slotHeap) Less(i, j int) bool {
	return h[i].Size > h[j].Size
}
func (h slotHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *slotHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (h *slotHeap) Push(e interface{}) {
	*h = append(*h, e.(*SlotEntry))
}
