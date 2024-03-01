package structure

/*
 * 实现 PING 操作回复 PONG
 */

import (
	"GoMiniCache/interface/resp"
	"GoMiniCache/resp/reply"
)

// Ping 回复 Pong
func Ping(db *DB, args [][]byte) resp.Reply {
	return reply.MakePongReply()
}

func init() {
	RegisterCommand("ping", Ping, -1) // PING 需要参数 >=1
}
