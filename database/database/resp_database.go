package database

/*
 * 实现我们使用的数据库内核
 */

import (
	"GoMiniCache/aof"
	"GoMiniCache/config"
	"GoMiniCache/database/structure"
	"GoMiniCache/interface/resp"
	"GoMiniCache/lib/logger"
	"GoMiniCache/resp/reply"
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
)

// Database 数据库集合
type Database struct {
	dbSet      []*structure.DB
	aofHandler *aof.HandlerAof
}

// NewDatabase 创建一个类 Redis 数据库
func NewDatabase() *Database {
	mdb := &Database{}
	if config.Properties.Databases == 0 {
		config.Properties.Databases = 16 // 默认 16 个
	}
	mdb.dbSet = make([]*structure.DB, config.Properties.Databases)
	for i := range mdb.dbSet {
		singleDB := structure.MakeDB() // 创建好底层存储
		singleDB.Index = i
		mdb.dbSet[i] = singleDB
	}
	if config.Properties.AppendOnly {
		aofHandler, err := aof.NewAOFHandler(mdb) // 启用 AOF 持久化
		if err != nil {
			panic(err)
		}
		mdb.aofHandler = aofHandler
		for _, db := range mdb.dbSet {
			singleDB := db
			singleDB.AddAof = func(line [][]byte) {
				mdb.aofHandler.AddAof(singleDB.Index, line)
			}
		}
	}
	return mdb
}

// Exec 执行命令
// 参数 cmdLine 存储的就是命令的内容
func (mdb *Database) Exec(c resp.Connection, cmdLine [][]byte) (result resp.Reply) {
	defer func() { // 推荐在核心流程使用 recover()
		if err := recover(); err != nil {
			logger.Warn(fmt.Sprintf("error occurs: %v\n%s", err, string(debug.Stack())))
		}
	}()

	cmdName := strings.ToLower(string(cmdLine[0]))
	if cmdName == "select" { // 选择数据库
		if len(cmdLine) != 2 {
			return reply.MakeArgNumErrReply("select")
		}
		return execSelect(c, mdb, cmdLine[1:])
	}

	selectedDB := mdb.dbSet[c.GetDBIndex()]
	return selectedDB.Exec(cmdLine) // 执行命令
}

// Close 关闭数据库（暂无逻辑）
func (mdb *Database) Close() {

}

// AfterClientClose 关闭客户端之后的操作
func (mdb *Database) AfterClientClose(resp.Connection) {

}

// execSelect 选择数据库
func execSelect(c resp.Connection, mdb *Database, args [][]byte) resp.Reply {
	dbIndex, err := strconv.Atoi(string(args[0]))
	if err != nil {
		return reply.MakeErrReply("ERR invalid DB index")
	}
	if dbIndex >= len(mdb.dbSet) {
		return reply.MakeErrReply("ERR DB index is out of range")
	}
	c.SelectDB(dbIndex)
	return reply.MakeOkReply()
}
