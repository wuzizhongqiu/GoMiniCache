package dict

/*
 * 使用分段锁 map 作为底层存储（TODO）
 */

import (
	"sync"
)

// LockDick 一个非并发安全的 map
type LockDick struct {
	m  map[string]interface{}
	mu sync.Mutex
}

// MakeLockDick 创建
func MakeLockDick() *LockDick {
	return &LockDick{
		m: make(map[string]interface{}),
	}
}

// Get 返回绑定值以及键是否存在
func (dict *LockDick) Get(key string) (val interface{}, exists bool) {
	val, ok := dict.m[key]
	return val, ok
}

// Len 返回字典的个数
func (dict *LockDick) Len() int {
	if dict.m == nil {
		panic("m is nil")
	}
	return len(dict.m)
}

// Put 将键值放入字典并返回新插入的键值的个数
func (dict *LockDick) Put(key string, val interface{}) (result int) {
	_, existed := dict.m[key]
	dict.m[key] = val
	if existed {
		return 0
	}
	return 1
}

// PutIfAbsent 如果键不存在，则返回值，并返回更新的键值的个数
func (dict *LockDick) PutIfAbsent(key string, val interface{}) (result int) {
	_, existed := dict.m[key]
	if existed {
		return 0
	}
	dict.m[key] = val
	return 1
}

// PutIfExists 如果键存在则放值，并返回插入的键值的个数
func (dict *LockDick) PutIfExists(key string, val interface{}) (result int) {
	_, existed := dict.m[key]
	if existed {
		dict.m[key] = val
		return 1
	}
	return 0
}

// Remove 删除键并返回已删除的键值的个数
func (dict *LockDick) Remove(key string) (result int) {
	_, existed := dict.m[key]
	delete(dict.m, key)
	if existed {
		return 1
	}
	return 0
}

// Keys 返回字典中的所有键
func (dict *LockDick) Keys() []string {
	result := make([]string, len(dict.m))
	i := 0
	for k := range dict.m {
		result[i] = k
	}
	return result
}

// ForEach 遍历字典
func (dict *LockDick) ForEach(consumer Consumer) {
	for k, v := range dict.m {
		if !consumer(k, v) {
			break
		}
	}
}

// RandomKeys 随机返回给定数字的键，可能包含重复的键
func (dict *LockDick) RandomKeys(limit int) []string {
	result := make([]string, limit)
	for i := 0; i < limit; i++ {
		for k := range dict.m {
			result[i] = k
			break
		}
	}
	return result
}

// RandomDistinctKeys 随机返回给定数字的键，不会包含重复的键
func (dict *LockDick) RandomDistinctKeys(limit int) []string {
	size := limit
	if size > len(dict.m) {
		size = len(dict.m)
	}
	result := make([]string, size)
	i := 0
	for k := range dict.m {
		if i == limit {
			break
		}
		result[i] = k
		i++
	}
	return result
}

// Clear 清空字典
func (dict *LockDick) Clear() {
	*dict = *MakeLockDick()
}
