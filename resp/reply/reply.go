package reply

/*
 * 实现一些动态的回复信息
 */

import (
	"GoMiniCache/interface/resp"
	"bytes"
	"strconv"
)

var (
	nullBulkReplyBytes = []byte("$-1")

	// CRLF 常用的序列化分隔符
	CRLF = "\r\n"
)

/* ---- 拼装一个普通的字符串进行回复 ---- */

// BulkReply 存储二进制安全的字符串（也就是用 []byte 来存）
type BulkReply struct {
	Arg []byte
}

// MakeBulkReply 创建 BulkReply
// 比如说: 从数据库中查询一个信息，就是这样拼装返回的
func MakeBulkReply(arg []byte) *BulkReply {
	return &BulkReply{
		Arg: arg,
	}
}

// ToBytes 序列化 resp.Reply
func (r *BulkReply) ToBytes() []byte {
	if len(r.Arg) == 0 {
		return nullBulkReplyBytes
	}
	// 序列化成 RESP 协议的形式
	return []byte("$" + strconv.Itoa(len(r.Arg)) + CRLF + string(r.Arg) + CRLF)
}

/* ---- 回复多个字符串切片 ---- */

// MultiBulkReply 存储字符串切片
type MultiBulkReply struct {
	Args [][]byte
}

// MakeMultiBulkReply 创建 MultiBulkReply
func MakeMultiBulkReply(args [][]byte) *MultiBulkReply {
	return &MultiBulkReply{
		Args: args,
	}
}

// ToBytes 序列化 resp.Reply
func (r *MultiBulkReply) ToBytes() []byte {
	argLen := len(r.Args)
	var buf bytes.Buffer // bytes.Buffer 包的字符串拼装，效率略高
	buf.WriteString("*" + strconv.Itoa(argLen) + CRLF)
	for _, arg := range r.Args {
		if arg == nil {
			buf.WriteString("$-1" + CRLF)
		} else {
			buf.WriteString("$" + strconv.Itoa(len(arg)) + CRLF + string(arg) + CRLF)
		}
	}
	return buf.Bytes()
}

/* ---- 回复状态信息 ---- */

// StatusReply 存储状态字符串
type StatusReply struct {
	Status string
}

// MakeStatusReply 创建 StatusReply
func MakeStatusReply(status string) *StatusReply {
	return &StatusReply{
		Status: status,
	}
}

// ToBytes 序列化 resp.Reply
func (r *StatusReply) ToBytes() []byte {
	return []byte("+" + r.Status + CRLF)
}

/* ---- 回复数字信息 ---- */

// IntReply 存储 int64 的数字
type IntReply struct {
	Code int64
}

func MakeIntReply(code int64) *IntReply {
	return &IntReply{
		Code: code,
	}
}

// ToBytes 序列化 resp.Reply
func (r *IntReply) ToBytes() []byte {
	return []byte(":" + strconv.FormatInt(r.Code, 10) + CRLF)
}

/* ---- 回复错误信息 ---- */

type ErrorReply interface {
	Error() string
	ToBytes() []byte
}

// StandardErrReply 处理程序 handle 出现错误
type StandardErrReply struct {
	Status string
}

// ToBytes 序列化 resp.Reply
func (r *StandardErrReply) ToBytes() []byte {
	return []byte("-" + r.Status + CRLF)
}

func (r *StandardErrReply) Error() string {
	return r.Status
}

// MakeErrReply 创建 StandardErrReply
func MakeErrReply(status string) *StandardErrReply {
	return &StandardErrReply{
		Status: status,
	}
}

// IsErrorReply 如果 resp.Reply 是错误的，返回 true
// 其实就是判断回复的内容第一个字符是不是 '-'
func IsErrorReply(reply resp.Reply) bool {
	return reply.ToBytes()[0] == '-'
}
