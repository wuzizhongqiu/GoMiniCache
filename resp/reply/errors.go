package reply

/*
 * 后续与客户端通信使用的，异常回复的常量信息
 */

/* ---- 分割线 ---- */

// UnknownErrReply 未知错误（汗流浃背了，老弟）
type UnknownErrReply struct{}

var unknownErrBytes = []byte("-Err unknown\r\n")

// ToBytes 序列化 resp.Reply
func (r *UnknownErrReply) ToBytes() []byte {
	return unknownErrBytes
}

func (r *UnknownErrReply) Error() string {
	return "Err unknown"
}

/* ---- 分割线 ---- */

// ArgNumErrReply 命令参数数目出错
type ArgNumErrReply struct {
	Cmd string
}

// ToBytes 序列化 resp.Reply
func (r *ArgNumErrReply) ToBytes() []byte {
	return []byte("-ERR wrong number of arguments for '" + r.Cmd + "' command\r\n")
}

func (r *ArgNumErrReply) Error() string {
	return "ERR wrong number of arguments for '" + r.Cmd + "' command"
}

func MakeArgNumErrReply(cmd string) *ArgNumErrReply {
	return &ArgNumErrReply{
		Cmd: cmd,
	}
}

/* ---- 分割线 ---- */

// SyntaxErrReply 语法错误
type SyntaxErrReply struct{}

var syntaxErrBytes = []byte("-Err syntax error\r\n")
var theSyntaxErrReply = &SyntaxErrReply{}

func MakeSyntaxErrReply() *SyntaxErrReply {
	return theSyntaxErrReply
}

// ToBytes 序列化 resp.Reply
func (r *SyntaxErrReply) ToBytes() []byte {
	return syntaxErrBytes
}

func (r *SyntaxErrReply) Error() string {
	return "Err syntax error"
}

/* ---- 分割线 ---- */

// WrongTypeErrReply 类型错误，操作值的类型出现错误
type WrongTypeErrReply struct{}

var wrongTypeErrBytes = []byte("-WRONG-TYPE Operation against a key holding the wrong kind of value\r\n")

// ToBytes 序列化 resp.Reply
func (r *WrongTypeErrReply) ToBytes() []byte {
	return wrongTypeErrBytes
}

func (r *WrongTypeErrReply) Error() string {
	return "WRONG-TYPE Operation against a key holding the wrong kind of value"
}

/* ---- 分割线 ---- */

// ProtocolErrReply 序列化错误，用户发来的指令不符合 RESP 协议的要求
type ProtocolErrReply struct {
	Msg string
}

// ToBytes 序列化 resp.Reply
func (r *ProtocolErrReply) ToBytes() []byte {
	return []byte("-ERR Protocol error: '" + r.Msg + "'\r\n")
}

func (r *ProtocolErrReply) Error() string {
	return "ERR Protocol error: '" + r.Msg
}
