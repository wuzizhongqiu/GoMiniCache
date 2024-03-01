package structure

/*
 * 实现基本的 KV 存储
 */

import (
	"GoMiniCache/interface/resp"
	"GoMiniCache/lib/utils"
	"GoMiniCache/lib/wildcard"
	"GoMiniCache/resp/reply"
)

// execDel 删除键值，例：K1 K2 K3
func execDel(db *DB, args [][]byte) resp.Reply {
	keys := make([]string, len(args))
	for i, v := range args {
		keys[i] = string(v)
	}

	deleted := db.Removes(keys...)
	if deleted > 0 {
		db.AddAof(utils.ToCmdLine2("del", args...))
	}
	return reply.MakeIntReply(int64(deleted)) // 回复有多少个操作
}

// execExists 查看键是否存在
func execExists(db *DB, args [][]byte) resp.Reply {
	result := int64(0)
	for _, arg := range args {
		key := string(arg)
		_, exists := db.GetEntity(key)
		if exists {
			result++
		}
	}
	return reply.MakeIntReply(result)
}

// execFlushDB 删除所有键值
func execFlushDB(db *DB, args [][]byte) resp.Reply {
	db.Flush()
	db.AddAof(utils.ToCmdLine2("flushdb", args...))
	return reply.MakeOkReply()
}

// execType 返回实体的类型，包括: string, list, hash, set and zset
func execType(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exists := db.GetEntity(key)
	if exists == false {
		return reply.MakeStatusReply("none")
	}
	switch entity.Data.(type) { // 目前只实现了 string
	case []byte: // string 存的是字节的切片
		return reply.MakeStatusReply("string")
	}
	// TODO: 其他的数据结构的实现 case
	return reply.MakeUnknownErrReply()
}

// execRename 给 key 改名称（底层是删除原键值，插入新键值）
func execRename(db *DB, args [][]byte) resp.Reply {
	if len(args) != 2 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'rename' command")
	}
	src := string(args[0])
	dest := string(args[1]) // 目标（需要改的名称）

	entity, ok := db.GetEntity(src)
	if !ok {
		return reply.MakeErrReply("no such key")
	}
	db.PutEntity(dest, entity)
	db.Remove(src)
	db.AddAof(utils.ToCmdLine2("rename", args...))
	return reply.MakeOkReply()
}

// execRenameNx 只有新名称不存在才改变新名词
func execRenameNx(db *DB, args [][]byte) resp.Reply {
	src := string(args[0])
	dest := string(args[1])

	_, ok := db.GetEntity(dest)
	if ok == true {
		return reply.MakeIntReply(0)
	}

	entity, ok := db.GetEntity(src)
	if ok == false {
		return reply.MakeErrReply("no such key")
	}
	db.Remove(src)
	db.PutEntity(dest, entity)
	db.AddAof(utils.ToCmdLine2("renamenx", args...))
	return reply.MakeIntReply(1)
}

// execKeys 返回所有的 key
func execKeys(db *DB, args [][]byte) resp.Reply {
	pattern := wildcard.CompilePattern(string(args[0])) // 将通配符取出转换成 pattern
	result := make([][]byte, 0)
	db.Data.ForEach(func(key string, val interface{}) bool {
		if pattern.IsMatch(key) { // 判断该字符串是否匹配
			result = append(result, []byte(key))
		}
		return true
	})
	return reply.MakeMultiBulkReply(result)
}

func init() {
	RegisterCommand("Del", execDel, -2)          // 删除键值的参数数量需要 >=2
	RegisterCommand("Exists", execExists, -2)    // 判断是否存在的参数需要 >=2
	RegisterCommand("Keys", execKeys, 2)         // 判断键是否存在参数需要 ==2
	RegisterCommand("FlushDB", execFlushDB, -1)  // 清空字典参数需要 >=1
	RegisterCommand("Type", execType, 2)         // 判断键值类型参数需要 ==2
	RegisterCommand("Rename", execRename, 3)     // 修改键名的参数需要 ==3
	RegisterCommand("RenameNx", execRenameNx, 3) // 修改键名的参数需要 ==3
}
