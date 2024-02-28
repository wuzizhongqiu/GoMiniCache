package database

import (
	"GoMiniCache/interface/resp"
)

// CmdLine 代表一个命令行
type CmdLine = [][]byte

// Database Redis 风格的存储引擎
type Database interface {
	Exec(client resp.Connection, args [][]byte) resp.Reply // 执行指令
	AfterClientClose(c resp.Connection)                    // 关闭后的操作（善后工作）
	Close()                                                // 关闭连接
}

// DataEntity 指代 Redis 的数据结构，包括 string, list, hash, set 等等
type DataEntity struct {
	Data interface{}
}
