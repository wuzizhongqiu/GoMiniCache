package dict

/*
 * 使用 sync.Map 作为数据库底层存储
 */

import "sync"

// SyncDict 使用 sync.Map 作为底层存储
type SyncDict struct {
	m sync.Map
}

// MakeSyncDict 创建
func MakeSyncDict() *SyncDict {
	return &SyncDict{}
}

// Get 返回绑定值以及键是否存在
func (dict *SyncDict) Get(key string) (val interface{}, exists bool) {
	val, ok := dict.m.Load(key)
	return val, ok
}

// Len 返回字典的个数
func (dict *SyncDict) Len() int {
	length := 0
	dict.m.Range(func(k, v interface{}) bool {
		length++
		return true
	})
	return length
}

// Put 将键值放入字典并返回新插入的键值的个数
func (dict *SyncDict) Put(key string, val interface{}) (result int) {
	_, existed := dict.m.Load(key)
	dict.m.Store(key, val)
	if existed {
		return 0
	}
	return 1
}

// PutIfAbsent 如果键不存在，则返回值，并返回更新的键值的个数
func (dict *SyncDict) PutIfAbsent(key string, val interface{}) (result int) {
	_, existed := dict.m.Load(key)
	if existed {
		return 0
	}
	dict.m.Store(key, val)
	return 1
}

// PutIfExists 如果键存在则放值，并返回插入的键值的个数
func (dict *SyncDict) PutIfExists(key string, val interface{}) (result int) {
	_, existed := dict.m.Load(key)
	if existed {
		dict.m.Store(key, val)
		return 1
	}
	return 0
}

// Remove 删除键并返回已删除的键值的个数
func (dict *SyncDict) Remove(key string) (result int) {
	_, existed := dict.m.Load(key)
	dict.m.Delete(key)
	if existed {
		return 1
	}
	return 0
}

// Keys 返回字典中的所有键
func (dict *SyncDict) Keys() []string {
	result := make([]string, dict.Len())
	i := 0
	dict.m.Range(func(key, value interface{}) bool {
		result[i] = key.(string)
		i++
		return true
	})
	return result
}

// ForEach 遍历字典
func (dict *SyncDict) ForEach(consumer Consumer) {
	dict.m.Range(func(key, value interface{}) bool {
		consumer(key.(string), value)
		return true
	})
}

// RandomKeys 随机返回给定数字的键，可能包含重复的键
func (dict *SyncDict) RandomKeys(limit int) []string {
	result := make([]string, limit)
	for i := 0; i < limit; i++ {
		dict.m.Range(func(key, value interface{}) bool {
			result[i] = key.(string)
			return false
		})
	}
	return result

}

// RandomDistinctKeys 随机返回给定数字的键，不会包含重复的键
func (dict *SyncDict) RandomDistinctKeys(limit int) []string {
	result := make([]string, limit)
	i := 0
	dict.m.Range(func(key, value interface{}) bool {
		result[i] = key.(string)
		i++
		if i == limit {
			return false
		}
		return true
	})
	return result
}

// Clear 删除字典中的所有键
func (dict *SyncDict) Clear() {
	*dict = *MakeSyncDict()
}
