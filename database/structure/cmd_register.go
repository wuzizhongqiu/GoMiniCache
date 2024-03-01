package structure

/*
 * 对指令的不同执行方法进行抽象封装
 */

import (
	"strings"
)

// CmdTable 给每个指令对应一个 command 结构体
var CmdTable = make(map[string]*command)

type command struct {
	Executor ExecFunc // 这个命令的执行方法
	Arity    int      // 这个命令的参数数量
}

// RegisterCommand 注册一个新命令（这样每个指令就能有他自己的实现了）
// name 是命令的名称，executor 是执行的方法，arity 是命令的参数数量
func RegisterCommand(name string, executor ExecFunc, arity int) {
	name = strings.ToLower(name)
	CmdTable[name] = &command{
		Executor: executor,
		Arity:    arity,
	}
}
