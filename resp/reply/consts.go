package reply

/*
 * 后续与客户端通信使用的，正常回复的常量信息
 */

/* ---- 分割线 ---- */

// PongReply 会 +PONG
// 小知识: 给 Redis 发一个 PING 他会回一个 PONG
type PongReply struct{}

var thePongReply = new(PongReply)

var pongBytes = []byte("+PONG\r\n")

// ToBytes 序列化 resp.Reply
func (r *PongReply) ToBytes() []byte {
	return pongBytes
}

func MakePongReply() *PongReply {
	return thePongReply
}

/* ---- 分割线 ---- */

// OkReply 会 +OK
type OkReply struct{}

var okBytes = []byte("+OK\r\n")

// ToBytes 序列化 resp.Reply
func (r *OkReply) ToBytes() []byte {
	return okBytes
}

var theOkReply = new(OkReply)

// MakeOkReply 返回一个 OkReply
func MakeOkReply() *OkReply {
	return theOkReply
}

/* ---- 分割线 ---- */

// nullBulkBytes $-1 表示空回复
var nullBulkBytes = []byte("$-1\r\n")

// NullBulkReply 为空字符串
type NullBulkReply struct{}

// ToBytes 序列化 resp.Reply
func (r *NullBulkReply) ToBytes() []byte {
	return nullBulkBytes
}

// MakeNullBulkReply creates a new NullBulkReply
func MakeNullBulkReply() *NullBulkReply {
	return &NullBulkReply{}
}

/* ---- 分割线 ---- */

// emptyMultiBulkBytes 用 *0 表示
var emptyMultiBulkBytes = []byte("*0\r\n")

// EmptyMultiBulkReply 一个空列表
type EmptyMultiBulkReply struct{}

// ToBytes 序列化 resp.Reply
func (r *EmptyMultiBulkReply) ToBytes() []byte {
	return emptyMultiBulkBytes
}

func MakeEmptyMultiBulkReply() *EmptyMultiBulkReply {
	return &EmptyMultiBulkReply{}
}

/* ---- 分割线 ---- */

// noBytes 真的是空，什么都没有
var noBytes = []byte("")

// NoReply 什么也不响应
type NoReply struct{}

// ToBytes 序列化 resp.Reply
func (r *NoReply) ToBytes() []byte {
	return noBytes
}
