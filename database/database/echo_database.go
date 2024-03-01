package database

/*
 * 用来测试的 echo 数据库内核实现
 */

import (
	"GoMiniCache/interface/resp"
	"GoMiniCache/lib/logger"
	"GoMiniCache/resp/reply"
)

type EchoDatabase struct {
}

func NewEchoDatabase() *EchoDatabase {
	return &EchoDatabase{}
}

// Exec 执行的就是直接 echo 回复
func (e EchoDatabase) Exec(client resp.Connection, args [][]byte) resp.Reply {
	return reply.MakeMultiBulkReply(args)
}

func (e EchoDatabase) AfterClientClose(c resp.Connection) {
	logger.Info("EchoDatabase AfterClientClose")
}

func (e EchoDatabase) Close() {
	logger.Info("EchoDatabase Close")
}
