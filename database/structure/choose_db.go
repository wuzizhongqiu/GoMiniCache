package structure

/*
 * 兼容 Redis 的数据库的接口，选择底层存储
 */

import (
	"GoMiniCache/datastruct/dict"
	"GoMiniCache/interface/database"
	"GoMiniCache/interface/resp"
	"GoMiniCache/resp/reply"
	"strings"
)

// DB 存储数据并执行用户命令
type DB struct {
	Index  int       // 使用哪个数据库
	Data   dict.Dict // 我们的底层可以在这里换实现
	AddAof func([][]byte)
}

// ExecFunc 执行函数的实现
type ExecFunc func(db *DB, args [][]byte) resp.Reply

// MakeDB 创建 DB 实例
func MakeDB() *DB {
	db := &DB{
		Data: dict.MakeSyncDict(), // 底层存储（可改）
	}
	return db
}

// Exec 执行命令（使用我们实现好的命令执行方法）
func (db *DB) Exec(cmdLine [][]byte) resp.Reply {
	cmdName := strings.ToLower(string(cmdLine[0])) // 统一执行小写的命令
	cmd, ok := CmdTable[cmdName]                   // 这个 map 是只读的，没有并发安全问题
	if ok == false {
		return reply.MakeErrReply("ERR unknown command '" + cmdName + "'")
	}
	if validateArity(cmd.Arity, cmdLine) == false { // 校验参数个数是否合法
		return reply.MakeArgNumErrReply(cmdName)
	}
	fun := cmd.Executor
	// SET K V （这里 set 就不需要了）
	return fun(db, cmdLine[1:])
}

// validateArity 校验参数个数是否正确
// 我们规定如果参数是固定的，就正常校验，如果参数是变长的，就设定为负数（校验的实收），举个例子
// SET K V -> arity == 3
// EXISTS K1 K2 ... arity == -2
func validateArity(arity int, cmdArgs [][]byte) bool {
	argNum := len(cmdArgs)
	if arity >= 0 {
		return argNum == arity
	}
	return argNum >= -arity
}

/* ---- 解析命令的实现 ----- */

// GetEntity 返回绑定的 DataEntity
func (db *DB) GetEntity(key string) (*database.DataEntity, bool) {
	raw, ok := db.Data.Get(key)
	if !ok {
		return nil, false
	}
	entity, _ := raw.(*database.DataEntity)
	return entity, true
}

// PutEntity 调用存入
func (db *DB) PutEntity(key string, entity *database.DataEntity) int {
	return db.Data.Put(key, entity)
}

// PutIfExists 调用存入
func (db *DB) PutIfExists(key string, entity *database.DataEntity) int {
	return db.Data.PutIfExists(key, entity)
}

// PutIfAbsent 调用存入
func (db *DB) PutIfAbsent(key string, entity *database.DataEntity) int {
	return db.Data.PutIfAbsent(key, entity)
}

// Remove 调用删除
func (db *DB) Remove(key string) {
	db.Data.Remove(key)
}

// Removes 删除多个键值
func (db *DB) Removes(keys ...string) (deleted int) {
	deleted = 0
	for _, key := range keys {
		_, exists := db.Data.Get(key)
		if exists == true {
			db.Remove(key)
			deleted++
		}
	}
	return deleted
}

// Flush 清空字典
func (db *DB) Flush() {
	db.Data.Clear()
}
