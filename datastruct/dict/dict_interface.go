package dict

/*
 * 数据库底层存储封装的方法的抽象
 */

// Consumer 用于遍历字典，如果返回 false 则遍历中断（sync.Map 的遍历规则，传入一个方法）
type Consumer func(key string, val interface{}) bool

// Dict 是 kv 存储数据结构的抽象（如果之后要改底层存储的实现，只需要修改实现即可，接口已经定义好了）
type Dict interface {
	Get(key string) (val interface{}, exists bool)        // key 获取 val（以及 key 是否存在）
	Len() int                                             // 返回数据长度
	Put(key string, val interface{}) (result int)         // 存入 kv
	PutIfAbsent(key string, val interface{}) (result int) // 如果不存在才存入 kv
	PutIfExists(key string, val interface{}) (result int) // 如果存在才存入 kv
	Remove(key string) (result int)                       // 删除
	ForEach(consumer Consumer)                            // 遍历整个字典
	Keys() []string                                       // 列出所有键
	RandomKeys(limit int) []string                        // 列出指定个数的键
	RandomDistinctKeys(limit int) []string                // 返回多个不重复的键
	Clear()                                               // 清空字典
}
